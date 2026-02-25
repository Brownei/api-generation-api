YELLOW := \033[0;33m
NC := \033[0m

test: ## Run tests
	@echo "$(YELLOW)Running all tests...$(NC)"
	go test ./tests

run: ## Run docker postgres and start the application
	@echo "$(YELLOW)Starting postgres container...$(NC)"
	(docker stop peppermint-db 2>/dev/null && docker rm peppermint-db) || true
	docker run -d --name peppermint-db \
		-e POSTGRES_USER=db \
		-e POSTGRES_PASSWORD=db \
		-e POSTGRES_DB=db \
		-p 5432:5432 \
		postgres:alpine
	@echo "$(YELLOW)Waiting for postgres to be ready...$(NC)"
	@sleep 5
	@echo "$(YELLOW)Starting application...$(NC)"
	go run main.go

stop: ## Stop and remove the postgres container
	@echo "$(YELLOW)Stopping postgres container...$(NC)"
	(docker stop peppermint-db && docker rm peppermint-db) || true
