COMPOSE_FILE=deployment/docker-compose.yml

.PHONY: init
init:
	./scripts/init.sh $(COMPOSE_FILE)

.PHONY: deploy
deploy:
	./scripts/deploy.py $(COMPOSE_FILE)

.PHONY: develop
develop:
	docker-compose -f $(COMPOSE_FILE) build app
	docker-compose -f $(COMPOSE_FILE) \
		run -d \
		--volume $(shell pwd):/app \
		--publish 8080:8080 \
		--name app-dev \
		--entrypoint /bin/sh app \
		-c 'while sleep 3600; do :; done'
	trap 'make stop' EXIT; docker exec -it app-dev /bin/sh

.PHONY: run-prod
run-prod:
	docker-compose -f $(COMPOSE_FILE) up -d --build

.PHONY: stop
stop:
	docker-compose -f $(COMPOSE_FILE) down --remove-orphans
