#pragma once

#include <memory>
#include "strategy/AttachedServer.hpp"

class IStrategy
{
public:
    // Get next attached server
    virtual std::shared_ptr<AttachedServer> Next() = 0;

    // Add new server to the group, if already in strategy do nothing
    virtual void AddServer(ServerConfig serverConfig) = 0;

    // Remove server from the strategy, if not in strategy do nothing
    virtual void RemoveServer(ServerConfig serverConfig) = 0;

    // Signal the server on connection close
    virtual void Signal(std::shared_ptr<AttachedServer> server) = 0;

    virtual ~IStrategy() = 0;
};
