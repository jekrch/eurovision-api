# eurovision-api :chart_with_upwards_trend:
A GO based REST API with an ElasticSearch db for receiving and reporting on eurovision rankings from https://eurovision-ranker.com

## Features

- Secure user registration with email verification
- Two-step password reset process
- JWT-based authentication
- Elasticsearch for data storage
- Kibana dashboard for data visualization

## Prerequisites

- Docker and Docker Compose
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
APP_BASE_URL=http://localhost:8080 # used for email verification links
SHORT_ID_SEED=123123 # used to generate short unique ids. can be any int64 
MAX_USER_RANKINGS=20 # indicates the max number of rankings a user may have
```

2. Start services:
```bash
docker-compose up -d
```

3. Access API at 
```
http://localhost:8080
```

4. Kibana available at 
```
http://localhost:5601
```

## API Endpoints

### Authentication

#### Initiate Registration
```
POST /auth/register/initiate
Content-Type: application/json

{
    "email": "user@example.com"
}
```

Response:
```json
{
    "message": "Please check your email to complete registration."
}
```

#### Complete Registration
```
POST /auth/register/complete
Content-Type: application/json

{
    "token": "token-from-email",
    "password": "yourpassword"
}
```

Response:
```json
{
    "message": "Registration completed successfully. You can now log in."
}
```

#### Login
```
POST /auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "yourpassword"
}
```

Response:
```json
{
    "token": "your-jwt-token"
}
```

Use this token in subsequent requests:
```
Authorization: Bearer <token>
```

#### Initiate Password Reset
```
POST /auth/password/reset
Content-Type: application/json

{
    "email": "user@example.com"
}
```

Response:
```json
{
    "message": "If your email exists in our system, you will receive password reset instructions."
}
```

#### Complete Password Reset
```
POST /auth/password/complete
Content-Type: application/json

{
    "token": "token-from-email",
    "new_password": "yournewpassword"
}
```

Response:
```json
{
    "message": "Password has been reset successfully. You can now log in with your new password."
}
```

### Rankings

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
        "ranking_id": "YawxtgErM",
        "name": "Final Jury Ranking",
        "description": "jury rankings for Eurovision 2024",
        "year": 2024,
        "ranking": "foiwgu7ebqzvhrxjy.b.ddp.c4nm",
        "group_ids": ["group1", "group2"],
        "created_at": "2024-02-09T23:22:41Z"
    }
]
```

#### Get Specific Ranking
```
GET /api/rankings/{id}
Authorization: Bearer <token>
```

Returns a specific ranking by its ID:
```json
{
    "user_id": "user-uuid",
    "ranking_id": "YawxtgErM",
    "name": "Final Jury Ranking",
    "description": "jury rankings for Eurovision 2024",
    "year": 2024,
    "ranking": "foiwgu7ebqzvhrxjy.b.ddp.c4nm",
    "public": true,
    "group_ids": ["group1", "group2"],
    "created_at": "2024-02-09T23:22:41Z"
}
```
The ranking must be either owned by the requesting user or marked public

#### Delete Ranking
```
DELETE /api/rankings/{id}
Authorization: Bearer <token>
```

Deletes a specific ranking. Returns `200` on success.

#### Update Ranking
```
PUT /api/rankings/{id}
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

## Auth features

- Passwords must be at least 8 characters long
- Email verification is required before account activation
- Rate limiting is applied to all authentication endpoints
- Password reset tokens expire after 24 hours
- JWT tokens expire after 24 hours