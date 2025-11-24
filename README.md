# Eurovision Ranker API

A containerized .NET 10.0 REST API for creating, sharing, and comparing Eurovision Song Contest rankings.

This project is designed to run as a "sidecar" application alongside other services (like Umami), using a shared PostgreSQL database with a dedicated `ranker` schema.

## Tech Stack

  * **Framework:** .NET 10.0 (Web API)
  * **Database:** PostgreSQL 15+
  * **Data Access:** Dapper (Raw SQL)
  * **Migrations:** Flyway
  * **Auth:** JWT (Bearer Token) + BCrypt
  * **Infrastructure:** Docker & Docker Compose

## Prerequisites

  * [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or Docker Engine + Compose)
  * (Optional) .NET 10.0 SDK for local non-containerized debugging

## Getting Started

### 1\. Configuration

The application relies on environment variables for configuration.

1.  Navigate to the `EurovisionRanker.Api` directory.
2.  Copy the template file:
    ```bash
    cp .env.template .env
    ```
3.  Open `.env` and configure your secrets (Database credentials, JWT Secret, etc.).

### 2\. Run with Docker Compose

This is the recommended way to run the application locally. It spins up the API, a local PostgreSQL instance, and runs Flyway migrations automatically.

```bash
docker-compose up --build
```

  * **API URL:** `http://localhost:5001` (or whatever `API_PORT` you defined in .env)
  * **Database Port:** `5488` (Exposed for local inspection tools)

### 3\. Access Documentation

Once the container is running, access the Swagger UI to test endpoints:

```
http://localhost:5001/swagger/index.html
```

*(Note: Adjust the port `5001` if you changed `API_PORT` in your `.env` file)*

## Database

This project uses a specific schema strategy to allow co-existence with other applications on a shared database.

  * **Schema:** `ranker` (All tables are isolated here)
  * **Migrations:** SQL scripts are located in `db/migration`. Flyway applies these automatically on container startup.
  * **Search Path:** The application connects with `SearchPath=ranker`, so SQL queries do not need to prefix tables (e.g., `SELECT * FROM users` works automatically).

### Core Tables

  * `users`: Auth and profile data.
  * `ranking`: User rankings for specific years (stored as text strings).
  * `groups`: Groups for comparing rankings.
  * `group_member`: Link table for users joining groups.

## Development Notes

  * **Dapper:** Dapper is used for high-performance data access. Ensure strict typing is used in DTOs.
  * **Authentication:** Endpoints are secured via JWT. Use the `/api/Auth/login` endpoint to retrieve a token, then authorize in Swagger using `Bearer <your-token>`.