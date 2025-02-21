.PHONY: run down down-prune-db logs scale-agents

run:
	docker-compose up -d --build

down:
	docker-compose down

down-prune-db:
	docker-compose down -v

logs:
	docker-compose logs -f

scale-agents:
	@echo "Scaling task-exec-agent to $(num) instance(s)..."
	docker-compose up -d --scale task-exec-agent=$(num)