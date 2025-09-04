#include <stdexcept>
#include "core/LoadBalancer.hpp"

LoadBalancer *LoadBalancer::instance = nullptr;
std::mutex *LoadBalancer::singletonMutex_ = new std::mutex;

LoadBalancer::LoadBalancer(Config *config) {
    serversMutex_ = new std::mutex();
    stateMutex_ = new std::mutex();
    state_ = ACTIVE;

    // Set config

}

LoadBalancer* LoadBalancer::SetInstance(Config *config)
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

LoadBalancer *LoadBalancer::GetInstance() {
    return instance;
}

void LoadBalancer::SetupServer() {

}

void LoadBalancer::StartWork()
{

}

void LoadBalancer::StopWork() {

}

LoadBalancer::~LoadBalancer()
{
    // Lock in case of SetIntance() call
    std::lock_guard<std::mutex> lockSingletonMutex(*singletonMutex_);
    // Lock in case of calls of methods making operations on servers_
    std::lock_guard<std::mutex> lockOpsMutex(*serversMutex_);
    // Lock in case of methods modifying state of load balancer
    std::lock_guard<std::mutex> lockStateMutex(*stateMutex_);

    // Stop the load balancer
    this->StopWork();

    // Reset static singleton fields
    instance = nullptr;
    singletonMutex_ = new std::mutex;

    // Closing servers and deleting them
    for (const auto server : servers_) {
        // Close connections
        delete server; // Free memory
    }
}
