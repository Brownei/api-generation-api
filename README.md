# API Generation API

A RESTful API for user authentication and API key management.

## Deployment

**Live URL:** https://api.brownson.tech

> **Note:** If you want to run locally, Docker is recommended for fast setup. See instructions below.

## API Documentation (Postman)

Import this collection into Postman to test the API:

```json
{
  "info": {
    "name": "Peppermint API",
    "description": "API for user authentication and API key management",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080/v1"
    }
  ],
  "item": [
    {
      "name": "Health Check",
      "item": [
        {
          "name": "Health Check",
          "request": {
            "method": "GET",
            "url": "{{baseUrl}}/api/health"
          }
        }
      ]
    },
    {
      "name": "Auth",
      "item": [
        {
          "name": "Register",
          "request": {
            "method": "POST",
            "url": "{{baseUrl}}/api/auth/register",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"user@example.com\",\n  \"password\": \"password123\",\n  \"firstname\": \"John\",\n  \"lastname\": \"Doe\"\n}"
            }
          }
        },
        {
          "name": "Login",
          "request": {
            "method": "POST",
            "url": "{{baseUrl}}/api/auth/login",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"user@example.com\",\n  \"password\": \"password123\"\n}"
            }
          }
        }
      ]
    },
    {
      "name": "API Keys",
      "item": [
        {
          "name": "Create API Key",
          "request": {
            "method": "POST",
            "url": "{{baseUrl}}/api/api-key",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              },
              {
                "key": "Authorization",
                "value": "Bearer <token>"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"name\": \"My API Key\",\n  \"permissions\": [\"read\", \"write\"]\n}"
            }
          }
        },
        {
          "name": "List API Keys",
          "request": {
            "method": "GET",
            "url": "{{baseUrl}}/api/api-key",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer <token>"
              }
            ]
          }
        },
        {
          "name": "Revoke API Key",
          "request": {
            "method": "GET",
            "url": "{{baseUrl}}/api/api-key/:id",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer <token>"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/v1/api/health` | Health check | No |
| POST | `/v1/api/auth/register` | Register new user | No |
| POST | `/v1/api/auth/login` | Login user | No |
| POST | `/v1/api/api-key` | Create API key | Yes |
| GET | `/v1/api/api-key` | List all API keys | Yes |
| GET | `/v1/api/api-key/{id}` | Revoke API key | Yes |

## Run Locally

### Prerequisites

- Go 1.20+
- Docker (recommended for fast postgres setup)

### Quick Start with Docker (Recommended)

```bash
# Start postgres and run the application
make run

# Stop the postgres container when done
make stop
```

### Manual Setup

1. **Start PostgreSQL:**

   ```bash
   # Using Docker
   docker run -d --name peppermint-db \
     -e POSTGRES_USER=db \
     -e POSTGRES_PASSWORD=db \
     -e POSTGRES_DB=db \
     -p 5432:5432 \
     postgres:alpine
   ```

2. **Run the application:**

   ```bash
   go run cmd/main.go
   ```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | Database host |
| `DB_PORT` | 5432 | Database port |
| `DB_USER` | db | Database user |
| `DB_PASSWORD` | db | Database password |
| `DB_NAME` | db | Database name |
| `JWT_SECRET` | your-secret-key | JWT signing secret |
| `SERVER_PORT` | 8080 | Server port |

## License

MIT
