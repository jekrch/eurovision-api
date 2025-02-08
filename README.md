# eurovision-voter :chart_with_upwards_trend:
A GO based REST API with an ElasticSearch db for receiving and reporting on eurovision rankings from https://eurovision-ranker.com

## Features

- User registration with email confirmation
- JWT-based authentication
- Elasticsearch for data storage
- Kibana dashboard for data visualization

## Prerequisites

- Docker and Docker Compose
- Go 1.21 or later
- Valid SMTP configuration for email sending

## Setup

1. Create a `.env` file with your configuration:
```bash
PORT=8080 # port for API
SMTP_HOST=your-smtp-host
SMTP_PORT=587
EMAIL_USERNAME=your-username
EMAIL_PASSWORD=your-password
JWT_SECRET=your-secret-key
```

3. Start services:
```bash
docker-compose up -d
```

4. Access API at 
```
http://localhost:8080
```

5. Kibana available at 
```
http://localhost:5601
```

## API Endpoints

### Register User
```
POST /auth/register
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "yourpassword"
}
```

### Confirm Registration
```
GET /auth/confirm?token=confirmation-token
```

### Login
```
POST /auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "yourpassword"
}
```

Returns a JWT token to be used in subsequent requests:
```
Authorization: Bearer <token>
```
