#!/bin/bash

# Script para indexar actividades desde MySQL a Solr
# Uso: ./index_activities_to_solr.sh

set -e  # Salir si hay algún error

# Configuración
SOLR_URL="${SOLR_URL:-http://localhost:8983/solr}"
CORE_NAME="${SOLR_CORE:-gym_search}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASS="${DB_PASS:-rootpassword}"
DB_SCHEMA="${DB_SCHEMA:-gym_activities}"

echo "=============================================="
echo "  INDEXACIÓN DE ACTIVIDADES EN SOLR"
echo "=============================================="
echo ""
echo "Configuración:"
echo "  Solr URL: $SOLR_URL/$CORE_NAME"
echo "  MySQL: $DB_USER@$DB_HOST:$DB_PORT/$DB_SCHEMA"
echo ""

# Verificar que Solr esté disponible
echo "🔍 Verificando conexión con Solr..."
if ! curl -s "$SOLR_URL/$CORE_NAME/admin/ping" > /dev/null 2>&1; then
    echo "❌ Error: No se puede conectar con Solr en $SOLR_URL/$CORE_NAME"
    echo "   Asegúrate de que Solr esté corriendo y que el core '$CORE_NAME' exista"
    exit 1
fi
echo "✅ Conexión con Solr exitosa"
echo ""

# Verificar que MySQL esté disponible
echo "🔍 Verificando conexión con MySQL..."
if ! mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -e "USE $DB_SCHEMA" 2>/dev/null; then
    echo "❌ Error: No se puede conectar con MySQL"
    echo "   Verifica las credenciales y que la base de datos exista"
    exit 1
fi
echo "✅ Conexión con MySQL exitosa"
echo ""

# Obtener actividades desde MySQL y crear documentos JSON para Solr
echo "📊 Obteniendo actividades desde MySQL..."

# Query SQL para obtener actividades con información de sucursal
QUERY="
SELECT
    CONCAT('activity_', a.id_actividad) as id,
    'actividad' as type,
    a.titulo,
    COALESCE(a.descripcion, '') as descripcion,
    COALESCE(a.categoria, '') as categoria,
    COALESCE(a.instructor, '') as instructor,
    a.dia,
    TIME_FORMAT(TIME(a.horario_inicio), '%H:%i') as horario_inicio,
    TIME_FORMAT(TIME(a.horario_final), '%H:%i') as horario_final,
    COALESCE(a.sucursal_id, 0) as sucursal_id,
    COALESCE(s.nombre, '') as sucursal_nombre,
    false as requiere_premium,
    (a.cupo - COALESCE(
        (SELECT COUNT(*)
         FROM inscripciones i
         WHERE i.actividad_id = a.id_actividad
           AND i.is_activa = TRUE
        ), 0)
    ) as cupo_disponible
FROM actividades a
LEFT JOIN sucursales s ON a.sucursal_id = s.id_sucursal
WHERE a.activa = TRUE
"

# Ejecutar query y generar JSON
TEMP_FILE=$(mktemp)
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -D"$DB_SCHEMA" -N -B -e "$QUERY" | \
awk -F'\t' 'BEGIN {print "["}
{
    if (NR > 1) print ",";
    printf "  {\n";
    printf "    \"id\": \"%s\",\n", $1;
    printf "    \"type\": [\"%s\"],\n", $2;
    printf "    \"titulo\": [\"%s\"],\n", $3;
    printf "    \"descripcion\": [\"%s\"],\n", $4;
    printf "    \"categoria\": [\"%s\"],\n", $5;
    printf "    \"instructor\": [\"%s\"],\n", $6;
    printf "    \"dia\": [\"%s\"],\n", $7;
    printf "    \"horario_inicio\": [\"%s\"],\n", $8;
    printf "    \"horario_final\": [\"%s\"],\n", $9;
    printf "    \"sucursal_id\": [%s],\n", $10;
    printf "    \"sucursal_nombre\": [\"%s\"],\n", $11;
    printf "    \"requiere_premium\": [%s],\n", ($12 == "1" ? "true" : "false");
    printf "    \"cupo_disponible\": [%s]\n", $13;
    printf "  }";
}
END {print "\n]"}' > "$TEMP_FILE"

# Contar actividades encontradas
ACTIVITY_COUNT=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -D"$DB_SCHEMA" -N -B -e "SELECT COUNT(*) FROM actividades WHERE activa = TRUE")

if [ "$ACTIVITY_COUNT" -eq 0 ]; then
    echo "⚠️  No se encontraron actividades para indexar"
    rm "$TEMP_FILE"
    exit 0
fi

echo "✅ Se encontraron $ACTIVITY_COUNT actividad(es)"
echo ""

# Mostrar muestra del JSON generado
echo "📄 Muestra del JSON generado:"
head -20 "$TEMP_FILE"
echo "..."
echo ""

# Limpiar documentos antiguos de actividades
echo "🗑️  Eliminando actividades antiguas de Solr..."
curl -s -X POST "$SOLR_URL/$CORE_NAME/update?commit=true" \
    -H 'Content-Type: application/json' \
    -d '{"delete": {"query": "type:actividad"}}' > /dev/null

if [ $? -eq 0 ]; then
    echo "✅ Actividades antiguas eliminadas"
else
    echo "⚠️  Error al eliminar actividades antiguas (continuando...)"
fi
echo ""

# Indexar actividades en Solr
echo "📤 Indexando actividades en Solr..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$SOLR_URL/$CORE_NAME/update?commit=true" \
    -H 'Content-Type: application/json' \
    --data-binary "@$TEMP_FILE")

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
RESPONSE_BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "✅ Actividades indexadas correctamente en Solr"
    echo ""

    # Verificar indexación
    echo "🔍 Verificando indexación..."
    COUNT=$(curl -s "$SOLR_URL/$CORE_NAME/select?q=type:actividad&rows=0&wt=json" | grep -o '"numFound":[0-9]*' | grep -o '[0-9]*')
    echo "✅ Total de actividades en Solr: $COUNT"

    if [ "$COUNT" -ne "$ACTIVITY_COUNT" ]; then
        echo "⚠️  Advertencia: El número de documentos indexados ($COUNT) no coincide con el esperado ($ACTIVITY_COUNT)"
    fi
else
    echo "❌ Error al indexar actividades en Solr (HTTP $HTTP_CODE)"
    echo "Respuesta:"
    echo "$RESPONSE_BODY"
    rm "$TEMP_FILE"
    exit 1
fi

# Limpiar archivo temporal
rm "$TEMP_FILE"

echo ""
echo "=============================================="
echo "  ✅ INDEXACIÓN COMPLETADA EXITOSAMENTE"
echo "=============================================="
echo ""
echo "Puedes verificar las actividades indexadas en:"
echo "  $SOLR_URL/$CORE_NAME/select?q=type:actividad"
echo ""
