COMPOSE_FILE=deployment/docker-compose.yml

.PHONY: init
init:
	./scripts/init.sh $(COMPOSE_FILE)

.PHONY: deploy
deploy:
	./scripts/deploy.py

.PHONY: develop
develop: export APP_ENVIRONMENT=dev
develop:
	docker-compose -f $(COMPOSE_FILE) build app
	docker-compose -f $(COMPOSE_FILE) \
		run -d \
		--volume $(shell pwd)/pkg:/app/pkg \
		--volume $(shell pwd)/web:/app/web \
		--publish 7070:7070 \
		--publish 8080:8080 \
		--publish 8081:8081 \
		--name app-dev \
		--entrypoint /bin/sh app \
		-c 'while sleep 3600; do :; done'
	trap 'make stop' EXIT; docker exec -it app-dev /bin/sh

.PHONY: run-prod
run-prod: export APP_ENVIRONMENT=prod
run-prod:
	docker-compose -f $(COMPOSE_FILE) up -d --build

.PHONY: stop
stop: export APP_ENVIRONMENT=prod
stop:
	docker-compose -f $(COMPOSE_FILE) down --remove-orphans
