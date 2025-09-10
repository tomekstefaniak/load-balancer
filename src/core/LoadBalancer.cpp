#include <stdexcept>
#include "core/LoadBalancer.hpp"
#include "strategy/LeastConnections.hpp"
#include "strategy/RoundRobin.hpp"

LoadBalancer *LoadBalancer::instance = nullptr;
std::mutex *LoadBalancer::singletonMutex_ = new std::mutex;

LoadBalancer::LoadBalancer(Config config)
    : strategy_(), stateMutex_(new std::mutex), state_(IDLE), serversMutex_(new std::shared_mutex())
{
    // Create strategy
    if (config.algorithmName == "leastconnections") {
        strategy_ = new LeastConnections(config.serversConfigs);
    } else if (config.algorithmName == "roundrobin") {
        strategy_ = new RoundRobin(config.serversConfigs);
    } else {
        throw std::invalid_argument("uknown strategy");
    }
}

LoadBalancer* LoadBalancer::SetInstance(Config config)
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
    std::unique_lock<std::shared_mutex> lock(*serversMutex_);
    strategy_->AttachServer(serverConfig);
}

void LoadBalancer::DettachServer(ServerConfig serverConfig)
{
    // Lock & enter safely
    std::unique_lock<std::shared_mutex> lock(*serversMutex_);
    strategy_->DettachServer(serverConfig);
}

std::vector<ServerConfig> LoadBalancer::GetServers()
{
    // Lock & enter safely
    std::shared_lock<std::shared_mutex> lock(*serversMutex_);
    return strategy_->GetServers();
}

void LoadBalancer::StartWork()
{

}

void LoadBalancer::StopWork()
{

}

LoadBalancer::~LoadBalancer()
{

}
