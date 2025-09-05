#pragma once

#include "strategy/IStrategy.hpp"

class RoundRobin : public IStrategy
{
public:
    explicit RoundRobin(const std::vector<ServerConfig> &serversConfigs);

    // Get next attached server
    std::shared_ptr<AttachedServer> Next() override;

    // Add new server to the group, if already in strategy do nothing
    void AttachServer(ServerConfig serverConfig) override;

    // Remove server from the strategy, if not in strategy do nothing
    void DettachServer(ServerConfig serverConfig) override;

    // Return list of servers configs objects
    std::vector<ServerConfig> GetServers() override;

    // Signal the server on connection close - empty in this strategy
    void Signal(std::shared_ptr<AttachedServer> server) override;

    ~RoundRobin();

private:
    std::vector<std::shared_ptr<AttachedServer>> servers_;
    int32_t curr_;
};
