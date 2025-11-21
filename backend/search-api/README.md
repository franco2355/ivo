# Search API

Microservicio de búsqueda con sistema de caché de dos niveles y consumidor de eventos RabbitMQ.

## Características

- **Búsqueda Avanzada**: Filtros, paginación, ordenamiento
- **Caché de Dos Niveles**:
  - **CCache** (local, in-memory): TTL 30s
  - **Memcached** (distribuido): TTL 60s
- **Consumidor RabbitMQ**: Indexación automática desde eventos
- **Implementación In-Memory**: Lista para desarrollo, fácil migración a Solr

## Tecnologías

- **Go 1.23**
- **In-Memory Search** (desarrollo) → Migrable a **Apache Solr**
- **Memcached** - Caché distribuido
- **RabbitMQ** - Consumidor de eventos
- **Gin** - Framework web

## Estructura

```
search-api/
├── cmd/api/                    # Punto de entrada
├── internal/
│   ├── config/                # Configuración
│   ├── models/                # SearchDocument, SearchRequest
│   ├── services/              # SearchService, CacheService
│   ├── consumers/             # RabbitMQ consumer
│   ├── handlers/              # API REST
│   └── middleware/            # CORS
├── .env.example
├── Dockerfile
└── go.mod
```

## Flujo de Datos

### Flujo de Búsqueda (con Caché)

```
Request → CCache (local, 30s)
            ↓ MISS
          Memcached (60s)
            ↓ MISS
          SearchService (in-memory/Solr)
            ↓
          Guarda en Memcached + CCache
            ↓
          Response + Header "X-Cache: HIT/MISS"
```

### Flujo de Indexación (desde RabbitMQ)

```
activities-api → Publica evento → RabbitMQ Exchange
subscriptions-api →                    │
                                       ▼
                              search-api consume
                                       │
                    ┌──────────────────┼─────────────────┐
                    │                  │                 │
              1. IndexDocument   2. InvalidateCache  3. ACK message
                    │                  │
                    ▼                  ▼
            SearchService       CacheService
```

## Endpoints

### Búsqueda

```bash
# Búsqueda rápida (query params)
GET /search?q=yoga&type=activity

# Búsqueda avanzada (POST)
POST /search
{
  "query": "yoga",
  "type": "activity",
  "filters": {
    "categoria": "fitness",
    "dia": "Lunes"
  },
  "page": 1,
  "page_size": 10
}

# Obtener documento
GET /search/:id

# Estadísticas del índice
GET /search/stats
```

### Administración

```bash
# Indexar documento manualmente
POST /search/index
{
  "id": "activity_123",
  "type": "activity",
  "titulo": "Yoga Matutino",
  "categoria": "fitness",
  ...
}

# Eliminar documento
DELETE /search/:id

# Health check
GET /healthz
```

## Eventos Consumidos

El microservicio escucha los siguientes eventos de RabbitMQ:

- `activity.create` → Indexa actividad
- `activity.update` → Actualiza actividad
- `activity.delete` → Elimina actividad
- `plan.create` → Indexa plan
- `plan.update` → Actualiza plan
- `subscription.create` → Indexa suscripción
- `subscription.update` → Actualiza suscripción

## Ejecución

```bash
# Local
go mod download
go run cmd/api/main.go

# Docker
docker build -t search-api .
docker run -p 8084:8084 --env-file .env search-api
```

## Migración a Apache Solr

Para migrar de in-memory a Solr en producción:

1. **Instalar Apache Solr** y crear el core `gym_search`
2. **Reemplazar** `internal/services/search_service.go` con cliente Solr
3. **Mantener** el resto de la arquitectura (caché, consumer, handlers)

Ejemplo de integración Solr en Go:
```go
import "github.com/rtt/Go-Solr"

solr := gosolr.NewSolrInterface("http://localhost:8983/solr", "gym_search")
```

## Ventajas de la Arquitectura

- **Desacoplamiento**: Los microservicios no necesitan conocer search-api
- **Escalabilidad**: Caché de dos niveles reduce carga
- **Consistencia Eventual**: RabbitMQ garantiza que los datos se indexen
- **Fácil Migración**: Cambiar de in-memory a Solr sin afectar otros servicios

## Headers de Respuesta

Todas las búsquedas incluyen el header `X-Cache`:
- `X-Cache: HIT` → Dato obtenido del caché
- `X-Cache: MISS` → Dato obtenido de búsqueda directa
