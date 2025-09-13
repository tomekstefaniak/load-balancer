#include <stdexcept>
#include "strategy/RoundRobin.hpp"
#include "config/Config.hpp"
#include <iostream>

RoundRobin::RoundRobin(const std::vector<ServerConfig> &serversConfigs)
    : curr_(0)
{
    // Create attached server for every server config
    for (const auto &serverConfig : serversConfigs)
    {
        auto server = std::make_shared<ServerConfig>(serverConfig);
        servers_.push_back(server);
    }
}

std::shared_ptr<ServerConfig> RoundRobin::Next()
{
    // Check is there are any attached servers
    if (servers_.empty())
    {
        throw std::runtime_error("RoundRobin: no servers");
    }

    auto result = servers_[curr_];
    std::cout << "RoundRobin::Next() : curr_=" << curr_ << ", server port=" << result->serverPort << std::endl; // debug
    curr_ = (curr_ + 1 >= servers_.size()) ? 0 : curr_ + 1; // Update curr
    return result;
}

void RoundRobin::AttachServer(ServerConfig serverConfig)
{
    // Check if server is already attached
    for (auto &server : servers_)
    {
        if (*(server) == serverConfig)
        {
            return;
        }
    }

    // Create and push back new attached server
    auto server = std::make_shared<ServerConfig>(serverConfig);
    servers_.push_back(server);
}

void RoundRobin::DettachServer(ServerConfig serverConfig)
{
    // If server is attached, then detach it
    for (int i = 0; i < servers_.size(); i++)
    {
        if (*(servers_[i]) == serverConfig)
        {
            servers_.erase(servers_.begin() + i);
            if (i < curr_)
            {
                curr_--;
            }
            if (curr_ >= servers_.size())
            {
                curr_ = 0;
            }
        }
    }
}

std::vector<ServerConfig> RoundRobin::GetServers()
{
    std::vector<ServerConfig> serversConfigs;
    for (const auto &server : servers_)
    {
        serversConfigs.push_back(*server);
    }
    return serversConfigs;
}

void RoundRobin::Signal(std::shared_ptr<ServerConfig> server) {}

void RoundRobin::Reset()
{
    curr_ = 0;
}

RoundRobin::~RoundRobin() = default;
