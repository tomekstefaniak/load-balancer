#include <sstream>
#include "config/Config.hpp"

ServerConfig::ServerConfig(std::string serverIP, uint16_t serverPort)
    : serverIP(serverIP), serverPort(serverPort)
{
    bool validationResult = ValidateIP();
    if (!validationResult) {
        throw std::runtime_error("invalid IP address:" + serverIP);
    }
}

// Overloading for convenient comparison
bool ServerConfig::operator==(const ServerConfig &other) const {
    return serverIP == other.serverIP && serverPort == other.serverPort;
}

// Validate server config
bool ServerConfig::ValidateIP() {
    // Check if there are exactly 3 dots
    int dotCount = 0;
    for (char c : serverIP) {
        if (c == '.') dotCount++;
    }
    if (dotCount != 3) {
        return false;
    }

    // Divide string
    std::vector<std::string> parts;
    std::stringstream ss(serverIP);
    std::string part;
    while (std::getline(ss, part, '.')) {
        parts.push_back(part);
    }

    // Check if there are exactly 4 parts
    if (parts.size() != 4) {
        return false;
    }

    // Check every part
    for (const std::string& p : parts) {
        // Check if part is not empty
        if (p.empty()) {
            return false;
        }
        // Make sure it contains only digits
        for (char c : p) {
            if (!isdigit(c)) {
                return false;
            }
        }
        // Check if it stars with 0, excluding case when there is only 0
        if (p.size() > 1 && p[0] == '0') {
            return false;
        }
        // Covnert digit and check its range
        try {
            int num = std::stoi(p);
            if (num < 0 || num > 255) {
                return false;
            }
        } catch (...) {
            return false;
        }
    }

    return true; // Validation succeded
}
