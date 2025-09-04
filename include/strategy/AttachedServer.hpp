#pragma once

#include <atomic>
#include "config/Config.hpp"

struct AttachedServer {
    ServerConfig serverConfig;
    std::atomic<bool> attached;
};
