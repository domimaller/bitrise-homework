# Task Executor Application

This application consists of two main components that work together to manage and execute tasks. User can create tasks via a restful API where shell commands can be given, those are executed by an executor agent in its container's shell.

## backend-api-server
  An API server where tasks can be created and their statuses retrieved. It exposes endpoints to create new tasks, query their status, and manage task execution.

### Endpoints

#### User Endpoints

- POST /tasks: Create a task with a command.
- GET /tasks: List all created tasks with their states.
- GET /tasks/<resource_id>: Retrieve details of a specific task by its resource ID.

#### Internal Endpoints (for Executor Agents)

These endpoints are intended to be called only by executor agents. In production, access could be restricted using an ingress controller or firewall rules to prevent external access.

- GET /tasks/pick: Allows an executor agent to pick a task for execution. If there are queued tasks available, this endpoint always returns the queued task that was created the earliest.
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
- [Docker](https://docs.docker.com/get-docker)
- [Docker Compose](https://docs.docker.com/compose/install)

## Configuration

The application can be configured via environment variables. Important variables include:

- **SERVER_PORT** (backend-api-server): The port on which the API server listens.
- **LOG_LEVEL:** Set to `info` or `debug` to control the verbosity of the logs.
- **DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME:** PostgreSQL configuration parameters.
- **POLL_INTERVAL** (task-exec-agent): Interval between polling requests for new tasks.

*Other environment variables can be modified if necessary, but these are the essential ones for the default setup.*

## Docker Compose Setup

The application is containerized using Docker / Docker Compose. Below is the provided `docker-compose.yaml` file:

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

**Service Communication:** Services communicate using their service names (e.g. the API server connects to the database using DB_HOST=db).

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

1. Build and start the Application:

In the root directory of the source code, run:

```make run```

This command builds the Docker images (if they are not built already) and starts the containers in detached mode.

2. Monitor Logs:

```make logs```

Use this command to view real-time logs for all services.

1. Scaling Executor Agents:

To scale the number of executor agents, use:

```make scale-agents num=NUMBER```

Replace NUMBER with the desired number of agent instances.

4. Stopping the Application:

```make down```

This stops and removes all running containers.

5. Resetting the Database:

To stop the application and remove the persistent database data, run:

```make down-prune-db```

## Proposed improvements for production:

### New features
Expand user API with DELETE /tasks/<resource_id> for deleting queued but not processed/executed task, or just create another subpath for tasks such as POST /tasks/<resource_id>/abort to cancel a queued task. Second approach might be better as the existing resource could still inform user that the task was cancelled.

### Deployment and scaling

#### Orchestration
Deploy the services on container orchestration platform such Kurbernetes/Openshift or Docker Swarm. This would allow implementing further functionalities such as automated scaling, self-healing, load-balancing and easier rollouts, better handling of life-cycle management. Also, the improved deployment would allow to hide internal APIs from users, which would be a must to keep the application in consistent states avoiding unauthorized access to these APIs.

#### Horizontal scaling
The backend-api-server and task-exec-agent services could be scaled horizontally based on demand using container orchestration solutions such as HPA in Kubernetes. This could also reduce costs of required resources when demand is low, but otherwise serve more users and execute more tasks seamlessly if demand is high.


### Database solution
The current Docker Compose setup with PostgreSQL is sufficient for development and testing but in production more sophisticated solutions might be required. In production, utilizing managed PostgreSQL service (AWS, GCP) would be highly desired, or deploy the DB in highly available clusters georedundantly with replications and automated backups.

Also, if the workload demands it would be wise to evaluate other database technologies that offer higher scalability or specific performance characteristics.

### Monitoring and logging

#### Centralized logging
Integrate a centralized logging solution that can aggregate logs from all services (for instance ELK stack: Fluentd for log collection and Elasticsearch for storing logs).

#### Application performance monitoring
Monitoring tools such as Prometheus could be easily integrated into the services using Prometheus Go library. It would require a Prometheus service running where the predefined counter/gauges or other metrics could be stored from the services. Also, other open-source tools could be used such as Grafana to visualize these metrics and evaluate them.

Healthchecks could be expanded to set up alerting for key metrics such as:
- Task queue lengths and processing times
- Instances where tasks are picked but never finished by an executor

### Further changes and enchancements

#### Error handling
Error handling could be enhanced in both backend service and executor agents. It could involve introducing retries, fallback mechanism or re-queuing tasks if an executions fails.

#### Security
Authentication and authorization could be implemented for API endpoints. On container orchestration platforms it is possible to integrate gateways for ingress controllers where authentication/authorization plugins are fairly easy to set up. Also, it could be considered to use TLS in inter-service communication if packets could reach public space. User API request for creating and reading tasks could potentially come from public space so TLS would be a must for ingress communication and responses and for egress in case backend would send out notifications or other kind of requests in an improved version.

#### Resilience
Evaluate push-based models using message broker like RabbitMQ to further improve task distribution and real-time processing.

#### Configuration management
In current version there are not so many configurations in the services that could be dynamically changed but in future versions it might be desired to utilize a centralized configuration management tool where service configurations could be set. Using the app for specific set of tasks would also require a dedicated config set.

#### CI/CD solutions

**Automated Builds and Testing**: Every commit, especially when a new branch is created, could trigger a pipeline that builds Docker images for both the backend-api-server and task-exec-agent. The pipeline would run unit tests, linting, and static analysis to catch errors early.

**Artifact Management:** After successful tests, the Docker images would be pushed to a container registry for versioning and traceability.

**Staging Deployment (vLAB)**: The pipeline could automatically deploy the application into a dedicated staging environment (e.g., a vLAB) using Docker Compose or Kubernetes. Integration tests and end-to-end tests could run in this environment to validate overall system behavior.

**Production Deployment:** Once the integration tests pass and the build is validated, the solution would promote the build to production.

**Monitoring and Feedback:** Throughout the pipeline, logs, metrics, and health checks could be collected and monitored, ensuring rapid detection and resolution of issues.

### Testing strategies

#### Unit tests
In current implementation the unit tests don't cover most of the code base which is highly desired to do so. More unit test would be needed with requirements such as 90% of each service's code base must be covered by unit tests.

#### Integration tests
Create tests that run the full Docker Compose stack to validate interraction between components. Test scaled agents how they execute tasks in integrated environment. Creating these tests might need utilizing testing frameworks such as Robot. These testcases could be run automatically as well in a CI solution.

#### End-to-End tests
Implement tests simulation user scenarios, such creating tasks and reading them in specific predefined ways.

#### Load and performance testing
Create load testing scenarios to understand how the system behaves under stress and identify bottlenecks.

#### Security and peneration testing
Perform regular security assessments. In this app case it would be highly desired as the commands given in tasks are executed in the agent shell and a malicious user could potentially penetrate the system if services allow specific harmful commands to be executed. Most probably commands would need extra validation as well before processing them.

