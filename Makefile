up:
	docker compose -f deployments/docker-compose.yml up -d

down:
	docker compose -f deployments/docker-compose.yml down

logs:
	docker compose -f deployments/docker-compose.yml logs -f

ps:
	docker compose -f deployments/docker-compose.yml ps

run:
	go run ./cmd/app