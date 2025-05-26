# worker-pool-service
A service that processes tasks (jobs) sent to a REST API via JSON using a pool of concurrent workers

# Features
* Create new "jobs" with type and payload
* Check the status of current "jobs" by ID
* Get statistics about the task scheduler

# Tech Stack
* Language: Go 1.21+
* Router: chi
* UUIDs: github.com/google/uuid
* Testing: Go standard library

# Project Structure
```
/worker-pool-service
├── cmd/              # Entry point (main.go)
├── internal/
│   ├── handler/      # HTTP handlers
│   ├── model/        # Data types and validation
│   ├── service/      # Business logic
│   └── pool/         # Pool of concurrent workers
├── test/             # Test files
└── go.mod
```

# Running the Service
```go run ./cmd/server```

# Example Usage (cURL)
## Create a waypoint
```
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "type": "sleep",
    "payload": {
        "duration": "5s"
    }
}'
```

## List jobs by id
```curl http://localhost:8080/jobs/{id}```

## List all jobs
```curl http://localhost:8080/jobs```

## Get generalized stats about the task scheduler service
```curl http://localhost:8080/stats```

# Design Considerations
* Dependency Injection is used for loose coupling between components.
* Interface-Driven Architecture enables testability and future extensibility (e.g., database-backed repo).
* Validation is handled at the request model level to separate concerns cleanly.
* The service layer enforces any domain-specific business rules.

# Tests
`go test ./...`
Tests cover handler logic, service behavior, and in-memory repo operations.

# Future Improvements / Next Steps
TBD

# Time Spent
About 6 hours with a break for lunch :)

# Author
David Nakolan - david.nakolan@gmail.com
