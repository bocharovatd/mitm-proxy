DOCKER_COMPOSE_PATH=docker/docker-compose.yaml

.PHONY: docker-build

docker-build:
	@docker compose -f $(DOCKER_COMPOSE_PATH) build

.PHONY: docker-start

docker-start:
	@docker compose -f $(DOCKER_COMPOSE_PATH) up -d

.PHONY: docker-stop

docker-stop:
	@docker compose -f $(DOCKER_COMPOSE_PATH) stop

.PHONY: docker-clean

docker-clean:
	@docker compose -f $(DOCKER_COMPOSE_PATH) down