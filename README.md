# Substack Auth Service

A complete authentication service with caching capabilities, built in Go.

## Features

- **Auth Basic**: Simple authentication service on port 8080
- **Auth Improved**: Authentication service with Redis caching on port 8081
- **Precache Worker**: Background worker that preloads user data into cache
- **User Seeder**: Generate random users for testing
- **Load Testing**: K6 scripts for performance testing

## Prerequisites

- Go 1.24+
- Docker and Docker Compose
- k6 (for load testing)

## Quick Start

1. **Clone and setup**:
   ```bash
   git clone <repository>
   cd substack-cache-auth
   cp env.example .env
   ```

2. **Quick Test (Optional)**:
   ```bash
   # Start infrastructure
   make infra-up
   
   # Seed a test user
   make seeder-single USERNAME=test@katakode.com PASSWORD=test123
   
   # Test login (copy-paste ready)
   curl -X POST http://localhost:8080/login \
     -H "Content-Type: application/json" \
     -d '{"username":"test@katakode.com","password":"test123"}'
   ```

3. **Start infrastructure**:
   ```bash
   make infra-up
   ```

3. **Generate RSA keys** (already done):
   ```bash
   # Keys are already generated in ./keys/
   ```

4. **Seed users**:
   ```bash
   make seeder N=1000000
   ```

   Side notes: 
   * Check the code, it will spawn 10 go routine, if you CPU is potato, adjust accordingly.
   * Generate based on your need, to simulate real condition, 1 million users is ok
   * Usernames will be generated in format: 00000001@katakode.com, 00000002@katakode.com, etc.

4.1 **Seed one (your) user**:
```bash
   make seeder-single USERNAME=user@katakode.com PASSWORD=123
```

This is required if you want to run load test
   

5. **Run services**:
   ```bash
   # Terminal 1: Auth Basic
   make auth-basic
   
   # Terminal 2: Auth Improved  
   make auth-improved
   
   # Terminal 3: Precache Worker
   make precache-worker
   ```

## Services

### Auth Basic (Port 8080)
- Simple authentication without caching
- Direct database queries
- JWT token generation

### Auth Improved (Port 8081)
- Authentication with Redis caching
- Feature toggle for cache on/off
- Falls back to database when cache miss

### Precache Worker
- Runs every minute via cron
- Loads users in batches (configurable)
- Feature toggle for enable/disable

## Configuration

Environment variables in `.env`, adjust it accordingly.
Some configuration are hardcoded to make the code simpler for the sake of simulation.

## API Endpoints

### POST /login

#### Request Body
```json
{
  "username": "user@katakode.com",
  "password": "password"
}
```

#### Response
```json
{
  "token": "jwt_token_here",
  "user": {
    "id": 1,
    "username": "user@katakode.com",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Test with the seeded user:**
```bash
# First, seed a test user
make seeder-single USERNAME=test@katakode.com PASSWORD=test123

# Then test auth-basic
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test@katakode.com","password":"test123"}'

# Test auth-improved
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test@katakode.com","password":"test123"}'
```


## Load Testing

Run load tests with k6:

```bash
make load-test
```

Tests run for 5 minutes with 10 virtual users.

## Make Commands

- `make help` - Show available commands
- `make auth-basic` - Run auth-basic service
- `make auth-improved` - Run auth-improved service  
- `make precache-worker` - Run precache worker
- `make seeder N=1000` - Generate N users with 8-digit zero-padded usernames (00000001@katakode.com, etc.)
- `make load-test` - Run k6 load tests
- `make infra-up` - Start Redis and MySQL
- `make infra-down` - Stop infrastructure


## Security

- Passwords hashed with bcrypt
- JWT tokens signed with RSA-256
- Private keys stored securely
- No sensitive data in logs

## Performance

- Redis caching reduces database load
- Batch processing for precaching
- Configurable batch sizes
- Connection pooling for database
