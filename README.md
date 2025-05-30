# apimocker

A lightweight mock API server with TUI interface for serving fake JSON data and static files based on YAML/JSON configuration.  
Ideal for frontend development, testing, and API prototyping. Supports dynamic fake data generation using the faker library and static file responses like images and videos. Written in Go.

---

## Features

- Define multiple API endpoints with HTTP method, path, and response data schema  
- Generate fake JSON responses dynamically using flexible templates  
- Serve static files (images, videos, etc.) as API responses  
- Configurable via YAML or JSON file  
- Interactive TUI showing running endpoints and allowing graceful exit  
- Simple CLI interface powered by Cobra

---

## Installation

### Install via AUR:

```bash
yay -S apimocker
```

### Build from source:

```bash
git clone https://github.com/yourusername/apimocker.git
cd apimocker
go build -o apimocker main.go
sudo mv apimocker /usr/bin/
```

---

## Usage

Run the mock server specifying the configuration file:

```bash
./apimocker -c path/to/mock.yaml
```

By default, it looks for mock.yaml in the current directory.

---

## Configuration

The config file (YAML or JSON) defines the server port and the endpoints to mock.

Example `mock.yaml`:

```yaml
port: 8080
endpoints:
  - path: /users
    method: GET
    count: 5
    data: '{"id": "uuid", "name": "name", "email": "email"}'

  - path: /image
    method: GET
    file: "./static/sample.jpg"
```

### Endpoint fields

 - `path` — URL path of the endpoint
 - `method` — HTTP method (GET, POST, etc.)
 - `count` — number of fake records to generate
 - `data` — JSON schema describing fields and their fake types (see supported types below)
 - `file` — path to static file to serve instead of JSON data

---

## Supported fake data types

 - `uuid`
 - `name`
 - `email`
 - `bool`
 - `int`
 - `string`
 - `lat` (latitude)
 - `lng` (longitude)
 - `ipv4`
 - `url`
 - `username`
 - `password`
 - `phone`
 - `date`
 - `timestamp` (current UNIX timestamp)

---

## Controls
 - Press q or Ctrl+C in the terminal UI to quit the application gracefully.
