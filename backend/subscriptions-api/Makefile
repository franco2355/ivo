# Makefile para Subscriptions API

.PHONY: run build test clean docker-up docker-down help

# Ejecutar el servicio
run:
	go run cmd/api/main.go

# Compilar el binario
build:
	go build -o bin/subscriptions-api.exe cmd/api/main.go

# Ejecutar tests
test:
	go test ./internal/services/... -v

# Tests con cobertura
test-coverage:
	go test ./internal/services/... -cover

# Limpiar binarios
clean:
	rm -rf bin/

# Levantar dependencias con Docker
docker-up:
	docker run -d --name mongo -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=root -e MONGO_INITDB_ROOT_PASSWORD=admin mongo:latest || echo "MongoDB ya está corriendo"
	docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management || echo "RabbitMQ ya está corriendo"
	@echo "✅ Dependencias levantadas"
	@echo "MongoDB: localhost:27017"
	@echo "RabbitMQ: localhost:5672"
	@echo "RabbitMQ Management: http://localhost:15672 (guest/guest)"

# Detener dependencias
docker-down:
	docker stop mongo rabbitmq || true
	docker rm mongo rabbitmq || true

# Instalar dependencias Go
deps:
	go mod tidy

# Ejecutar script de tests
test-api:
	bash test-api.sh

# Ayuda
help:
	@echo "Comandos disponibles:"
	@echo "  make run            - Ejecutar el servicio"
	@echo "  make build          - Compilar binario"
	@echo "  make test           - Ejecutar tests unitarios"
	@echo "  make test-coverage  - Tests con cobertura"
	@echo "  make docker-up      - Levantar MongoDB y RabbitMQ"
	@echo "  make docker-down    - Detener dependencias"
	@echo "  make deps           - Instalar dependencias Go"
	@echo "  make test-api       - Ejecutar script de tests"
	@echo "  make clean          - Limpiar binarios"
