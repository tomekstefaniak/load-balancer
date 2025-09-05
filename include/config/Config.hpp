#pragma once

#include <vector>
#include <string>
#include <cstdint>

struct ServerConfig
{
public:
    std::string serverIP;
    uint16_t serverPort;

    /**
     * @brief Validates ip and creates new ServerConfig
     * @throws std::runtime_exception If validation failed
     */
    ServerConfig(std::string serverIP, uint16_t serverPort);

    // Overloading for convenient comparison
    bool operator==(const ServerConfig &other) const;

private:
    // Validate server config
    bool ValidateIP();
};

struct Config
{
    uint16_t clientsPort;
    std::string algorithmName;
    std::vector<ServerConfig> serversConfigs;
};
