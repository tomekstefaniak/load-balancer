#include <fstream>
#include <sstream>
#include <algorithm>
#include <cctype>
#include <nlohmann/json.hpp>
#include "config/ConfigParser.hpp"
#include "config/Config.hpp"

Config ConfigParser::ParseConfig(const std::string &configFilePath)
{
    // Read the config file content
    std::ifstream configFile(configFilePath);
    if (!configFile.is_open())
    {
        throw std::runtime_error("could not open config file: " + configFilePath);
    }
    std::stringstream configBuffer;
    configBuffer << configFile.rdbuf(); // Read the entire config file content
    std::string configString = configBuffer.str();

    // Parse content of the json config file to Config struct
    Config config{
        "",
        0,
        std::vector<ServerConfig>()
    };
    
    try
    {
    // Parse json config file
        nlohmann::json j = nlohmann::json::parse(configBuffer.str());

    // Extract config fields
        config.clientsPort = j["clientsPort"].get<uint16_t>();

        config.algorithmName = j["algorithmName"].get<std::string>();
        std::transform(
            config.algorithmName.begin(),
            config.algorithmName.end(),
            config.algorithmName.begin(),
            [](unsigned char c) -> unsigned char
            { return std::tolower(c); }); // Convert algorithm name to lower case

        for (const auto &serverConfig : j["serversConfigs"])
        {
            config.serversConfigs.push_back(ServerConfig{
                serverConfig["serverIP"].get<std::string>(),
                serverConfig["serverPort"].get<uint16_t>()});
        }
    }
    catch (const nlohmann::json::exception &e)
    {
        throw std::runtime_error("error parsing json: " + std::string(e.what()));
    }

    // Return ready to use config
    return config;
}
