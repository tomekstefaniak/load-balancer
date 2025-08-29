#include "core/LoadBalancer.hpp"

LoadBalancer *LoadBalancer::instance = nullptr;
std::mutex *LoadBalancer::mutex_;

LoadBalancer::LoadBalancer() {}

LoadBalancer *LoadBalancer::GetInstance()
{
    // Quick check in case instance already exists
    if (instance == nullptr)
    {
        // Safe check for safety in case of concurrency
        std::lock_guard<std::mutex> lock(*mutex_);
        if (instance == nullptr)
        {
            instance = new LoadBalancer();
        }
    }
    return instance;
}

void LoadBalancer::StartWork()
{
}

void LoadBalancer::EndWork()
{
}
