#include <gtest/gtest.h>
#include "utils/ConfigParser.hpp"

TEST(ConfigParserTest, ParseConfigFile) {
    // Config file
    std::string configFile = "../config.json"; // Relative path

    // Parse contents of config file
    Config *config = ConfigParser::ParseConfig(configFile);
    std::cout << "\033[38;5;41msuccessfuly parsed config:\033[0m" << std::endl;

    // Display parsed configuration
    std::cout << "\tclientsPort: " << config->clientsPort << std::endl;
    std::cout << "\talgorithmName: " << config->algorithmName << std::endl;
    for (const auto &serverConfig: config->serversConfigs) {
        std::cout << "\t*" << std::endl;
        std::cout << "\t\tdomain: " << serverConfig.domain << std::endl;
        std::cout << "\t\tserverPort: " << serverConfig.serverPort << std::endl;
        std::cout << "\t\tserverIP: " << serverConfig.serverIP << std::endl;
    }
    std::cout << std::endl;
}
