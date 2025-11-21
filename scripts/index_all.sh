#!/bin/bash

echo "=========================================="
echo "🚀 Indexando todo el contenido en Search API"
echo "=========================================="
echo ""

# Verificar que los contenedores estén corriendo
echo "🔍 Verificando que los servicios estén activos..."
if ! curl -s http://localhost:8082/healthz > /dev/null 2>&1; then
    echo "❌ Error: Activities API no está disponible"
    echo "   Asegúrate de que los contenedores estén corriendo: docker compose up -d"
    exit 1
fi

if ! curl -s http://localhost:8084/search/stats > /dev/null 2>&1; then
    echo "❌ Error: Search API no está disponible"
    echo "   Asegúrate de que los contenedores estén corriendo: docker compose up -d"
    exit 1
fi

echo "✅ Servicios activos"
echo ""

# Indexar actividades
echo "=========================================="
echo "📋 Indexando Actividades"
echo "=========================================="
python3 scripts/index_actividades.py
echo ""

# Indexar planes
echo "=========================================="
echo "💳 Indexando Planes de Suscripción"
echo "=========================================="
python3 scripts/index_planes_from_mongo.py
echo ""

# Verificar indexación
echo "=========================================="
echo "✅ Verificando indexación"
echo "=========================================="

ACTIVIDADES_COUNT=$(curl -s "http://localhost:8084/search?type=activity&page=1&page_size=100" | jq -r '.total_count')
PLANES_COUNT=$(curl -s "http://localhost:8084/search?type=plan&page=1&page_size=100" | jq -r '.total_count')

echo "   📊 Actividades indexadas: $ACTIVIDADES_COUNT"
echo "   💳 Planes indexados: $PLANES_COUNT"
echo ""

echo "=========================================="
echo "✨ ¡Todo indexado correctamente!"
echo "=========================================="
echo ""
echo "Tu aplicación está lista para usar:"
echo "   🌐 Frontend: http://localhost:5173"
echo "   📋 Actividades: http://localhost:5173/actividades"
echo "   💳 Planes: http://localhost:5173/planes"
echo ""
