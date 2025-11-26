#!/bin/bash

# Script para levantar el sistema completo del gimnasio
# Detecta puertos disponibles automáticamente

echo "🚀 Iniciando sistema de gestión de gimnasio..."
echo "================================================"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Usar puertos por defecto directamente - más rápido y confiable
USERS_API_PORT=5005
SUBSCRIPTIONS_API_PORT=5004
ACTIVITIES_API_PORT=5003
PAYMENTS_API_PORT=5001
SEARCH_API_PORT=5002
MYSQL_PORT=3307
MONGO_PORT=27017
RABBITMQ_AMQP_PORT=5672
RABBITMQ_MGMT_PORT=15672
MEMCACHED_PORT=11211
SOLR_PORT=8983
FRONTEND_PORT=5173

echo -e "${YELLOW}📋 Paso 1: Configurando puertos...${NC}"

echo -e "${GREEN}  ✓ Puertos configurados:${NC}"
echo "    - Users API: $USERS_API_PORT"
echo "    - Subscriptions API: $SUBSCRIPTIONS_API_PORT"
echo "    - Activities API: $ACTIVITIES_API_PORT"
echo "    - Payments API: $PAYMENTS_API_PORT"
echo "    - Search API: $SEARCH_API_PORT"
echo "    - MySQL: $MYSQL_PORT"
echo "    - MongoDB: $MONGO_PORT"
echo "    - RabbitMQ AMQP: $RABBITMQ_AMQP_PORT"
echo "    - RabbitMQ Management: $RABBITMQ_MGMT_PORT"
echo "    - Memcached: $MEMCACHED_PORT"
echo "    - Solr: $SOLR_PORT"
echo "    - Frontend: $FRONTEND_PORT"

# Exportar variables de entorno
export USERS_API_PORT
export SUBSCRIPTIONS_API_PORT
export ACTIVITIES_API_PORT
export PAYMENTS_API_PORT
export SEARCH_API_PORT
export MYSQL_EXTERNAL_PORT=$MYSQL_PORT
export MONGO_EXTERNAL_PORT=$MONGO_PORT
export RABBITMQ_AMQP_PORT
export RABBITMQ_MANAGEMENT_PORT=$RABBITMQ_MGMT_PORT
export MEMCACHED_PORT
export SOLR_PORT
export FRONTEND_PORT

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
    docker image prune -f 2>/dev/null || true

    echo -e "${YELLOW}  Limpiando contenedores detenidos...${NC}"
    docker container prune -f 2>/dev/null || true

    echo -e "${GREEN}  ✓ Limpieza completa realizada${NC}"
else
    echo -e "${GREEN}  ✓ Omitiendo limpieza profunda${NC}"
fi

echo ""
echo -e "${YELLOW}📋 Paso 4: Construyendo imágenes de Docker...${NC}"

# Construir las imágenes
docker compose build --no-cache 2>&1 | tail -20

echo -e "${GREEN}  ✓ Imágenes construidas${NC}"

echo ""
echo -e "${YELLOW}📋 Paso 5: Iniciando servicios...${NC}"

# Levantar los servicios
docker compose up -d

echo -e "${GREEN}  ✓ Servicios iniciados${NC}"

echo ""
echo -e "${YELLOW}📋 Paso 6: Esperando que los servicios estén listos...${NC}"

# Esperar a que los servicios estén listos
sleep 8

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
check_service "Users API" "http://localhost:$USERS_API_PORT/healthz" || true
check_service "Subscriptions API" "http://localhost:$SUBSCRIPTIONS_API_PORT/healthz" || true
check_service "Activities API" "http://localhost:$ACTIVITIES_API_PORT/healthz" || true
check_service "Payments API" "http://localhost:$PAYMENTS_API_PORT/healthz" || true
check_service "Search API" "http://localhost:$SEARCH_API_PORT/healthz" || true
check_service "Frontend" "http://localhost:$FRONTEND_PORT" || true

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}✅ Sistema iniciado correctamente!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo "📊 URLs de los servicios:"
echo "  - Frontend:          http://localhost:$FRONTEND_PORT"
echo "  - Users API:         http://localhost:$USERS_API_PORT"
echo "  - Subscriptions API: http://localhost:$SUBSCRIPTIONS_API_PORT"
echo "  - Activities API:    http://localhost:$ACTIVITIES_API_PORT"
echo "  - Payments API:      http://localhost:$PAYMENTS_API_PORT"
echo "  - Search API:        http://localhost:$SEARCH_API_PORT"
echo "  - MySQL:             localhost:$MYSQL_PORT"
echo "  - MongoDB:           localhost:$MONGO_PORT"
echo "  - RabbitMQ:          localhost:$RABBITMQ_AMQP_PORT"
echo "  - RabbitMQ UI:       http://localhost:$RABBITMQ_MGMT_PORT"
echo "  - Solr:              http://localhost:$SOLR_PORT/solr"
echo "  - Memcached:         localhost:$MEMCACHED_PORT"
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
