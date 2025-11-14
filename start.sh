#!/bin/bash

# Script para levantar el sistema completo del gimnasio
# Soluciona problemas de puertos y limpia el entorno antes de iniciar

set -e  # Detener en caso de error

echo "🚀 Iniciando sistema de gestión de gimnasio..."
echo "================================================"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Puertos que necesita el sistema
PORTS=(8080 8081 8082 8083 8084 5173 3306 27017)

echo -e "${YELLOW}📋 Paso 1: Verificando y liberando puertos...${NC}"

# Función para matar procesos en un puerto específico
kill_port() {
    local port=$1
    local pids=$(lsof -ti:$port 2>/dev/null)

    if [ ! -z "$pids" ]; then
        echo -e "${YELLOW}  ⚠️  Puerto $port ocupado, liberando...${NC}"
        kill -9 $pids 2>/dev/null || true
        sleep 1
        echo -e "${GREEN}  ✓ Puerto $port liberado${NC}"
    else
        echo -e "${GREEN}  ✓ Puerto $port disponible${NC}"
    fi
}

# Liberar todos los puertos necesarios
for port in "${PORTS[@]}"; do
    kill_port $port
done

echo ""
echo -e "${YELLOW}📋 Paso 2: Deteniendo contenedores anteriores...${NC}"

# Detener y eliminar contenedores existentes
docker compose down 2>/dev/null || true
echo -e "${GREEN}  ✓ Contenedores detenidos${NC}"

echo ""
echo -e "${YELLOW}📋 Paso 3: Limpiando recursos de Docker (opcional)...${NC}"

# Preguntar si quiere hacer limpieza profunda
read -p "¿Desea eliminar volúmenes y hacer limpieza completa? (s/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[SsYy]$ ]]; then
    echo -e "${YELLOW}  Eliminando volúmenes...${NC}"
    docker compose down -v 2>/dev/null || true

    echo -e "${YELLOW}  Limpiando imágenes huérfanas...${NC}"
    docker image prune -f

    echo -e "${YELLOW}  Limpiando contenedores detenidos...${NC}"
    docker container prune -f

    echo -e "${GREEN}  ✓ Limpieza completa realizada${NC}"
else
    echo -e "${GREEN}  ✓ Omitiendo limpieza profunda${NC}"
fi

echo ""
echo -e "${YELLOW}📋 Paso 4: Construyendo imágenes de Docker...${NC}"

# Construir las imágenes
docker compose build --no-cache 2>&1 | while read line; do
    if echo "$line" | grep -q "ERROR"; then
        echo -e "${RED}$line${NC}"
    else
        echo "$line"
    fi
done

echo -e "${GREEN}  ✓ Imágenes construidas${NC}"

echo ""
echo -e "${YELLOW}📋 Paso 5: Iniciando servicios...${NC}"

# Levantar los servicios
docker compose up -d

echo -e "${GREEN}  ✓ Servicios iniciados${NC}"

echo ""
echo -e "${YELLOW}📋 Paso 6: Esperando que los servicios estén listos...${NC}"

# Esperar a que los servicios estén listos
sleep 5

# Función para verificar si un servicio está listo
check_service() {
    local name=$1
    local url=$2
    local max_attempts=30
    local attempt=0

    echo -n "  Verificando $name... "

    while [ $attempt -lt $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC}"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 2
    done

    echo -e "${RED}✗ (timeout)${NC}"
    return 1
}

# Verificar servicios principales
check_service "Users API" "http://localhost:8080/healthz"
check_service "Subscriptions API" "http://localhost:8081/healthz"
check_service "Activities API" "http://localhost:8082/healthz"
check_service "Payments API" "http://localhost:8083/healthz"
check_service "Search API" "http://localhost:8084/healthz"
check_service "Frontend" "http://localhost:5173"

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}✅ Sistema iniciado correctamente!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo "📊 URLs de los servicios:"
echo "  - Frontend:          http://localhost:5173"
echo "  - Users API:         http://localhost:8080"
echo "  - Subscriptions API: http://localhost:8081"
echo "  - Activities API:    http://localhost:8082"
echo "  - Payments API:      http://localhost:8083"
echo "  - Search API:        http://localhost:8084"
echo "  - MySQL:             localhost:3306"
echo "  - MongoDB:           localhost:27017"
echo ""
echo "📝 Comandos útiles:"
echo "  - Ver logs:          docker compose logs -f [servicio]"
echo "  - Detener todo:      docker compose down"
echo "  - Reiniciar:         ./start.sh"
echo ""
echo -e "${YELLOW}💡 Usuario de prueba creado:${NC}"
echo "  Username: test"
echo "  Password: 123"
echo ""
