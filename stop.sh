#!/bin/bash

# Script para detener el sistema completo del gimnasio

echo "🛑 Deteniendo sistema de gestión de gimnasio..."
echo "================================================"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}📋 Deteniendo contenedores de Docker...${NC}"

# Detener contenedores
docker compose down

echo -e "${GREEN}✓ Contenedores detenidos${NC}"
echo ""

# Preguntar si quiere eliminar volúmenes
read -p "¿Desea eliminar también los volúmenes (bases de datos)? (s/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[SsYy]$ ]]; then
    echo -e "${YELLOW}📋 Eliminando volúmenes...${NC}"
    docker compose down -v
    echo -e "${GREEN}✓ Volúmenes eliminados${NC}"
fi

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}✅ Sistema detenido correctamente!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
