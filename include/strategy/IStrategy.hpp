#pragma once

#include <memory>
#include "strategy/AttachedServer.hpp"

class IStrategy
{
public:
    // Get next attached server
    virtual std::shared_ptr<AttachedServer> Next() = 0;

    // Add new server to the group, if already in strategy do nothing
    virtual void AttachServer(ServerConfig serverConfig) = 0;

    // Remove server from the strategy, if not in strategy do nothing
    virtual void DettachServer(ServerConfig serverConfig) = 0;

    // Return list of servers configs objects
    virtual std::vector<ServerConfig> GetServers() = 0;

    // Signal the server on connection close
    virtual void Signal(std::shared_ptr<AttachedServer> server) = 0;

    virtual ~IStrategy() = 0;
};

// Even pure virtual destructors need an implementation
inline IStrategy::~IStrategy() = default;
