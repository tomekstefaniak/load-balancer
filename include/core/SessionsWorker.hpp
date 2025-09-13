#pragma once

#include <boost/bimap.hpp>
#include <boost/asio/ip/tcp.hpp>
#include <memory>
#include <shared_mutex>
#include <unordered_map>

using skt = boost::asio::ip::tcp::socket;
using sktPtr = std::shared_ptr<skt>;

// Forward declarations to avoid circular dependency
class LoadBalancer;
struct ServerConfig;

class SessionsWorker
{
public:
    SessionsWorker(LoadBalancer *loadBalancer);

    // Context, event handler for provided sockets
    std::shared_ptr<boost::asio::io_context> workerContext;

    // Passing through messages on event loop on provided sockets
    void StartWork();

    // Add new sockets pair to sessions worker
    void NewSession(skt client, skt server, std::shared_ptr<ServerConfig> serverConfig);

private:
    // Pointer to load balancer for signaling
    LoadBalancer *loadBalancer_;

    // Mutex for synchronization of socket pairs
    std::shared_mutex socketsBimapMutex_;
    // Bimap of client & server sockets pairs
    boost::bimap<sktPtr, sktPtr> socketsBimap_;
    // Map to track server config for each client socket
    std::unordered_map<sktPtr, std::shared_ptr<ServerConfig>> clientToServer_;

    // Start relay between two sockets
    void startRelay(sktPtr from, sktPtr to);

    // Close connection and cleanup
    void closeConnection(sktPtr socket1, sktPtr socket2);
};
