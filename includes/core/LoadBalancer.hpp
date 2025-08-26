#pragma once

#include <mutex>

class LoadBalancer {
    public:
        LoadBalancer(const LoadBalancer&) = delete;
        LoadBalancer& operator=(const LoadBalancer&) = delete;

        static LoadBalancer* GetInstance();

        // Create and add new server to LoadBalancer
        bool SetupServer();

        // Remove server from LoadBalancer without sudden closing of opened connections with it
        void RemoveServerSafely();
        // Remove server from LoadBalancer suddenly closing all of opened connections with it
        void RemoveServerBrutally();

        // Change load balancing algorithm
        void ChangeAlgo();

        void Start();
        void End();

    private:
        static LoadBalancer* instance;
        static std::mutex* mutex_;

        LoadBalancer();
};
