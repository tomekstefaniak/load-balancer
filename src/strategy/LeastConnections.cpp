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
        auto attachedServer = std::make_shared<AttachedServer>(
            serverConfig,
            true
        );
        servers_[attachedServer] = 0;
    }
}

std::shared_ptr<AttachedServer> LeastConnections::Next()
{
    // Check is there are any attached servers
    if (servers_.empty()) {
        throw std::runtime_error("LeastConnections: no servers");
    }

    std::shared_ptr<AttachedServer> result;
    uint32_t min = ~0;
    for (const auto server : servers_) {
        if (server.second < min) {
            result = server.first;
        }
    }

    servers_[result]++; // Increment connections
    return result;
}

void LeastConnections::AddServer(ServerConfig serverConfig)
{
    // Check if server is already attached
    for (const auto &server : servers_) {
        if (server.first->serverConfig == serverConfig) {
            return;
        }
    }

    // Create and push back new attached server
    auto server = std::make_shared<AttachedServer>(
        serverConfig,
        true
    );
    servers_[server] = 0;
}

void LeastConnections::RemoveServer(ServerConfig serverConfig)
{
    // If server is attached, then detach it
    for (const auto &server : servers_) {
        if (server.first->serverConfig == serverConfig) {
            servers_.erase(server.first);
            server.first->attached.store(false);
        }
    }
}

void LeastConnections::Signal(std::shared_ptr<AttachedServer> server) {
    auto it = servers_.find(server);
    // If server is attached
    if (it != servers_.end()) {
        servers_[server]--; // Decrement connections
    }
}
