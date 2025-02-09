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

#### Create Ranking
```
POST /api/rankings
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "Final Jury Ranking",
    "description": "jury rankings for Eurovision 2024",
    "year": 2024,
    "ranking": "foiwgu7ebqzvhrxjy.b.ddp.c4nm",
    "group_ids": ["group1", "group2"]  // optional
}
```

Response: `201 Created`

#### Get User Rankings
```
GET /api/rankings
Authorization: Bearer <token>
```

Returns an array of the authenticated user's rankings:
```json
[
    {
        "user_id": "user-uuid",
        "ranking_id": "ranking-uuid",
        "name": "Final Jury Ranking",
        "description": "jury rankings for Eurovision 2024",
        "year": 2024,
        "ranking": "foiwgu7ebqzvhrxjy.b.ddp.c4nm",
        "group_ids": ["group1", "group2"],
        "created_at": "2024-02-09T23:22:41Z"
    }
]
```

#### Update Ranking
```
PUT /api/rankings/{ranking_id}
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "Jury Ranking",
    "description": "My updated rankings",
    "year": 2024,
    "ranking": "foiwgu7ebqzvhrxjy.b.ddp.c4nm",
    "group_ids": ["group1", "group2"]
}
```