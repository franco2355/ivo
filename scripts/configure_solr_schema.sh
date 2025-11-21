#!/bin/bash

# Script para configurar el schema de Solr con los campos necesarios
SOLR_URL="http://localhost:8983/solr"
CORE_NAME="gym_search"

echo "🔧 Configurando schema de Solr para $CORE_NAME..."

# Función para agregar un campo
add_field() {
    local field_name=$1
    local field_type=$2
    local stored=${3:-true}
    local indexed=${4:-true}

    curl -X POST "$SOLR_URL/$CORE_NAME/schema" \
        -H 'Content-type:application/json' \
        -d "{
            \"add-field\": {
                \"name\": \"$field_name\",
                \"type\": \"$field_type\",
                \"stored\": $stored,
                \"indexed\": $indexed
            }
        }" 2>/dev/null

    echo "  ✓ Campo agregado: $field_name ($field_type)"
}

# Agregar campos necesarios
echo ""
echo "📝 Agregando campos..."

add_field "type" "string"
add_field "titulo" "text_general"
add_field "descripcion" "text_general"
add_field "categoria" "string"
add_field "instructor" "string"
add_field "dia" "string"
add_field "horario_inicio" "string"
add_field "horario_final" "string"
add_field "sucursal_id" "string"
add_field "sucursal_nombre" "string"
add_field "requiere_premium" "boolean"
add_field "cupo_disponible" "pint"
add_field "plan_nombre" "string"
add_field "plan_precio" "pdouble"
add_field "plan_tipo_acceso" "string"

echo ""
echo "✅ Schema de Solr configurado correctamente!"
echo ""
echo "🔍 Verificando campos..."
curl -s "$SOLR_URL/$CORE_NAME/schema/fields" | grep -o '"name":"[^"]*"' | head -20
echo ""
