# Go Expert - OpenTelemetry Project

This project is part of the "Go Expert" course and demonstrates the implementation of distributed tracing using OpenTelemetry (OTEL). The architecture consists of two microservices:

1.  **`go-weather-input`**: A simple HTTP service that acts as a gateway. It receives a CEP (brazilian zip code), validates its format, and forwards the request to the `go-weather-api` service.
2.  **`go-weather-api`**: The main business logic service. It receives a CEP, calls the ViaCEP API to get the city, and then calls the WeatherAPI to get the temperature for that city.

The entire flow is instrumented with OpenTelemetry to send traces to a collector, which then forwards them to a tracing backend. This setup allows for visualization of the entire request lifecycle across multiple services.

---

## Table of Contents
- [Project Structure](#project-structure)
- [How It Works](#how-it-works)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
  - [Environment Variables (.env)](#environment-variables-env)
  - [Example .env File](#example-env-file)
- [Running the Project](#running-the-project)
- [How to Test](#how-to-test)
    - [Send a Request to the Gateway Service](#1-send-a-request-to-the-gateway-service)
    - [Visualize Traces in Zipkin](#2-visualize-traces-in-zipkin)
---

## Project Structure

```text
go-expert-otel/
|-- go-weather-input/
|   |-- cmd/
|   |   |-- server/              # Main entry point for the HTTP server
|   |   |   |-- [main.go]
|   |-- internal/
|   |   |-- handlers/            # HTTP handler for the / endpoint
|   |   |   |-- [weather.go]
|   |   |-- util/                # Validator for zip codes
|   |   |   |-- [validator.go]
|   |-- [Dockerfile]              # Containerization recipe
|   |-- [go.mod]
|   |-- [go.sum]
|-- go-weather-api/
|   |-- cmd/
|   |   |-- server/              # Main entry point for the HTTP server
|   |   |   |-- [main.go]
|   |-- internal/
|   |   |-- handlers/            # HTTP handler for the /weather endpoint
|   |   |   |-- [weather.go]
|   |   |-- services/            # Service layer for fetching data
|   |   |   |-- [viacep.go]
|   |   |   |-- [weatherapi.go]
|   |-- [Dockerfile]              # Containerization recipe
|   |-- [go.mod]
|   |-- [go.sum]
|-- otel-collector-zipkin/
|   |-- [otel-collector.yml]     # OTEL Collector configuration
|-- .env                         # Environment variables for local development
|-- [docker-compose.yml]         # Local development environment setup
|-- [README.md]
```

## How It Works

The project's architecture is orchestrated using Docker Compose and includes the following components:

-   **`go-weather-input`**: Go service (Gateway)
-   **`go-weather-api`**: Go service (Business Logic)
-   **`otel-collector`**: The OpenTelemetry Collector, responsible for receiving traces via OTLP and exporting them to Zipkin.
-   **`zipkin`**: The tracing backend, where you can visualize the spans and traces.

## Getting Started
Prerequisites
-   Go (version 1.22 or higher)
-   Docker
-   **Docker Compose V2** (version 2.x.x or higher)
-   An API key from [WeatherAPI](https://www.weatherapi.com/).

Clone the repository
```bash
git clone https://github.com/ricardolindner/go-expert-otel.git
cd go-expert-otel
```

Download the dependencies:
```bash
go mod tidy
```

## Configuration
All configuration is done via environment variables.
For local development, you should create a .env file in the project root.

### Environment Variables (.env)

**Main variables**
* `WEATHER_API_KEY`: Your API key for the weather API service.

### Example .env File
```.env
WEATHER_API_KEY=YOURKEY
```

## Running the Project

Build and run all the services. The `--build` flag ensures that the Go service images are built from the Dockerfiles.

**Note**: This project uses Docker Compose V2. The command no longer has a hyphen.
```bash
docker compose up --build
```

The services will start, and the `otel-collector` will begin listening for traces.

## How to Test

### 1. Send a Request to the Gateway Service
Use `curl` to send a request to the `go-weather-input` service (running on port 8080).

#### Success Case (Valid CEP)
This request will trigger a full distributed trace, including calls to ViaCEP and WeatherAPI.
```bash
curl -X POST -H "Content-Type: application/json" -d '{"cep": "04552000"}' http://localhost:8080/
```

Expected output:
```json
{"city":"São Paulo","temp_C":20.1,"temp_F":68.18,"temp_K":293.1}
```

#### Validation Error (Invalid CEP Format)
This request will fail fast in the `go-weather-input` service, resulting in a trace with a single span.

```bash
curl -X POST -H "Content-Type: application/json" -d '{"cep": "123"}' http://localhost:8080/
```

Expected output:
```json
{"error": "invalid zipcode"}
```

#### Business Logic Error (CEP Not Found)
This request will pass the initial validation but will fail at the `go-weather-api` service because the CEP does not exist in the ViaCEP database.

```bash
curl -X POST -H "Content-Type: application/json" -d '{"cep": "00000000"}' http://localhost:8080/
```

Expected output:
```json
{"error": "can not find zipcode"}
```

### 2. Visualize Traces in Zipkin
Open your web browser and navigate to the Zipkin UI:

[http://localhost:9411](http://localhost:9411)

- Click on the **`Find a trace`** tab.
- Click the **`RUN QUERY`** button to see the latest traces.
- Select a trace to view the distributed spans, their durations, and the flow of the request across both services.
