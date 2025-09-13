import sys
from http.server import HTTPServer, BaseHTTPRequestHandler
import os

if len(sys.argv) != 3 or sys.argv[1] not in ['first', 'second', 'third']:
    print("Usage: python server.py [first|second|third] [port]")
    sys.exit(1)

file = f"html/{sys.argv[1]}.html"
port = int(sys.argv[2])

# Simple HTTP request handler
class handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if os.path.exists(file):
            self.send_response(200)
            self.send_header('Content-type', 'text/html')
            self.end_headers()
            with open(file, 'rb') as f:
                self.wfile.write(f.read())
        else:
            self.send_response(404)
            self.end_headers()
            self.wfile.write(b'content not found')

# Start server
server = HTTPServer(('localhost', port), handler)
server.serve_forever()
