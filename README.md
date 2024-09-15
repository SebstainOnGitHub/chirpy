# Chirpy API

Chirpy, a play on the name twitter (now X), is a basic API for managing user authentication, authorization, and posting/deleting "chirps" (short messages). It includes user management, token-based authentication, and chirp management, including the creation and deletion of chirps. Additionally, it supports metrics monitoring for administrators.

## Features

- **User Authentication**: Create and manage users.
- **JWT-based Authorization**: Secure endpoints using JSON Web Tokens (JWT).
- **Chirps Management**: Create, retrieve, and delete chirps.
- **Admin Metrics**: View basic usage metrics like the number of visits.
- **Health Checks**: Simple health check endpoint to verify the server is running.
- **Database**: Stores data in a JSON-based local file (`database.json`).

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/chirpy-api.git
   cd chirpy-api
2. Install the dependencies:
   ```bash
   go mod download
3. Set up environment variables in a .env file at the root of your project:
    ```
    JWT_SECRET=your_jwt_secret
    POLKA_API_KEY=your_polka_api_key
    ```
   (Note, the polka api key is whatever you want it to be, and is just meant to represent a payment service)
4. Build and run the project:
    ```bash
    go build && ./chirpy
5. Make requests and test!


## Endpoints

### Authentication

| Method | Endpoint         | Description                                    |
|--------|------------------|------------------------------------------------|
| POST   | `/api/login`      | Create a JWT token.                            |
| POST   | `/api/refresh`    | Refresh the JWT token using a refresh token.   |
| POST   | `/api/revoke`     | Revoke access by deleting the refresh token.   |

### User Management

| Method | Endpoint          | Description                                      |
|--------|-------------------|--------------------------------------------------|
| POST   | `/api/users`       | Create a new user.                               |
| PUT    | `/api/users`       | Update user information (requires a valid JWT).  |

### Chirps Management

| Method  | Endpoint               | Description                                             |
|---------|------------------------|---------------------------------------------------------|
| POST    | `/api/chirps`           | Create a new chirp (max 140 characters).                |
| GET     | `/api/chirps`           | Retrieve all chirps or filter by author using `?author_id`. |
| GET     | `/api/chirps/{id}`      | Retrieve a single chirp by ID.                          |
| DELETE  | `/api/chirps/{chirpID}` | Delete a chirp (only the author can delete).            |

### Admin Metrics

| Method | Endpoint             | Description                                      |
|--------|----------------------|--------------------------------------------------|
| GET    | `/admin/metrics`      | View basic metrics (HTML response).              |
| GET    | `/api/metrics`        | View metrics as plain text.                      |
| POST   | `/api/reset`          | Reset the server hit metrics.                    |

### Health Check

| Method | Endpoint         | Description             |
|--------|------------------|-------------------------|
| GET    | `/api/healthz`    | Simple health check.    |

### Polka Webhooks

| Method | Endpoint                | Description                                            |
|--------|-------------------------|--------------------------------------------------------|
| POST   | `/api/polka/webhooks`    | Handle webhook events (requires valid `POLKA_API_KEY`). |
