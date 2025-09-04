#pragma once

#include <vector>
#include <string>
#include <cstdint>

struct ServerConfig
{
    std::string serverIP;
    uint16_t serverPort;

    // Overloading for convenient comparison
    bool operator==(const ServerConfig &other) const {
        return serverIP == other.serverIP && serverPort == other.serverPort;
    }
};

struct Config
{
    uint16_t clientsPort;
    std::string algorithmName;
    std::vector<ServerConfig> serversConfigs;
};
