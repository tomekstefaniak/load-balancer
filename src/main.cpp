#include <iostream>
#include <exception>
#include "utils/ConfigParser.hpp"
#include "core/LoadBalancer.hpp"

int main(int argc, char **argv)
{
    // Set config file path
    if (argc < 2) {
        std::cout << "config file path was not provied as argument" << std::endl;
    }
    std::string configFilePath = argv[1];

    // Set config
    try {
        Config* config = ConfigParser::ParseConfig(configFilePath);
    } catch (const std::exception& e) {
        std::cout << e.what() << std::endl;
    }

    // Initialize and start load balancer
    LoadBalancer *loadBalancer = LoadBalancer::GetInstance();
    loadBalancer->Start();

    std::cout << "Load Balancer" << std::endl;
    return 0;
}
