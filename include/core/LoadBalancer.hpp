#pragma once

#include <mutex>
#include "../config/Config.hpp"

enum State {
    DEAD,
    ACTIVE,
    IDLE
};

class LoadBalancer
{
public:
    // Forbid copying LoadBalancer instance
    LoadBalancer(const LoadBalancer&) = delete;
    LoadBalancer& operator=(const LoadBalancer&) = delete;

    /**
     * @brief Set instance of LoadBalancer with provided config
     * @throws std::runtime_exception When instance already exists
     * @param config
     * @return Pointer to LoadBalancer instance
     */
    static LoadBalancer* SetInstance(Config *config);

    // Get instance
    static LoadBalancer* GetInstance();

    /**
     * @breief Create and add a new server to LoadBalancer
     * @throws std::runtime_exception When server creation was not successful
     */
    void SetupServer();

    // Remove server from LoadBalancer suddenly closing all of opened connections with it
    void RemoveServer();

    // Return list of servers objects
    std::vector<ServerConfig*> GetServers();

    // Change load balancing algorithm
    void ChangeAlgo();

    // Start balancing load
    void StartWork();

    // Stop balancing load
    void StopWork();

    /**
     * @brief Set instance to nullptr and deletes servers_
     */
    ~LoadBalancer();

private:
    static LoadBalancer* instance;
    static std::mutex* singletonMutex_;

    // Mutex for servers operations
    std::mutex *serversMutex_;
    // Server instances vector
    std::vector<ServerConfig*> servers_;

    // Mutex for StartWork() and StopWork() methods
    std::mutex* stateMutex_;
    // State indicating the mode of load balancer
    State state_;


    /**
     * @brief Sets fields from config, creates sockets one for each server and one clients
     * @throw std::runtime_exception
     * @param config
     */
    LoadBalancer(Config *config);
};
