#!/bin/sh
# Entrypoint script para esperar a que RabbitMQ esté listo

set -e

# Extraer host de la URL de RabbitMQ (formato: amqp://user:pass@host:port/)
if [ -n "$RABBITMQ_URL" ]; then
    RABBITMQ_HOST=$(echo "$RABBITMQ_URL" | sed -E 's|amqp://[^@]*@([^:]+):.*|\1|')
    export RABBITMQ_HOST
fi

# Ejecutar script de espera
/wait-for-rabbitmq.sh

echo "Iniciando activities-api..."
exec ./activities-api
