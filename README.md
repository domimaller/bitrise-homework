# Task Executor Application

This application consists of two main components that work together to manage and execute tasks. User can create task where shell commands can be given, those are executed by a simple executor in its shell.

## backend-api-server
  An API server where tasks can be created and their statuses retrieved. It exposes endpoints to create new tasks, query their status, and manage task execution.

### Endpoints

#### User Endpoints

- POST /tasks: Create a task with a command.
- GET /tasks: List all created tasks with their states.
- GET /tasks/<resource_id>: Retrieve details of a specific task by its resource ID.

#### Internal Endpoints (for Executor Agents)

These endpoints are intended to be called only by executor agents. In production, access could be restricted using an ingress controller or firewall rules to prevent external access.

- GET /tasks/pick: Allows an executor agent to pick a task for execution. If there are queued tasks available, this endpoint always returns the queued task that was created the earliest
- POST /tasks/<resource_id>/finish: Called by an executor agent to update the state of an executed task.

## task-exec-agent
  A client application that periodically polls the backend API server for new tasks. When a task is received, the agent executes it and updates its state with the result. Note that each executor agent can execute only one task at a time.

## Solution Approach

The application was designed considering two approaches:
- **Pull Model (Implemented):**  
  Executor agents periodically poll the backend API server to retrieve tasks. This approach is resilient because if one polling cycle fails or a message is lost, the next cycle will still pick up the task.
- **Push Model (Alternative):**  
  In a push-based design, agents would subscribe to the backend API server and receive tasks immediately when they are created. This could result in more efficient task execution and easier on-demand scaling.

## Prerequisites

Before running the application, ensure that you have the following installed on your system:
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Configuration

The application can be configured via environment variables. Important variables include:

- **SERVER_PORT** (backend-api-server): The port on which the API server listens.
- **LOG_LEVEL:** Set to `info` or `debug` to control the verbosity of the logs.
- **DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME:** PostgreSQL configuration parameters.
- **POLL_INTERVAL** (task-exec-agent): Interval between polling requests for new tasks.

*Other environment variables can be modified if necessary, but these are the essential ones for the default setup.*

## Docker Compose Setup

The application is containerized using Docker Compose. Below is the provided `docker-compose.yaml` file:

```yaml
version: "3.9"

services:
  backend-api-server:
    build:
      context: ./backend-api-server
      dockerfile: Dockerfile
    ports:
      - "3500:3500"
    environment:
      - SERVER_PORT=3500
      - LOG_LEVEL=info
      - DB_USER=postgres
      - DB_PASSWORD=secretpass
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=postgres
    depends_on:
      db:
        condition: service_healthy
    healthcheck:
      test: ["CMD-SHELL", "nc -z db 5432"]
      interval: 10s
      timeout: 5s
      retries: 5

  task-exec-agent:
    build:
      context: ./task-exec-agent
      dockerfile: Dockerfile
    ports:
      - "3000"
    environment:
      - BACKEND_API_HOST=backend-api-server
      - BACKEND_API_PORT=3500
      - POLL_INTERVAL=5s
      - LOG_LEVEL=info
    depends_on:
      backend-api-server:
        condition: service_healthy
    healthcheck:
      test: ["CMD-SHELL", "curl --fail http://backend-api-server:3500 || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5

  db:
    image: postgres:latest
    restart: always
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: secretpass
      POSTGRES_DB: postgres
    ports:
      - "5432:5432"
    volumes:
      - db-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  db-data:
```

### Explanation of Key Points:

**Service Communication:** Services communicate using their service names (e.g., the API server connects to the database using DB_HOST=db).

**Healthchecks:** Each service defines a healthcheck to ensure dependencies are ready before starting.

**Volumes:** The db-data named volume persists PostgreSQL data. Removing this volume (via docker-compose down -v) will reset the database.

## Makefile

A Makefile is provided to simplify common tasks. Below is the content of the Makefile along with explanations:

```
.PHONY: run down down-prune-db logs scale-agents

# Build images and run all services in detached mode.
run:
	docker-compose up -d --build

# Stop all running containers.
down:
	docker-compose down

# Stop all containers and remove persistent database data.
down-prune-db:
	docker-compose down -v

# Follow logs from all services.
logs:
	docker-compose logs -f

# Scale the task-exec-agent service.
# Usage: make scale-agents num=3
scale-agents:
	@echo "Scaling task-exec-agent to $(num) instance(s)..."
	docker-compose up -d --scale task-exec-agent=$(num)
```

### Makefile Command Descriptions:

```run```: Builds the images (if necessary) and starts all services.

```down```: Stops and removes all containers.

```down-prune-db```: Stops the application and deletes the database data (by removing the db-data volume). Use this when you want a fresh database on the next run.

```logs```: Tails the logs from all services to help with debugging.

```scale-agents```: Scales the number of task-exec-agent instances. For example, run make scale-agents num=3 to start three agent containers. 

## Running the Application

1. Build and Start the Application:

In the root directory of the source code, run:

```make run```

This command builds the Docker images (if they are not built already) and starts the containers in detached mode.

2. Monitor Logs:

```make logs```

Use this command to view real-time logs for all services. (You may want to include instructions on how to filter logs for a specific service if needed.)

3. Scaling Executor Agents:

To scale the number of executor agents, use:

```make scale-agents num=NUMBER```

Replace NUMBER with the desired number of agent instances.

4. Stopping the Application:

```make down```

This stops and removes all running containers.

5. Resetting the Database:

To stop the application and remove the persistent database data, run:

```make down-prune-db```