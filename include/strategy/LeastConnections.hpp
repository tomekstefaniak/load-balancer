#pragma once

#include <map>
#include "strategy/IStrategy.hpp"

class LeastConnections : public IStrategy
{
public:
    explicit LeastConnections(const std::vector<ServerConfig> &serversConfigs);

    // Get next attached server
    std::shared_ptr<AttachedServer> Next() override;

    // Add new server to the group, if already in strategy do nothing
    void AddServer(ServerConfig serverConfig) override;

    // Remove server from the strategy, if not in strategy do nothing
    void RemoveServer(ServerConfig serverConfig) override;

    // Signal the server on connection close - logic depends on specific strategy
    void Signal(std::shared_ptr<AttachedServer> server) override;

private:
    std::map<std::shared_ptr<AttachedServer>, uint32_t> servers_;
};
