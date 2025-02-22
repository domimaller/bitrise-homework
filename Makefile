.PHONY: run down down-prune-db logs scale-agents

project ?= task-executor-app

run:
	docker-compose -p $(project) up -d --build

down:
	docker-compose -p $(project) down

down-prune-db:
	docker-compose -p $(project) down -v

logs:
	docker-compose -p $(project) logs -f

scale-agents:
	@echo "Scaling task-exec-agent to $(num) instance(s) in project $(project)..."
	docker-compose -p $(project) up -d --scale task-exec-agent=$(num)
