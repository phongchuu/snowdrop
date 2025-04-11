### Snowdrop

#### Overview
Snowdrop is a modular backend application built using Go. It provides various features such as user management, authentication, health checks, OpenAPI documentation, and static file serving. The application is structured into multiple modules, each responsible for a specific functionality.

#### Features
- **User Management**: Handles user-related operations such as creating, retrieving, and managing user data.
- **Authentication**: Provides routes for user registration and login.
- **Health Checks**: Includes a `/healthz` endpoint to check the service's health.
- **OpenAPI Documentation**: Serves API documentation at `/api/openapi`.
- **Static File Serving**: Serves static files from the `resources/public/static` directory.

#### Modules
1. **Core**: Provides configuration management and utility functions.
2. **Auth**: Handles user authentication and registration.
3. **Healthz**: Implements health check routes.
4. **OpenAPI**: Serves API documentation.
5. **StaticFile**: Manages static file serving.
6. **User Management**: Manages user-related operations.

#### Development Setup
1. Clone the repository.
2. Install Go (version 1.24.0 or later).
3. Run `go mod tidy` in each module directory to install dependencies.
4. Use `docker-compose` to set up the database and other services:
   ```bash
   docker-compose up
   ```
5. Run the application:
   ```bash
   go run backend/console/main.go
   ```

#### Testing
- Unit tests are located in the `backend/testing` directory.
- Run tests using:
  ```bash
  go test ./...
  ```

#### Admin Account
The admin account credentials are configured via environment variables. Ensure the following variables are set in your `.env.dev` file:
- `SNOWDROP_ADMIN_USERNAME`: Admin username.
- `SNOWDROP_ADMIN_PASSWORD`: Admin password.

> **Note**: Please change the default credentials after the first login for security purposes.

#### API Endpoints
- **Health Check**: `GET /healthz`
- **OpenAPI Documentation**: `GET /api/openapi`
- **User Registration**: `POST /auth/register`
- **User Login**: `POST /auth/login`

#### Environment Variables
Environment variables are defined in the `.env.dev` file. Key variables include:
- `POSTGRES_DB_SUPER_USER`: PostgreSQL superuser name.
- `POSTGRES_DB_SUPER_PASSWORD`: PostgreSQL superuser password.
- `SNOWDROP_DB_NAME`: Database name for Snowdrop.
- `SNOWDROP_DB_ADMIN_USER`: Database admin username.
- `SNOWDROP_DB_ADMIN_PASSWORD`: Database admin password.
- `SNOWDROP_ADMIN_USERNAME`: Admin username for the application.
- `SNOWDROP_ADMIN_PASSWORD`: Admin password for the application.

#### Code Structure
- **Backend**: Contains the main application logic.
  - `common`: Shared modules and utilities.
  - `console`: Entry point for the application.
  - `internal/snowdrop`: Core framework and utilities.
  - `testing`: Unit tests and mocks.
- **Resources**: Static files, translations, and database migrations.

#### Contributing
1. Fork the repository.
2. Create a feature branch.
3. Commit your changes with clear messages.
4. Submit a pull request.

#### License
This project is licensed under the MIT License.
