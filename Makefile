

export COMPOSE_PROJECT_NAME := go-workers

.PHONY: deps/start
deps/start:
	echo "${COMPOSE_PROJECT_NAME}"
	docker-compose -p '${COMPOSE_PROJECT_NAME}' up -d

.PHONY: deps/stop
deps/stop:
	docker-compose -p '${COMPOSE_PROJECT_NAME}' down

test:
	@make deps/start
	go get github.com/customerio/gospec
	go test -v
	@make deps/stop
