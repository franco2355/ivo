#!/bin/bash

# Wrapper para ejecutar la indexación desde el host (fuera de Docker)
# Configura las variables de entorno correctas para acceder a los servicios desde localhost

export SOLR_URL="http://localhost:8983/solr"
export SOLR_CORE="gym_search"
export DB_HOST="127.0.0.1"  # Usar IP para forzar conexión TCP
export DB_PORT="3307"  # Puerto externo de MySQL
export DB_USER="root"
export DB_PASS="root123"
export DB_SCHEMA="gym_activities"

# Ejecutar el script de indexación
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$DIR/index_activities_to_solr.sh"
