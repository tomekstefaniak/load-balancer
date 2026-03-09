package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	args := os.Args[1:]

	port, args, err := parsePort(args)
	if err != nil {
		fatal(err.Error())
	}

	if len(args) == 0 {
		fatal(usage())
	}

	command := args[0]
	commandArgs := args[1:]
	base := fmt.Sprintf("http://localhost:%d", port)

	switch command {
	case "status":
		doGet(base + "/status")

	case "strategy":
		if len(commandArgs) != 1 {
			fatal("usage: client --port PORT strategy <roundrobin|leastconnections|random>")
		}
		doPost(base+"/strategy", fmt.Sprintf(`{"strategy":"%s"}`, commandArgs[0]))

	case "backend":
		if len(commandArgs) == 0 {
			fatal("usage: client --port PORT backend <add|remove|ls|conns> ...")
		}
		switch commandArgs[0] {
		case "ls":
			doGet(base + "/backend/ls")
		case "conns":
			doGet(base + "/backend/conns")
		case "add":
			if len(commandArgs) != 4 {
				fatal("usage: client --port PORT backend add ADDRESS PORT MAX_CONNECTIONS")
			}
			backendPort, err := strconv.Atoi(commandArgs[2])
			if err != nil {
				fatal("invalid backend port: " + commandArgs[2])
			}
			maxConns, err := strconv.Atoi(commandArgs[3])
			if err != nil {
				fatal("invalid max connections: " + commandArgs[3])
			}
			doPost(base+"/backend/add", fmt.Sprintf(
				`{"address":"%s","port":%d,"max_connections":%d}`,
				commandArgs[1], backendPort, maxConns,
			))
		case "remove":
			if len(commandArgs) != 3 {
				fatal("usage: client --port PORT backend remove ADDRESS PORT")
			}
			backendPort, err := strconv.Atoi(commandArgs[2])
			if err != nil {
				fatal("invalid backend port: " + commandArgs[2])
			}
			doPost(base+"/backend/remove", fmt.Sprintf(
				`{"address":"%s","port":%d}`,
				commandArgs[1], backendPort,
			))
		default:
			fatal("unknown backend subcommand: " + commandArgs[0])
		}

	case "timeout":
		if len(commandArgs) != 1 {
			fatal("usage: client --port PORT timeout SECONDS")
		}
		timeout, err := strconv.Atoi(commandArgs[0])
		if err != nil {
			fatal("invalid timeout: " + commandArgs[0])
		}
		doPost(base+"/timeout", fmt.Sprintf(`{"idle_timeout_sec":%d}`, timeout))

	case "max-connections":
		if len(commandArgs) != 1 {
			fatal("usage: client --port PORT max-connections COUNT")
		}
		maxConns, err := strconv.Atoi(commandArgs[0])
		if err != nil {
			fatal("invalid max connections: " + commandArgs[0])
		}
		doPost(base+"/max-connections", fmt.Sprintf(`{"max_connections":%d}`, maxConns))

	case "stop":
		if len(commandArgs) != 1 {
			fatal("usage: client --port PORT stop <graceful|immediate>")
		}
		mode := commandArgs[0]
		if mode != "graceful" && mode != "immediate" {
			fatal("stop mode must be 'graceful' or 'immediate'")
		}
		doPost(base+"/stop", fmt.Sprintf(`{"mode":"%s"}`, mode))

	case "start":
		doPost(base+"/start", `{}`)

	case "shutdown":
		timeoutSec := 5
		if len(commandArgs) > 0 {
			t, err := strconv.Atoi(commandArgs[0])
			if err != nil {
				fatal("invalid timeout: " + commandArgs[0])
			}
			timeoutSec = t
		}
		resp, err := http.Post(
			base+"/shutdown", "application/json",
			bytes.NewBufferString(fmt.Sprintf(`{"timeout_sec":%d}`, timeoutSec)),
		)
		if err != nil {
			// Connection reset = server shut down successfully
			fmt.Println("shutdown complete")
			return
		}
		resp.Body.Close()
		fmt.Println("shutdown complete")

	default:
		fatal("unknown command: " + command + "\n" + usage())
	}
}

func parsePort(args []string) (int, []string, error) {
	for i, arg := range args {
		if arg == "--port" && i+1 < len(args) {
			port, err := strconv.Atoi(args[i+1])
			if err != nil {
				return 0, nil, fmt.Errorf("invalid port: %s", args[i+1])
			}
			remaining := make([]string, 0, len(args)-2)
			remaining = append(remaining, args[:i]...)
			remaining = append(remaining, args[i+2:]...)
			return port, remaining, nil
		}
		if strings.HasPrefix(arg, "--port=") {
			port, err := strconv.Atoi(arg[7:])
			if err != nil {
				return 0, nil, fmt.Errorf("invalid port: %s", arg[7:])
			}
			remaining := make([]string, 0, len(args)-1)
			remaining = append(remaining, args[:i]...)
			remaining = append(remaining, args[i+1:]...)
			return port, remaining, nil
		}
	}
	return 0, nil, fmt.Errorf("--port is required\n%s", usage())
}

func doGet(url string) {
	resp, err := http.Get(url)
	if err != nil {
		fatal("request failed: " + err.Error())
	}
	defer resp.Body.Close()
	printResponse(resp)
}

func doPost(url string, jsonBody string) {
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(jsonBody))
	if err != nil {
		fatal("request failed: " + err.Error())
	}
	defer resp.Body.Close()
	printResponse(resp)
}

func printResponse(resp *http.Response) {
	body, _ := io.ReadAll(resp.Body)
	fmt.Print(string(body))
	if resp.StatusCode >= 400 {
		os.Exit(1)
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func usage() string {
	return `usage: client --port PORT <command> [args]

commands:
	status                                      show balancer status
	strategy <roundrobin|leastconnections|random>  change load balancing strategy
	backend add <address> <port> <max_conns>    add a backend
	backend remove <address> <port>             remove a backend
	backend ls                                  list backends
	backend conns                               show connections per backend
	timeout <seconds>                           set idle timeout
	max-connections <count>                     set max connections
	stop <graceful|immediate>                   stop the balancer
	start									    start the balancer
	shutdown [timeout_sec]                     shutdown the whole application (default: 5s)`
}
