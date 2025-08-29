cmake_minimum_required(VERSION 3.20)

# Add vcpkg toolchain file
set(CMAKE_TOOLCHAIN_FILE "$ENV{HOME}/.vcpkg/scripts/buildsystems/vcpkg.cmake" CACHE STRING "")

project(LoadBalancer LANGUAGES CXX)

# C++ settings
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find libraries
find_package(nlohmann_json CONFIG REQUIRED)
find_package(GTest CONFIG REQUIRED)

# Include directories
include_directories(include)

# All source files
set(SOURCES
        src/cli/CLI.cpp
        src/core/LoadBalancer.cpp
        src/core/Server.cpp
        src/core/SessionsWorker.cpp
        src/core/algos/RoundRobin.cpp
        src/core/algos/LeastConnections.cpp
        src/utils/ConfigParser.cpp
)

set(TESTS
        ${SOURCES}
        tests/ConfigParserTest.cpp
)

# Create loadbalancer executable
add_executable(loadbalancer src/main.cpp ${SOURCES})
# Create example executable
add_executable(example example/example.cpp ${SOURCES})
# Create tests executable
add_executable(tests ${TESTS})

# Link libraries with loadbalancer
target_link_libraries(loadbalancer
        nlohmann_json::nlohmann_json
)

# Link libraries with example
target_link_libraries(example
        nlohmann_json::nlohmann_json
)

# Link libraries with tests
target_link_libraries(tests
        nlohmann_json::nlohmann_json
        GTest::gtest
        GTest::gtest_main
)
