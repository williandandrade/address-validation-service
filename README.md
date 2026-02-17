# Address Validation Service

REST API for normalizing and validating US addresses. Accepts free-form address strings and returns structured, standardized address components with confidence metadata.

## Prerequisites

### Local Development

- **Go 1.25+** ([install](https://go.dev/doc/install))
- **golangci-lint** (optional, for linting): `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **mockgen** (optional, for regenerating mocks): `go install go.uber.org/mock/mockgen@latest`

> **Note:** The local build uses a pure Go address parser. For the full gopostal/libpostal parser, use the Docker setup below.

### Docker

- **Docker** ([install](https://docs.docker.com/get-docker/))

The Docker image builds with libpostal, providing production-grade address parsing.

## Getting Started

### 1. Clone and setup

```bash
git clone https://github.com/williandandrade/address-validation-service.git
cd address-validation-service
```

### 2. Install dependencies

```bash
make deps
```

### 3. Run locally

```bash
make run
```

The server starts on `http://localhost:8080`.

### 4. Test the endpoint

```bash
curl -s -X POST http://localhost:8080/api/v1/validate-address \
  -H "Content-Type: application/json" \
  -d '{"address": "123 main st, new york, ny 10001"}' | jq
```

Expected response:

```json
{
  "success": true,
  "address": {
    "street_address": "123 Main St",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "address_type": "standard_street",
    "formatted_address": "123 Main St, New York, NY 10001"
  },
  "confidence": {
    "state_confidence": "direct",
    "city_confidence": "direct",
    "postal_confidence": "direct"
  },
  "corrections_applied": ["Standardized capitalization"],
  "message": "Address validated successfully"
}
```

## Running with Docker

### Build the image

```bash
make docker-build
```

> The first build takes a few minutes as it compiles libpostal and downloads its training data (~2 GB).

### Run the container

```bash
make docker-run
```

### Test it

```bash
curl -s -X POST http://localhost:8080/api/v1/validate-address \
  -H "Content-Type: application/json" \
  -d '{"address": "456 oak ave, los angeles, ca 90210"}' | jq
```

## Development

### Available Make targets

| Command | Description |
|---------|-------------|
| `make build` | Compile the binary to `./build/` |
| `make run` | Run the server locally |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with HTML coverage report |
| `make lint` | Run golangci-lint |
| `make check` | Run lint + test |
| `make fmt` | Format Go source files |
| `make docker-build` | Build Docker image |
| `make docker-run` | Run Docker container |
| `make clean` | Remove build artifacts |
| `make deps` | Download and tidy dependencies |
| `make dev-tools` | Install golangci-lint and mockgen |

### Project structure

```
cmd/server/main.go                          # Entry point
internal/
  api/
    dto/                                    # Request/response DTOs
    handler/                                # HTTP handlers
  domain/
    entity/                                 # Address entity, validation rules
    errors/                                 # Domain error types
  usecase/                                  # Business logic, repository interface
  infrastructure/
    address_parser/                         # Address parsing implementations
tests/integration/                          # Integration tests
specs/001-address-normalization/            # Feature specification and contracts
  contracts/openapi.yaml                    # OpenAPI 3.0 spec
```

### Parser implementations

The service ships with two address parsers, selected at build time:

| Parser | Build tag | CGO | Use case |
|--------|-----------|-----|----------|
| Regex (default) | none | No | Local development, CI |
| gopostal | `-tags gopostal` | Yes (requires libpostal) | Docker, production |

Both implement the same `ValidateAddressRepository` interface, so the rest of the codebase is unaffected.

## API Reference

### `POST /api/v1/validate-address`

Normalizes a free-form US address into structured components.

**Request:**

```json
{ "address": "123 main st, new york, ny 10001" }
```

**Responses:**

| Status | Meaning |
|--------|---------|
| `200` | Address normalized successfully |
| `400` | Missing or invalid request body |
| `422` | Address could not be parsed |

See [`specs/001-address-normalization/contracts/openapi.yaml`](specs/001-address-normalization/contracts/openapi.yaml) for the full schema.

## Configuration

Environment variables (see `configs/.env.example`):

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_NAME` | `address-validation-service` | Application name |
| `APP_VERSION` | `0.1.0` | Application version |
| `LOG_LEVEL` | `debug` | Log level |
| `HTTP_PORT` | `8080` | Server port |
| `SHUTDOWN_GRACE_PERIOD` | `30s` | Graceful shutdown timeout |
| `REQUEST_TIMEOUT` | `10s` | Request timeout |
