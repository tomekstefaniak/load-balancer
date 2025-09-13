#include "cli/CLI.hpp"
#include "core/LoadBalancer.hpp"
#include "config/Config.hpp"
#include <iostream>
#include <sstream>
#include <algorithm>

void CLI::StartCLI(LoadBalancer *loadBalancer)
{
    if (!loadBalancer)
    {
        std::cout << "Error: LoadBalancer is null!" << std::endl;
        return;
    }

    std::cout << "=== Load Balancer CLI ===" << std::endl;
    std::cout << "Type 'help' for available commands" << std::endl;

    std::string input;
    bool running = true;

    while (running)
    {
        printPrompt();

        if (!std::getline(std::cin, input))
        {
            // EOF or input error
            break;
        }

        // Trim whitespace
        input.erase(0, input.find_first_not_of(" \t"));
        input.erase(input.find_last_not_of(" \t") + 1);

        if (input.empty())
        {
            continue;
        }

        running = processCommand(input, loadBalancer);
    }

    std::cout << "CLI exiting..." << std::endl;
}

bool CLI::processCommand(const std::string &command, LoadBalancer *loadBalancer)
{
    auto tokens = splitCommand(command);
    if (tokens.empty())
    {
        return true;
    }

    std::string cmd = tokens[0];
    std::transform(cmd.begin(), cmd.end(), cmd.begin(), ::tolower);

    if (cmd == "help")
    {
        showHelp();
    }
    else if (cmd == "start")
    {
        startBalancer(loadBalancer);
    }
    else if (cmd == "stop")
    {
        stopBalancer(loadBalancer);
    }
    else if (cmd == "servers")
    {
        showServers(loadBalancer);
    }
    else if (cmd == "add")
    {
        if (tokens.size() >= 2)
        {
            std::string args = command.substr(command.find_first_of(' ') + 1);
            addServer(loadBalancer, args);
        }
        else
        {
            std::cout << "Usage: add <ip> <port>" << std::endl;
        }
    }
    else if (cmd == "remove")
    {
        if (tokens.size() >= 2)
        {
            std::string args = command.substr(command.find_first_of(' ') + 1);
            removeServer(loadBalancer, args);
        }
        else
        {
            std::cout << "Usage: remove <ip> <port>" << std::endl;
        }
    }
    else if (cmd == "exit" || cmd == "quit")
    {
        return false;
    }
    else
    {
        std::cout << "Unknown command: " << cmd << std::endl;
        std::cout << "Type 'help' for available commands" << std::endl;
    }

    return true;
}

void CLI::showHelp()
{
    std::cout << "\nAvailable commands:" << std::endl;
    std::cout << "  help           - Show this help message" << std::endl;
    std::cout << "  start          - Start the load balancer" << std::endl;
    std::cout << "  stop           - Stop the load balancer" << std::endl;
    std::cout << "  servers        - List all configured servers" << std::endl;
    std::cout << "  add <ip> <port> - Add a new server to the pool" << std::endl;
    std::cout << "  remove <ip> <port> - Remove a server from the pool" << std::endl;
    std::cout << "  exit           - Exit the CLI and stop the load balancer" << std::endl;
    std::cout << std::endl;
}

void CLI::startBalancer(LoadBalancer *loadBalancer)
{
    try
    {
        loadBalancer->StartWork();
        std::cout << "Load balancer started successfully!" << std::endl;
    }
    catch (const std::exception &e)
    {
        std::cout << "Failed to start load balancer: " << e.what() << std::endl;
    }
}

void CLI::stopBalancer(LoadBalancer *loadBalancer)
{
    try
    {
        loadBalancer->StopWork();
        std::cout << "Load balancer stopped successfully!" << std::endl;
    }
    catch (const std::exception &e)
    {
        std::cout << "Failed to stop load balancer: " << e.what() << std::endl;
    }
}

void CLI::showServers(LoadBalancer *loadBalancer)
{
    try
    {
        auto servers = loadBalancer->GetServers();

        if (servers.empty())
        {
            std::cout << "No servers configured" << std::endl;
            return;
        }

        std::cout << "Configured servers (" << servers.size() << "):" << std::endl;
        for (size_t i = 0; i < servers.size(); ++i)
        {
            std::cout << "  " << (i + 1) << ". " << servers[i].serverIP << ":" << servers[i].serverPort << std::endl;
        }
    }
    catch (const std::exception &e)
    {
        std::cout << "Failed to get servers: " << e.what() << std::endl;
    }
}

void CLI::addServer(LoadBalancer *loadBalancer, const std::string &args)
{
    auto tokens = splitCommand(args);
    if (tokens.size() < 2)
    {
        std::cout << "Usage: add <ip> <port>" << std::endl;
        return;
    }

    std::string ip = tokens[0];
    std::string portStr = tokens[1];

    try
    {
        uint16_t port = static_cast<uint16_t>(std::stoi(portStr));
        ServerConfig serverConfig(ip, port);

        loadBalancer->AttachServer(serverConfig);
        std::cout << "Server " << ip << ":" << port << " added successfully!" << std::endl;
    }
    catch (const std::exception &e)
    {
        std::cout << "Failed to add server: " << e.what() << std::endl;
    }
}

void CLI::removeServer(LoadBalancer *loadBalancer, const std::string &args)
{
    auto tokens = splitCommand(args);
    if (tokens.size() < 2)
    {
        std::cout << "Usage: remove <ip> <port>" << std::endl;
        return;
    }

    std::string ip = tokens[0];
    std::string portStr = tokens[1];

    try
    {
        uint16_t port = static_cast<uint16_t>(std::stoi(portStr));
        ServerConfig serverConfig(ip, port);

        loadBalancer->DettachServer(serverConfig);
        std::cout << "Server " << ip << ":" << port << " removed successfully!" << std::endl;
    }
    catch (const std::exception &e)
    {
        std::cout << "Failed to remove server: " << e.what() << std::endl;
    }
}

std::vector<std::string> CLI::splitCommand(const std::string &command)
{
    std::vector<std::string> tokens;
    std::istringstream iss(command);
    std::string token;

    while (iss >> token)
    {
        tokens.push_back(token);
    }

    return tokens;
}

void CLI::printPrompt()
{
    std::cout << "load-balancer> ";
    std::cout.flush();
}