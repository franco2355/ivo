#!/bin/bash

# Script para indexar todas las actividades existentes en Search API
# Este script debe ejecutarse después de inicializar la base de datos

echo "🔍 Iniciando indexación de actividades en Search API..."

# Obtener actividades desde Activities API
ACTIVITIES=$(curl -s http://localhost:8082/actividades)

# Verificar si hay actividades
if [ -z "$ACTIVITIES" ] || [ "$ACTIVITIES" == "[]" ]; then
    echo "❌ No se encontraron actividades para indexar"
    exit 1
fi

echo "📊 Actividades obtenidas desde Activities API"

# Parsear el JSON y crear documentos para cada actividad
echo "$ACTIVITIES" | jq -c '.[]' | while read -r actividad; do
    # Extraer datos de la actividad
    ID=$(echo "$actividad" | jq -r '.id')
    TITULO=$(echo "$actividad" | jq -r '.titulo')
    DESCRIPCION=$(echo "$actividad" | jq -r '.descripcion')
    CATEGORIA=$(echo "$actividad" | jq -r '.categoria')
    INSTRUCTOR=$(echo "$actividad" | jq -r '.instructor')
    DIA=$(echo "$actividad" | jq -r '.dia')
    HORARIO_INICIO=$(echo "$actividad" | jq -r '.horario_inicio')
    HORARIO_FINAL=$(echo "$actividad" | jq -r '.horario_final')
    CUPO=$(echo "$actividad" | jq -r '.cupo')
    LUGARES=$(echo "$actividad" | jq -r '.lugares')
    SUCURSAL_ID=$(echo "$actividad" | jq -r '.sucursal_id')

    # Crear documento para Search API
    DOC=$(cat <<EOF
{
  "id": "activity_${ID}",
  "type": "activity",
  "titulo": "${TITULO}",
  "descripcion": "${DESCRIPCION}",
  "categoria": "${CATEGORIA}",
  "instructor": "${INSTRUCTOR}",
  "dia": "${DIA}",
  "horario_inicio": "${HORARIO_INICIO}",
  "horario_final": "${HORARIO_FINAL}",
  "cupo_disponible": ${LUGARES},
  "sucursal_id": ${SUCURSAL_ID}
}
EOF
)

    # Indexar en Search API
    RESPONSE=$(curl -s -X POST http://localhost:8084/search/index \
        -H "Content-Type: application/json" \
        -d "$DOC")

    if echo "$RESPONSE" | grep -q "Documento indexado correctamente"; then
        echo "✅ Actividad indexada: $TITULO (ID: $ID)"
    else
        echo "❌ Error indexando actividad $ID: $RESPONSE"
    fi
done

echo "🎉 Indexación completada!"
