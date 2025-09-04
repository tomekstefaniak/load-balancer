#include <stdexcept>
#include "strategy/RoundRobin.hpp"
#include "config/Config.hpp"

RoundRobin::RoundRobin(const std::vector<ServerConfig> &serversConfigs)
    : servers_(), curr_(0)
{
    // Create attached server for every server config
    for (const auto &serverConfig : serversConfigs) {
        auto attachedServer = std::make_shared<AttachedServer>(
            serverConfig,
            true
        );
        servers_.push_back(attachedServer);
    }
}

std::shared_ptr<AttachedServer> RoundRobin::Next()
{
    // Check is there are any attached servers
    if (servers_.empty()) {
        throw std::runtime_error("RoundRobin: no servers");
    }

    auto result = servers_[curr_];
    curr_ = (curr_ + 1 >= servers_.size()) ? 0 : curr_ + 1; // Update curr
    return result;
}

void RoundRobin::AddServer(ServerConfig serverConfig)
{
    // Check if server is already attached
    for (auto &server : servers_) {
        if (server->serverConfig == serverConfig) {
            return;
        }
    }

    // Create and push back new attached server
    auto server = std::make_shared<AttachedServer>(
        serverConfig,
        true
    );
    servers_.push_back(server);
}

void RoundRobin::RemoveServer(ServerConfig serverConfig)
{
    // If server is attached, then detach it
    for (int i = 0 ; i < servers_.size() ; i++) {
        if (servers_[i]->serverConfig == serverConfig) {
            servers_.erase(servers_.begin() + i);
            if (i < curr_) {
                curr_--;
            }
            if (curr_ >= servers_.size()) {
                curr_ = 0;
            }
        }
    }
}

void RoundRobin::Signal(std::shared_ptr<AttachedServer> server) { }
