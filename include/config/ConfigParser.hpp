#pragma once

#include <string>
#include "Config.hpp"

class ConfigParser
{
public:
    static Config ParseConfig(const std::string &configFilePath);
};
