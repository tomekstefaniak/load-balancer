#include <iostream>
#include <exception>
#include <thread>
#include <chrono>
#include "../include/config/ConfigParser.hpp"
#include "core/LoadBalancer.hpp"
#include "cli/CLI.hpp"

int main(int argc, char **argv)
{
    // Set config file path
    if (argc < 2)
    {
        std::cout << "Config file path not provided as argument" << std::endl;
        return 1;
    }
    std::string configPath = argv[1];

    // Set config
    try
    {
        Config config = ConfigParser::ParseConfig(configPath);

        // Initialize and start load balancer
        LoadBalancer *loadBalancer = LoadBalancer::SetInstance(config);

        std::cout << "Load Balancer initialized with config from: " << configPath << std::endl;
        std::cout << "Type 'help' to see available commands." << std::endl;

        // Start CLI
        CLI::StartCLI(loadBalancer);

        std::cout << "Shutting down Load Balancer..." << std::endl;
    }
    catch (const std::exception &e)
    {
        std::cout << "Error: " << e.what() << std::endl;
        return 1;
    }

    return 0;
}
