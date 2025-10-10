COMPOSE_FILE = ./docker-compose.yml
ENV_FILE = .env

.PHONY: all
all: run

# Запуск приложения через Docker Compose
.PHONY: run
run:
	@docker-compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up --build

# Запуск в фоновом режиме
.PHONY: up
up:
	@docker-compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d --build

# Остановка контейнеров
.PHONY: stop
stop:
	@docker-compose -f $(COMPOSE_FILE) down

# Перезапуск контейнеров
.PHONY: restart
restart:
	@docker-compose -f $(COMPOSE_FILE) restart

# Просмотр логов
.PHONY: logs
logs:
	@docker-compose -f $(COMPOSE_FILE) logs -f

# Показать статус контейнеров
.PHONY: ps
ps:
	@docker-compose -f $(COMPOSE_FILE) ps

# Остановка и очистка
.PHONY: down
down:
	@docker-compose -f $(COMPOSE_FILE) down

# Полная очистка (контейнеры, volumes)
.PHONY: clean
clean:
	@docker-compose -f $(COMPOSE_FILE) down -v --remove-orphans