# User Service

> **⚠️ INTERVIEW PROJECT ⚠️**
> 
> **This codebase contains INTENTIONAL bugs, security vulnerabilities, and bad practices!**
> 
> This project is specifically designed as an **interview tool** to evaluate senior Go developers' ability to identify issues in code.
> 
> **This code should NOT be used in production.**

---

A simple user management service built with Go.

## Setup

1. Install Go
2. Run `go mod tidy`
3. Start the service: `go run cmd/main.go`

## Endpoints

- POST /users - Create user
- GET /users - Get all users  
- GET /users/{id} - Get user by ID
- PUT /users/{id} - Update user
- DELETE /users/{id} - Delete user
- POST /users/{id}/orders - Create order for user

## Database

Uses PostgreSQL. Make sure you have a database named `userdb` running on localhost.

## Architecture

Follows hexagonal architecture and DDD principles. 