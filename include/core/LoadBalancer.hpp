#pragma once

#include <mutex>
#include <shared_mutex>
#include <thread>
#include <atomic>
#include <memory>
#include <boost/asio/ip/tcp.hpp>
#include "../config/Config.hpp"
#include "../strategy/IStrategy.hpp"
#include "../core/SessionsWorker.hpp"

// There are exactly as many sessions workers as logic threads on the machine
const unsigned int NO_OF_CORES = std::thread::hardware_concurrency();

enum State
{
    IDLE,
    ACTIVE
};

class LoadBalancer
{
public:
    // Forbid copying LoadBalancer instance
    LoadBalancer(const LoadBalancer &) = delete;
    LoadBalancer &operator=(const LoadBalancer &) = delete;

    /**
     * @brief Set instance of LoadBalancer with provided config
     * @param config
     * @throws std::runtime_exception When instance already exists
     * @return Pointer to LoadBalancer instance
     */
    static LoadBalancer *SetInstance(Config config);

    // Get instance
    static LoadBalancer *GetInstance();

    /**
     * @breief Create and add a new server to LoadBalancer
     * @throws std::runtime_exception When server creation was not successful
     */
    void AttachServer(ServerConfig serverConfig);

    // Remove server from LoadBalancer suddenly closing all of opened connections with it
    void DettachServer(ServerConfig serverConfig);

    // Return list of servers configs objects
    std::vector<ServerConfig> GetServers();

    /**
     * @brief Signal server on connection close for strategy update
     * @param server Server to signal
     */
    void Signal(std::shared_ptr<ServerConfig> server);

    /**
     * @brief Start balancing load When balancer has already started
     * @throws std::runtime_exception
     */
    void StartWork();

    // Stop balancing load
    void StopWork();

    /**
     * @brief Set instance to nullptr and deletes servers_
     */
    ~LoadBalancer();

private:
    // Mutex for synchronization of singleton creation
    static std::mutex *singletonMutex_;
    // Singleton instance
    static LoadBalancer *instance;

    // Port
    uint16_t clientsPort_;

    // Mutex for managing state
    std::mutex stateMutex_;
    // State
    State state_;

    // Mutex for operations on servers in strategy, name comes from servers_ field that requires syncing
    std::shared_mutex serversMutex_;

    // Strategy of balancing load
    IStrategy *strategy_;

    // List of sessions workers instances
    std::vector<SessionsWorker *> workers_;

    // List of sessions workers thread
    std::vector<std::thread *> workersThreads_;

    // Round robin index for sessions workers
    std::atomic<uint32_t> currentWorkerIndex_{0};

    // Context for accepting new clients connections
    std::shared_ptr<boost::asio::io_context> balancerContext_;

    // Acceptor for new tcp connections
    std::unique_ptr<boost::asio::ip::tcp::acceptor> acceptor_;

    // Thread for running balancer context
    std::thread *balancerThread_;

    /**
     * @brief Sets fields from config, creates sockets one for each server and one clients
     * @throws std::invalid_argument
     * @param config
     */
    LoadBalancer(Config config);

    /**
     * @brief Get next sessions worker using round robin
     * @return Pointer to the next SessionsWorker
     */
    SessionsWorker *getNextWorker();

    /**
     * @brief Start accepting new client connections
     */
    void startAccepting();
};
