#pragma once

#include <vector>
#include <string>
#include <cstdint>

struct ServerConfig
{
    std::string domain;
    std::string serverIP;
    uint16_t serverPort;
};

struct Config
{
    uint16_t clientsPort;
    std::string algorithmName;
    std::vector<ServerConfig> serversConfigs;
};
