.PHONY: up
up:
	docker-compose up -d

.PHONY: down
down:
	docker-compose down -v

.PHONY: test
test:
	./tests/postman/run-tests.sh -e local
