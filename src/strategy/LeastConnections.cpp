#include <stdexcept>
#include <algorithm>
#include <memory>
#include "strategy/LeastConnections.hpp"
#include "config/Config.hpp"

LeastConnections::LeastConnections(const std::vector<ServerConfig> &serversConfigs)
    : servers_()
{
    // Create attached server for every server config
    for (const auto &serverConfig : serversConfigs) {
        auto server = std::make_shared<AttachedServer>(
            serverConfig,
            true
        );
        servers_.emplace(
            std::piecewise_construct,
            std::forward_as_tuple(server),
            std::forward_as_tuple(0)
        );
    }
}

std::shared_ptr<AttachedServer> LeastConnections::Next()
{
    // Check if there are any attached servers
    if (servers_.empty()) {
        throw std::runtime_error("LeastConnections: no servers");
    }

    std::shared_ptr<AttachedServer> result;
    uint32_t min = ~0;
    for (const auto& server : servers_) {
        if (server.second.load() < min) {
            result = server.first;
        }
    }

    servers_[result].fetch_add(1); // Increment connections
    return result;
}

void LeastConnections::AttachServer(ServerConfig serverConfig)
{
    // Check if server is already attached
    for (const auto &server : servers_) {
        if (server.first->serverConfig == serverConfig) {
            return;
        }
    }

    // Create attached server in a shared pointer
    auto server = std::make_shared<AttachedServer>(
        serverConfig,
        true
    );

    // Attach server to servers_
    servers_.emplace(
        std::piecewise_construct,
        std::forward_as_tuple(server),
        std::forward_as_tuple(0)
    );
}

void LeastConnections::DettachServer(ServerConfig serverConfig)
{
    // If server is attached, then detach it
    for (const auto &server : servers_) {
        if (server.first->serverConfig == serverConfig) {
            servers_.erase(server.first);
            server.first->attached.store(false);
        }
    }
}

std::vector<ServerConfig> LeastConnections::GetServers()
{
    std::vector<ServerConfig> serversConfigs;
    for (const auto &server : servers_) {
        serversConfigs.push_back(server.first->serverConfig);
    }
    return serversConfigs;
}

void LeastConnections::Signal(std::shared_ptr<AttachedServer> server) {
    auto it = servers_.find(server);
    // If server is attached
    if (it != servers_.end()) {
        servers_[server]--; // Decrement connections
    }
}

LeastConnections::~LeastConnections() = default;
