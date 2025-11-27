#!/bin/sh
# Script para esperar a que RabbitMQ esté completamente disponible

set -e

RABBITMQ_HOST="${RABBITMQ_HOST:-rabbitmq}"
RABBITMQ_PORT="${RABBITMQ_PORT:-5672}"
RABBITMQ_USER="${RABBITMQ_USER:-guest}"
RABBITMQ_PASS="${RABBITMQ_PASS:-guest}"
MAX_RETRIES=30
RETRY_INTERVAL=2

echo "Esperando a que RabbitMQ esté disponible en $RABBITMQ_HOST:$RABBITMQ_PORT..."

# Función para verificar si RabbitMQ está disponible
check_rabbitmq() {
    # Intentar conectarse usando nc (netcat)
    nc -z "$RABBITMQ_HOST" "$RABBITMQ_PORT" 2>/dev/null
    return $?
}

# Esperar a que el puerto esté abierto
retry_count=0
until check_rabbitmq; do
    retry_count=$((retry_count + 1))
    if [ $retry_count -ge $MAX_RETRIES ]; then
        echo "ERROR: RabbitMQ no está disponible después de $MAX_RETRIES intentos"
        exit 1
    fi
    echo "RabbitMQ no está listo todavía (intento $retry_count/$MAX_RETRIES). Esperando ${RETRY_INTERVAL}s..."
    sleep $RETRY_INTERVAL
done

echo "RabbitMQ está disponible en el puerto $RABBITMQ_PORT"

# Esperar un poco más para asegurar que RabbitMQ esté completamente inicializado
echo "Esperando 3 segundos adicionales para asegurar que RabbitMQ esté completamente inicializado..."
sleep 3

echo "RabbitMQ está listo para aceptar conexiones"
