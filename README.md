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
- Support for query parameters to control response data (e.g. `?count=5&sort=name&order=desc`)

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
apimocker -c path/to/mock.yaml
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
    status: 200
    delay: "500ms"
    headers:
        X-Custom-Header: "ReverofAtir!"
    errors:
        - probability: 0.2
          status: 500
          message: "Internal Server Error"
        - probability: 0.1
          status: 403
          message: "Forbidden"
          

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
 - `status` - HTTP response status code (default 200)
 - `delay` - Response delay (`300ms`, `2s`, `1m`, etc.)
 - `headers` - Custom HTTP headers
 - `errors` - Probabilistic errors - an array of `probability`, `status`, `message`

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

## Query Parameters

Dynamic JSON endpoints support optional query parameters to customize the response:

| Parameter       | Description                                     |
| --------------- | ----------------------------------------------- |
| `count`/`limit` | Number of items to return                       |
| `offset`        | Number of items to skip                         | 
| `sort`          | Field name to sort by                           |
| `order`         | `asc` (default) or `desc`                       |
| `filter`        | Filter record by a field, format: `field:value` |

### Example usage:

```bash
GET /usrs?count=10&sort=name&order=desc&filter=email:gmail.com
```

---

## Static files

If the endpoint contains a file field, the server will simply return the contents of the file with the correct MIME type.

### Supported formats:
 - Images: `.jpg`, `.jpeg`, `.png`, `.gif`
 - Videos: `.mp4`
 - Other: `application/octet-stream`

---

## Controls
 - Press q or Ctrl+C in the terminal UI to quit the application gracefully.
