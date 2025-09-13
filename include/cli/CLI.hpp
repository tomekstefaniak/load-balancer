#pragma once

#include <string>
#include <vector>

// Forward declaration to avoid circular dependency
class LoadBalancer;

class CLI
{
public:
    /**
     * @brief Start CLI interface with LoadBalancer control
     * @param loadBalancer Pointer to LoadBalancer instance to control
     */
    static void StartCLI(LoadBalancer *loadBalancer);

private:
    // Process user command
    static bool processCommand(const std::string &command, LoadBalancer *loadBalancer);

    // Command handlers
    static void showHelp();
    static void startBalancer(LoadBalancer *loadBalancer);
    static void stopBalancer(LoadBalancer *loadBalancer);
    static void showServers(LoadBalancer *loadBalancer);
    static void addServer(LoadBalancer *loadBalancer, const std::string &args);
    static void removeServer(LoadBalancer *loadBalancer, const std::string &args);

    // Helper functions
    static std::vector<std::string> splitCommand(const std::string &command);
    static void printPrompt();
};