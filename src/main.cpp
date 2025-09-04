#include <iostream>
#include <exception>
#include "../include/config/ConfigParser.hpp"
#include "core/LoadBalancer.hpp"

int main(int argc, char **argv)
{
    // Set config file path
    if (argc < 2) {
        std::cout << "config file path was not provied as argument" << std::endl;
        return 1;
    }
    std::string configFilePath = argv[1];

    // Set config
    try {
        Config* config = ConfigParser::ParseConfig(configFilePath);
    } catch (const std::exception& e) {
        std::cout << e.what() << std::endl;
        return 1;
    }

    // Initialize and start load balancer
    LoadBalancer *loadBalancer = LoadBalancer::GetInstance();
    loadBalancer->StartWork();

    std::cout << "Load Balancer" << std::endl;
    return 0;
}
