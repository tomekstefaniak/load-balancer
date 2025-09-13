#include <stdexcept>
#include <algorithm>
#include <memory>
#include "strategy/LeastConnections.hpp"
#include "config/Config.hpp"

LeastConnections::LeastConnections(const std::vector<ServerConfig> &serversConfigs)
    : servers_()
{
    // Create attached server for every server config
    for (const auto &serverConfig : serversConfigs)
    {
        auto server = std::make_shared<ServerConfig>(serverConfig);
        servers_.emplace(
            std::piecewise_construct,
            std::forward_as_tuple(server),
            std::forward_as_tuple(0));
    }
}

std::shared_ptr<ServerConfig> LeastConnections::Next()
{
    // Check if there are any attached servers
    if (servers_.empty())
    {
        throw std::runtime_error("LeastConnections: no servers");
    }

    std::shared_ptr<ServerConfig> result;
    uint32_t min = ~0;
    for (const auto &server : servers_)
    {
        if (server.second.load() < min)
        {
            result = server.first;
        }
    }

    servers_[result].fetch_add(1); // Increment connections
    return result;
}

void LeastConnections::AttachServer(ServerConfig serverConfig)
{
    // Check if server is already attached
    for (const auto &server : servers_)
    {
        if (*(server.first) == serverConfig)
        {
            return;
        }
    }

    // Create attached server in a shared pointer
    auto server = std::make_shared<ServerConfig>(serverConfig);

    // Attach server to servers_
    servers_.emplace(
        std::piecewise_construct,
        std::forward_as_tuple(server),
        std::forward_as_tuple(0));
}

void LeastConnections::DettachServer(ServerConfig serverConfig)
{
    // If server is attached, then detach it
    for (const auto &server : servers_)
    {
        if (*(server.first) == serverConfig)
        {
            servers_.erase(server.first);
        }
    }
}

std::vector<ServerConfig> LeastConnections::GetServers()
{
    std::vector<ServerConfig> serversConfigs;
    for (const auto &server : servers_)
    {
        serversConfigs.push_back(*(server.first));
    }
    return serversConfigs;
}

void LeastConnections::Signal(std::shared_ptr<ServerConfig> server)
{
    auto it = servers_.find(server);
    // If server is attached
    if (it != servers_.end())
    {
        servers_[server]--; // Decrement connections
    }
}

void LeastConnections::Reset()
{
    // Reset all connection counters to 0
    for (auto &server : servers_)
    {
        server.second.store(0);
    }
}

LeastConnections::~LeastConnections() = default;
