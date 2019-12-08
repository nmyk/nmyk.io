COMPOSE_FILE=deployment/docker-compose.yml

.PHONY: init
init:
	./scripts/init $(COMPOSE_FILE)

.PHONY: deploy
deploy:
	./scripts/deploy $(COMPOSE_FILE)

.PHONY: develop
develop:
	docker-compose -f $(COMPOSE_FILE) \
		run --rm -d \
		--volume $(shell pwd):/app \
		--publish 8080:8080 \
		--name app-dev \
		--entrypoint /bin/sh app \
		-c 'while sleep 3600; do :; done'
	docker exec -it app-dev /bin/sh

.PHONY: stop
stop:
	docker-compose -f $(COMPOSE_FILE) down --remove-orphans
