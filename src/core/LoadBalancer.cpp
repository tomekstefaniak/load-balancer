#include <stdexcept>
#include "core/LoadBalancer.hpp"
#include <boost/asio/ip/tcp.hpp>
#include <boost/asio/ip/address.hpp>
#include <boost/asio/io_context.hpp>
#include "strategy/LeastConnections.hpp"
#include "strategy/RoundRobin.hpp"

LoadBalancer *LoadBalancer::instance = nullptr;
std::mutex *LoadBalancer::singletonMutex_ = new std::mutex;

LoadBalancer::LoadBalancer(Config config)
    : strategy_(), state_(IDLE), balancerContext_(std::make_shared<boost::asio::io_context>()),
      acceptor_(nullptr), balancerThread_(nullptr)
{
    // Create strategy
    if (config.algorithmName == "leastconnections")
    {
        strategy_ = new LeastConnections(config.serversConfigs);
    }
    else if (config.algorithmName == "roundrobin")
    {
        strategy_ = new RoundRobin(config.serversConfigs);
    }
    else
    {
        throw std::invalid_argument("unknown strategy");
    }

    // Set clients port
    clientsPort_ = config.clientsPort;
}

LoadBalancer *LoadBalancer::SetInstance(Config config)
{
    // Quick check in case instance already exists
    if (instance == nullptr)
    {
        // Safe check for safety in case of concurrency
        std::lock_guard<std::mutex> lock(*singletonMutex_);
        if (instance == nullptr)
        {
            instance = new LoadBalancer(config);
            return instance;
        }
    }

    throw std::runtime_error("LoadBalancer instance exists");
}

LoadBalancer *LoadBalancer::GetInstance()
{
    return instance;
}

void LoadBalancer::AttachServer(ServerConfig serverConfig)
{
    // Lock & enter safely
    std::unique_lock<std::shared_mutex> lock(serversMutex_);
    strategy_->AttachServer(serverConfig);
}

void LoadBalancer::DettachServer(ServerConfig serverConfig)
{
    // Lock & enter safely
    std::unique_lock<std::shared_mutex> lock(serversMutex_);
    strategy_->DettachServer(serverConfig);
}

std::vector<ServerConfig> LoadBalancer::GetServers()
{
    // Lock & enter safely
    std::shared_lock<std::shared_mutex> lock(serversMutex_);
    return strategy_->GetServers();
}

void LoadBalancer::Signal(std::shared_ptr<ServerConfig> server)
{
    // Lock & enter safely
    std::unique_lock<std::shared_mutex> lock(serversMutex_);
    strategy_->Signal(server);
}

void LoadBalancer::StartWork()
{
    {
        // Lock state mutex because we set state to ACTIVE
        std::lock_guard<std::mutex> lock(stateMutex_);

        if (state_ != IDLE)
        {
            throw std::runtime_error("LoadBalancer instance has already started");
        }

        // Start sessions workers
        for (int i = 0; i < NO_OF_CORES; i++)
        {
            auto worker = new SessionsWorker(this);

            auto workerThread = new std::thread(
                [worker]()
                { worker->StartWork(); });

            // Push back new session worker
            workers_.push_back(worker);
            // Push back new thread
            workersThreads_.push_back(workerThread);
        }

        state_ = ACTIVE;
    }

    // Create acceptor for new tcp connections
    acceptor_ = std::make_unique<boost::asio::ip::tcp::acceptor>(
        *balancerContext_,
        boost::asio::ip::tcp::endpoint(boost::asio::ip::tcp::v4(), clientsPort_));

    // Start accepting connections
    startAccepting();

    balancerThread_ = new std::thread(
        [this]()
        { balancerContext_->run(); } // Run the io_context
    );
}

SessionsWorker *LoadBalancer::getNextWorker()
{
    if (workers_.empty())
    {
        return nullptr;
    }

    size_t index = currentWorkerIndex_.fetch_add(1) % workers_.size();
    return workers_[index];
}

void LoadBalancer::startAccepting()
{
    // Get next SessionsWorker using round robin
    if (SessionsWorker *worker = getNextWorker())
    {
        // Create new socket for client connection
        auto clientSocket = std::make_shared<boost::asio::ip::tcp::socket>(*(worker->workerContext));

        acceptor_->async_accept(
            *clientSocket,
            [this, worker, clientSocket](const boost::system::error_code &ec)
            {
                if (!ec)
                {
                    // Get next server using strategy
                    std::shared_lock<std::shared_mutex> lock(serversMutex_);
                    auto server = strategy_->Next();
                    lock.unlock();

                    if (server)
                    {
                        // Create server socket and connect to server
                        auto serverSocket = std::make_shared<boost::asio::ip::tcp::socket>(*(worker->workerContext));

                        boost::asio::ip::tcp::endpoint serverEndpoint(
                            boost::asio::ip::make_address(server->serverIP),
                            server->serverPort);

                        serverSocket->async_connect(
                            serverEndpoint,
                            [this, worker, clientSocket, serverSocket, server](const boost::system::error_code &connectEc)
                            {
                                if (!connectEc)
                                {
                                    // Pass both sockets and server config to SessionsWorker
                                    worker->NewSession(std::move(*clientSocket), std::move(*serverSocket), server);
                                }
                                else
                                {
                                    // Connection to server failed, close client socket
                                    clientSocket->close();
                                }
                            });
                    }
                    else
                    {
                        // No available server, close client connection
                        clientSocket->close();
                    }
                }

                // Continue accepting new connections
                startAccepting();
            });
    }
    else
    {
        throw std::runtime_error("no sessions workers available");
    }
}

void LoadBalancer::StopWork()
{
    // Lock state mutex because we change state
    std::lock_guard<std::mutex> lock(stateMutex_);

    if (state_ != ACTIVE)
    {
        throw std::runtime_error("LoadBalancer is not active");
    }

    // Stop balancer context
    if (balancerContext_)
    {
        balancerContext_->stop();
    }

    // Join balancer thread
    if (balancerThread_ && balancerThread_->joinable())
    {
        balancerThread_->join();
        delete balancerThread_;
    }

    // Stop all worker contexts
    for (auto *worker : workers_)
    {
        if (worker && worker->workerContext)
        {
            worker->workerContext->stop();
        }
    }

    // Join all worker threads
    for (auto *thread : workersThreads_)
    {
        if (thread && thread->joinable())
        {
            thread->join();
            delete thread;
        }
    }

    // Clean up workers
    for (auto *worker : workers_)
    {
        if (worker)
        {
            delete worker;
        }
    }

    // Clear containers
    workers_.clear();
    workersThreads_.clear();

    // Reset contexts
    balancerContext_.reset();
    balancerContext_ = std::make_shared<boost::asio::io_context>();

    // Reset strategy state
    {
        std::unique_lock<std::shared_mutex> lock(serversMutex_);
        if (strategy_)
        {
            strategy_->Reset();
        }
    }

    state_ = IDLE;
}

LoadBalancer::~LoadBalancer()
{
    // Stop balancer if it's still active
    if (state_ == ACTIVE)
    {
        try
        {
            StopWork();
        }
        catch (...)
        {
            // Ignore exceptions in destructor - just ensure cleanup
        }
    }

    // Delete strategy and its resources
    if (strategy_)
    {
        delete strategy_;
        strategy_ = nullptr;
    }

    // Clean up any remaining workers (should be empty after StopWork)
    for (auto *worker : workers_)
    {
        if (worker)
        {
            delete worker;
        }
    }
    workers_.clear();

    // Clean up any remaining threads (should be empty after StopWork)
    for (auto *thread : workersThreads_)
    {
        if (thread)
        {
            if (thread->joinable())
            {
                thread->join();
            }
            delete thread;
        }
    }
    workersThreads_.clear();

    // Reset singleton instance
    instance = nullptr;
}
