#pragma once

#include <mutex>
#include <shared_mutex>
#include "../config/Config.hpp"
#include "../strategy/IStrategy.hpp"

enum State {
    IDLE,
    BUSY
};

class LoadBalancer
{
public:
    // Forbid copying LoadBalancer instance
    LoadBalancer(const LoadBalancer&) = delete;
    LoadBalancer& operator=(const LoadBalancer&) = delete;

    /**
     * @brief Set instance of LoadBalancer with provided config
     * @param config
     * @throws std::runtime_exception When instance already exists
     * @return Pointer to LoadBalancer instance
     */
    static LoadBalancer* SetInstance(Config config);

    // Get instance
    static LoadBalancer* GetInstance();

    /**
     * @breief Create and add a new server to LoadBalancer
     * @throws std::runtime_exception When server creation was not successful
     */
    void AttachServer(ServerConfig serverConfig);

    // Remove server from LoadBalancer suddenly closing all of opened connections with it
    void DettachServer(ServerConfig serverConfig);

    // Return list of servers configs objects
    std::vector<ServerConfig> GetServers();

    // Start balancing load
    void StartWork();

    // Stop balancing load
    void StopWork();

    /**
     * @brief Set instance to nullptr and deletes servers_
     */
    ~LoadBalancer();

private:
    // Mutex for synchronization of singleton creation
    static std::mutex* singletonMutex_;
    // Signleton instance
    static LoadBalancer* instance;

    // Mutex for managing state
    std::mutex* stateMutex_;
    // State
    State state_;

    // Mutex for operations on servers in strategy, name comes from servers_ field that requires syncing
    std::shared_mutex* serversMutex_;

    // Strategy of balancing load
    IStrategy* strategy_;

    /**
     * @brief Sets fields from config, creates sockets one for each server and one clients
     * @throws std::invalid_argument
     * @param config
     */
    LoadBalancer(Config config);
};
