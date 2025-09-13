#pragma once

#include <map>
#include <atomic>
#include "strategy/IStrategy.hpp"

class LeastConnections : public IStrategy
{
public:
    explicit LeastConnections(const std::vector<ServerConfig> &serversConfigs);

    // Get next attached server
    std::shared_ptr<ServerConfig> Next() override;

    // Add new server to the group, if already in strategy do nothing
    void AttachServer(ServerConfig serverConfig) override;

    // Remove server from the strategy, if not in strategy do nothing
    void DettachServer(ServerConfig serverConfig) override;

    // Return list of servers configs objects
    std::vector<ServerConfig> GetServers() override;

    // Signal the server on connection close - logic depends on specific strategy
    void Signal(std::shared_ptr<ServerConfig> server) override;

    // Reset strategy state
    void Reset() override;

    ~LeastConnections();

private:
    std::map<std::shared_ptr<ServerConfig>, std::atomic<uint32_t>> servers_;
};
