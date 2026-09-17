# Multi-tenant User Management REST API

A Go + MongoDB + Redis REST API for managing users across multiple tenants.

The API supports JWT authentication, role-based access, Redis caching, and
CSV bulk user import.

## Project Structure

The project follows a simple layered structure:

routes -> controller -> service -> repository

```text
cmd/
    server/
        main.go              Application entry point
    gentoken/
        main.go              Helper for generating test JWTs

internal/
    auth/
        JWT validation and token generation

    cache/
        Redis cache implementation

    config/
        Environment configuration

    controller/
        user_controller.go
        csv_controller.go

    db/
        mongo.go
        redis.go

    middleware/
        auth.go
        logger.go

    model/
        User and request/response structures

    repository/
        user_repository.go
        MongoDB queries and indexes

    routes/
        routes.go

    service/
        user_service.go
        csv_service.go

    utils/
        response.go