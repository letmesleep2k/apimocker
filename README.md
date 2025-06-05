# apimocker

A lightweight mock API server with TUI interface for serving fake JSON data and static files based on YAML/JSON configuration.  
Ideal for frontend development, testing, and API prototyping. Supports dynamic fake data generation using the faker library and static file responses like images and videos. Written in Go.

---

## Features

- Define multiple API endpoints with HTTP method, path, and response data schema  
- Generate fake JSON responses dynamically using flexible templates  
- Serve static files (images, videos, etc.) as API responses  
- Authentication support - Basic Auth and Bearer Token authentication
- Configurable via YAML or JSON file  
- Interactive TUI showing running endpoints and allowing graceful exit  
- Simple CLI interface powered by Cobra
- Support for query parameters to control response data (e.g. `?count=5&sort=name&order=desc`)
- Allows logging to a file and to the console, with the ability to select the format
- Enchanced logging with authentication details

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
logging:
  enabled: true
  format: plain
  output: stdout
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

  - path: /admin
    method: GET
    auth: 
        type: basic
        username: admin
        password: tsebehtsiogram
    data: |
        {
            "id":"uuid",
            "name":"name",
            "email":"email"
        }

  - path: /checkAccess
    method: GET
    auth:
        type: bearer
        token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJNYXJnYXJpdGEiLCJpYXQiOjE3NDkxMjg3NDAsInN1YiI6ImlzIHRoZSBiZXN0IGdpcmwifQ.YzfJg-lYeAA5av1ZQAA_dC_GTV-n8ph0HzVc0Drt6lw"
    data: |
        {
            "id": "uuid"
        }
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
 - `auth` - Authentication configuration (optional)

---

### Authentication

The `apimocker` supports two types of authentication that can be configured per endpoint:

#### Basic Authentication

Configure Basic Auth with username and password:
```yaml
auth:
    type: basic
    username: "superADMIN"
    password: "atirevoli"
```

##### Usage:
```bash
# Using curl with Basic Auth
curl -H "Authorization: Basic c3VwZXJBRE1JTjphdGlyZXZvbGk=" http://localhost:8080/admin

# Or using username:password format
curl -u superADMIN:atirevoli http://localhost:8080/admin
```

#### Bearer Token Authentication

Configure Bearer Token authentication:
```yaml
auth:
    type: bearer
    token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJNYXJnYXJpdGEiLCJpYXQiOjE3NDkxMjg3NDAsInN1YiI6ImlzIHRoZSBiZXN0IGdpcmwifQ.YzfJg-lYeAA5av1ZQAA_dC_GTV-n8ph0HzVc0Drt6lw"
```

##### Usage:
```bash
# Using curl with Bearer Token
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJNYXJnYXJpdGEiLCJpYXQiOjE3NDkxMjg3NDAsInN1YiI6ImlzIHRoZSBiZXN0IGdpcmwifQ.YzfJg-lYeAA5av1ZQAA_dC_GTV-n8ph0HzVc0Drt6lw" http://localhost:8080/users
```

#### Authentication Behavior
 - Protected endpoints: Return `401 Unauthorized` when authentication fails
 - Public endpoints: No `auth` field means the endpoint is publicly accessible
 - Authentication logging: All authentication attempts are logger with result status
 - Error responses: Failed authentication return JSON error message

#### Authentication Error Responses

When authentication fails, the server return:
```json
{
    "error": "Authentication required"
}
```
With HTTP status code `401 Unauthorized`

---

### Logging

The `apimocker` supports request logging with customizable output format and destination.

#### Cofiguration of Logging

Add the following section to your YAML/JSON config:
```yaml
logging:
  enabled: true # Enable or disable logging
  format: "plain" # "plain" or "json"
  output: "stdout" # "stdout" or file file path (e.g. "logs.txt")
```

#### Log Formats

 - `plain`:
```bash
[2025-05-31T13:15:42Z] GET /api/users?page=1 - 200 - 12ms - 127.0.0.1:49322 - 642 bytes - Auth: bearer (success)
```
 - `json`:
```json
{
  "timestamp": "2025-05-31T13:15:42Z",
  "method": "GET",
  "path": "/api/users",
  "query": "page=1",
  "status_code": 200,
  "response_time": "12ms",
  "user_agent": "curl/8.0.1",
  "remote_addr": "127.0.0.1:49322",
  "content_length": 642,
  "auth_type": "bearer",
  "auth_result": "success"
}
```

#### Authentication Log Fields

The following authentication-related fields are included in logs:
 - `auth_type`: Type of authentication used (`basic`, `bearer`, or empty for public endpoints)
 - `auth_result`: Result of authentication attempts:
    - `success`: Authentication successed
    - `no-auth`: No authentication configured for endpoint
    - `missing-auth`: No Authorization header provided
    - `invalid-credentials`: Wrong username/password for Basic Auth
    - `invalid-token`: Wrong token for Bearer Auth
    - `invalid-basic-format`: Malformed Basic Auth header
    - `invalid-bearer-format`: Malformed Bearer Token header
    - `invalid-base64`: Invalid Base64 encoding in Basic Auth
    - `invalid-credentials-format`: Invalid format in Basic Auth credentials

#### Output

 - `stdout`: logs are printed directly to the terminal.
 - `file path`: logs are appended to the specified file.

#### Example

```yaml
logging:
    enabled: true
    format: json
    output: logs/mock-api.log
```
```yaml
logging:
    enabled: false
```

---

## Supported fake data types

 - `uuid` - Universally unique identifier (e.g. `a1b2c3d4-e5f6-7890-abcd-ef1234567890`)
 - `name` - Full name of a person (e.g. `Margo Hani`)
 - `email` - Random email address (e.g. `john.gabby@example.com`)
 - `bool` - Boolean value (`true` or `false`)
 - `int` - Integer number (default range: 0-999)
 - `string` - Random single word (e.g. `lorem`, `ipsum`)
 - `lat` - Latitude value as float (e.g. `51.5074`)
 - `lng` - Longitude value as float (e.g. `-0.1278`)
 - `ipv4` - Random IPv4 address (e.g. `192.168.0.1`)
 - `url` - Random URL (e.g. `https://exmpl.com`)
 - `username` - Random username (e.g. `roufRegard`)
 - `password` - Random password string (e.g. `8D#wq2ID`)
 - `phone` - Random phone number (e.g. `+44-74-0537-1411`)
 - `date` - Random date string (format: `YYYY-MM-DD`)
 - `timestamp` - Current Unix timestamp (e.g. `1717144854`)

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
# Public endpoint
GET /users?count=10&sort=name&order=desc&filter=email:gmail.com

# Protected endpoint with Bearer token
curl -H "Authorization: Bearer mysecrettoken123" \
  "http://localhost:8080/users?count=5&meta=true"

# Protected endpoint with Basic auth
curl -u admin:secret123 \
  "http://localhost:8080/admin?count=1"
```

---

## Metadata Response
When `meta=true` is included in the query parameters, the response includes additional metadata:

```json
{
  "data": [...],
  "meta": {
    "count": 5,
    "total": 100,
    "offset": "0",
    "limit": "5",
    "sort": "name",
    "order": "asc",
    "filter": "",
    "status": 200
  }
}
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
