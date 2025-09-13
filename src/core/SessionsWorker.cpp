#include "core/SessionsWorker.hpp"
#include "core/LoadBalancer.hpp"
#include "config/Config.hpp"
#include <boost/asio/write.hpp>

SessionsWorker::SessionsWorker(LoadBalancer *loadBalancer)
    : loadBalancer_(loadBalancer)
{
    workerContext = std::make_shared<boost::asio::io_context>();
}

void SessionsWorker::StartWork()
{
    // Create work without which context would end work without any tasks
    auto work = boost::asio::make_work_guard(*workerContext);
    workerContext->run();
}

void SessionsWorker::NewSession(skt client, skt server, std::shared_ptr<ServerConfig> serverConfig)
{
    // Convert to shared_ptr for safe lifetime management
    auto clientPtr = std::make_shared<skt>(std::move(client));
    auto serverPtr = std::make_shared<skt>(std::move(server));

    // Insert new pair to bimap and track server config
    {
        std::unique_lock<std::shared_mutex> lock(socketsBimapMutex_);
        socketsBimap_.insert({clientPtr, serverPtr});
        clientToServer_[clientPtr] = serverConfig;
    }

    // Start relay in both directions
    startRelay(clientPtr, serverPtr); // client -> server
    startRelay(serverPtr, clientPtr); // server -> client
}

void SessionsWorker::startRelay(sktPtr from, sktPtr to)
{
    // Create buffer for this relay session
    auto buffer = std::make_shared<std::array<uint8_t, 4096>>();

    from->async_read_some(
        boost::asio::buffer(*buffer),
        [this, from, to, buffer](const boost::system::error_code &ec, std::size_t bytes)
        {
            if (!ec && bytes > 0)
            {
                // Forward data to destination
                boost::asio::async_write(
                    *to,
                    boost::asio::buffer(*buffer, bytes),
                    [this, from, to](const boost::system::error_code &writeEc, std::size_t)
                    {
                        if (!writeEc)
                        {
                            // Continue reading from source
                            startRelay(from, to);
                        }
                        else
                        {
                            // Write failed, close connection
                            closeConnection(from, to);
                        }
                    });
            }
            else
            {
                // Read failed or connection closed
                closeConnection(from, to);
            }
        });
}

void SessionsWorker::closeConnection(sktPtr socket1, sktPtr socket2)
{
    std::shared_ptr<ServerConfig> serverConfig;

    // Remove from bimap and get server config
    {
        std::unique_lock<std::shared_mutex> lock(socketsBimapMutex_);

        // Get server config from client socket
        auto serverIt = clientToServer_.find(socket1);
        if (serverIt != clientToServer_.end())
        {
            serverConfig = serverIt->second;
            clientToServer_.erase(serverIt);
        }
        else
        {
            // Maybe socket1 is server and socket2 is client
            auto serverIt2 = clientToServer_.find(socket2);
            if (serverIt2 != clientToServer_.end())
            {
                serverConfig = serverIt2->second;
                clientToServer_.erase(serverIt2);
            }
        }

        // Remove from bimap
        auto it = socketsBimap_.left.find(socket1);
        if (it != socketsBimap_.left.end())
        {
            socketsBimap_.left.erase(it);
        }
        else
        {
            auto it_right = socketsBimap_.right.find(socket1);
            if (it_right != socketsBimap_.right.end())
            {
                socketsBimap_.right.erase(it_right);
            }
        }
    }

    // Close both sockets
    if (socket1->is_open())
    {
        socket1->close();
    }
    if (socket2->is_open())
    {
        socket2->close();
    }

    // Signal load balancer to signal strategy
    if (serverConfig && loadBalancer_)
    {
        loadBalancer_->Signal(serverConfig);
    }
}
