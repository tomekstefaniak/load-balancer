#pragma once

#include <string>
#include "core/Config.hpp"

class ConfigParser
{
public:
    static Config* ParseConfig(const std::string &configFilePath);
};
