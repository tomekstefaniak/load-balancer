import sys
from http.server import HTTPServer, BaseHTTPRequestHandler, SimpleHTTPRequestHandler
import threading

html = """
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>First Server</title>
    <style>
        body {{
            margin: 0;
            padding: 0;
            height: 100vh;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            background: linear-gradient(135deg, #d3d3d3 0%, #ffffff 30%, #f0f0f0 70%, #d3d3d3 100%);
            font-family: Arial, sans-serif;
        }}

        .main-text {{
            font-size: 4rem;
            font-weight: bold;
            color: #000;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
            letter-spacing: 2px;
            margin-bottom: 20px;
        }}
    </style>
</head>

<body>
    <div class="main-text">{name}</div>
</body>

</html>
"""

if len(sys.argv) != 3:
    print("Usage: python server.py <number_of_servers> <first_server_port>")
    sys.exit(1)
number_of_servers = int(sys.argv[1])
first_server_port = int(sys.argv[2])

def handler_factory(name):
    # Simple HTTP request handler that responds with server name
    class Handler(SimpleHTTPRequestHandler):
        def do_GET(self):
            self.send_response(200)
            self.send_header("Content-Type", "text/html")
            self.send_header("Connection", "close")
            self.end_headers()
            self.wfile.write(html.format(name=name).encode())
    return Handler

# Start servers
for i in range(number_of_servers):
    port = first_server_port + i
    server = HTTPServer(('localhost', port), handler_factory(f"Server {i + 1}"))
    threading.Thread(target=server.serve_forever, daemon=True).start()
    print(f"Server {i + 1} started on port {port}")

input("Press enter to stop servers...")
