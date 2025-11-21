# 🏋️ Sistema de Gestión de Gimnasio - Instrucciones de Uso

## 📋 Tabla de Contenidos
- [Requisitos Previos](#requisitos-previos)
- [Inicio Rápido](#inicio-rápido)
- [Servicios Disponibles](#servicios-disponibles)
- [Usuario Admin](#usuario-admin)
- [Métodos de Pago](#métodos-de-pago)
- [Arquitectura](#arquitectura)
- [Troubleshooting](#troubleshooting)

---

## 🔧 Requisitos Previos

- **Docker** y **Docker Compose** instalados
- **Go 1.23+** (para desarrollo local)
- **Node.js 18+** (para frontend local)
- Puertos disponibles: 3306, 5672, 8080-8084, 5173, 8983, 11211, 15672, 27017

---

## 🚀 Inicio Rápido

### 1. Iniciar el proyecto (RECOMENDADO)

```bash
cd ivo
./start.sh
```

El script automáticamente:
- ✅ Verifica puertos disponibles
- ✅ Limpia contenedores antiguos
- ✅ Construye las imágenes
- ✅ Inicia todos los servicios
- ✅ Espera a que estén saludables
- ✅ Muestra URLs de acceso

### 2. Alternativa: Docker Compose manual

```bash
cd ivo

# Opción A: Configuración mejorada (RECOMENDADO)
docker-compose -f docker-compose.new.yml up -d

# Opción B: Configuración clásica
docker-compose up -d
```

### 3. Verificar que todo esté funcionando

```bash
# Ver logs de todos los servicios
docker-compose -f docker-compose.new.yml logs -f

# Ver logs de un servicio específico
docker-compose -f docker-compose.new.yml logs -f users-api
```

---

## 🌐 Servicios Disponibles

Una vez iniciado, los servicios estarán disponibles en:

| Servicio | URL | Descripción |
|----------|-----|-------------|
| **Frontend** | http://localhost:5173 | Interfaz de usuario React |
| **Users API** | http://localhost:8080 | Autenticación y usuarios |
| **Subscriptions API** | http://localhost:8081 | Planes y suscripciones |
| **Activities API** | http://localhost:8082 | Actividades e inscripciones |
| **Payments API** | http://localhost:8083 | Procesamiento de pagos |
| **Search API** | http://localhost:8084 | Búsqueda indexada |
| **RabbitMQ Management** | http://localhost:15672 | Panel de administración (guest/guest) |
| **Solr Admin** | http://localhost:8983 | Panel de administración Solr |

### Health Checks

Verificar salud de cada servicio:

```bash
curl http://localhost:8080/healthz  # users-api
curl http://localhost:8081/healthz  # subscriptions-api
curl http://localhost:8082/healthz  # activities-api
curl http://localhost:8083/healthz  # payments-api
curl http://localhost:8084/healthz  # search-api
```

---

## 👤 Usuario Admin

El sistema crea automáticamente un usuario administrador:

```
Usuario: admin
Email: admin@gym.com
Contraseña: admin123
```

**Usar este usuario para:**
- ✅ Acceso al panel de administración
- ✅ Crear/editar/eliminar actividades
- ✅ Gestionar planes de suscripción
- ✅ Confirmar pagos en efectivo

---

## 💳 Métodos de Pago

### 1. Mercado Pago (Online)
- Tarjetas de crédito/débito
- Procesamiento automático
- Webhooks configurados

### 2. Efectivo (Manual) ✨ NUEVO
- Pago en sucursal
- Estado: **PENDING** hasta confirmación
- Código de confirmación generado automáticamente

**Flujo de pago en efectivo:**
1. Usuario selecciona "Efectivo" al pagar
2. Sistema genera código único: `CASH-1234567890-userId`
3. Pago queda en estado `PENDING`
4. Usuario se presenta en sucursal con el código
5. Admin confirma el pago manualmente
6. Estado cambia a `COMPLETED`

**Confirmar pago en efectivo (API):**
```bash
curl -X PUT http://localhost:8083/api/payments/{paymentId}/status \
  -H "Authorization: Bearer {admin-token}" \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}'
```

---

## 🏗️ Arquitectura

### Microservicios

```
┌─────────────────────────────────────────────────────┐
│                    Frontend (React)                  │
│                   localhost:5173                     │
└────────────────────┬────────────────────────────────┘
                     │
        ┌────────────┼────────────┬──────────┐
        │            │            │          │
   ┌────▼───┐   ┌───▼────┐  ┌───▼────┐ ┌──▼─────┐
   │ Users  │   │ Subs   │  │Activity│ │Payments│
   │  8080  │   │  8081  │  │  8082  │ │  8083  │
   └────┬───┘   └───┬────┘  └───┬────┘ └────┬───┘
        │           │           │           │
        │      ┌────▼───────────▼───────────▼───┐
        │      │       RabbitMQ (Events)        │
        │      └────────────┬───────────────────┘
        │                   │
   ┌────▼─────────────┐ ┌──▼──────┐
   │  MySQL           │ │ MongoDB │
   │  (users, acts)   │ │ (subs)  │
   └──────────────────┘ └─────────┘
                 │
            ┌────▼────────┐
            │  Search API │
            │    8084     │
            └──────┬──────┘
                   │
         ┌─────────┼─────────┐
    ┌────▼───┐  ┌─▼────┐ ┌──▼────┐
    │  Solr  │  │Memca-│ │RabbitMQ│
    │  8983  │  │ched  │ │Consumer│
    └────────┘  └──────┘ └────────┘
```

### Bases de Datos

| Base de Datos | Servicio | Puerto |
|---------------|----------|--------|
| **MySQL** | users-api, activities-api | 3306 |
| **MongoDB** | subscriptions-api, payments-api | 27017 |
| **Solr** | search-api (indexación) | 8983 |
| **Memcached** | search-api (cache L2) | 11211 |

### Sistema de Eventos (RabbitMQ)

**Exchange:** `gym_events` (tipo: topic)

**Eventos publicados:**

| Routing Key | Servicio | Cuándo |
|-------------|----------|--------|
| `subscription.create` | subscriptions-api | Nueva suscripción |
| `subscription.update` | subscriptions-api | Actualización |
| `subscription.delete` | subscriptions-api | Cancelación |
| `payment.created` | payments-api | Pago iniciado |
| `payment.completed` | payments-api | Pago confirmado |
| `payment.failed` | payments-api | Pago fallido |
| `payment.refunded` | payments-api | Reembolso |
| `activity.create` | activities-api | Nueva actividad |
| `activity.update` | activities-api | Actividad editada |
| `activity.delete` | activities-api | Actividad eliminada |
| `inscription.create` | activities-api | Nueva inscripción |
| `inscription.delete` | activities-api | Desinscripción |

**Consumidores:**
- **search-api**: Indexa automáticamente todos los eventos

---

## 🗄️ Cache

### Niveles de Cache Implementados

#### 1. Search API - Cache de Dos Niveles
- **L1 (In-Memory):** 30 segundos, local a cada instancia
- **L2 (Memcached):** 60 segundos, compartido entre instancias
- **Qué se cachea:** Resultados de búsquedas
- **Invalidación:** Automática al recibir eventos RabbitMQ

#### 2. Subscriptions API - Cache de Planes
- **Tipo:** In-Memory
- **TTL:** 1 hora
- **Qué se cachea:** Lista de planes activos
- **Invalidación:** Al crear/actualizar/eliminar planes
- **Razón:** Los planes cambian muy poco pero se consultan constantemente

#### 3. Activities API - Cache de Actividades
- **Tipo:** In-Memory
- **TTL:** 5 minutos
- **Qué se cachea:** Lista completa de actividades
- **Invalidación:** Al crear/actualizar/eliminar actividades
- **Razón:** Se consultan mucho en la página principal

---

## 🛑 Detener el Proyecto

```bash
cd ivo
./stop.sh

# O manualmente:
docker-compose -f docker-compose.new.yml down

# Para eliminar también los volúmenes (CUIDADO: borra datos):
docker-compose -f docker-compose.new.yml down -v
```

---

## 🔍 Troubleshooting

### Problema: Puerto ya en uso

```bash
# Ver qué proceso usa el puerto
lsof -i :8080  # En Linux/Mac
netstat -ano | findstr :8080  # En Windows

# Matar el proceso
kill -9 <PID>  # Linux/Mac
taskkill /PID <PID> /F  # Windows
```

### Problema: Contenedor no inicia

```bash
# Ver logs del contenedor
docker logs gym-users-api

# Ver logs en tiempo real
docker logs -f gym-users-api

# Inspeccionar contenedor
docker inspect gym-users-api
```

### Problema: Base de datos no se inicializa

```bash
# Recrear volúmenes (CUIDADO: borra datos)
docker-compose -f docker-compose.new.yml down -v
docker-compose -f docker-compose.new.yml up -d

# Verificar que BDD existe
ls -la BDD/

# Ver logs de MySQL
docker logs gym-mysql
```

### Problema: RabbitMQ no conecta

```bash
# Verificar credenciales en .env
# Por defecto: guest/guest

# Verificar que el servicio esté corriendo
docker ps | grep rabbitmq

# Acceder al panel de administración
# http://localhost:15672
# User: guest, Pass: guest
```

### Problema: Frontend no puede conectarse a APIs

- Verificar que las URLs en `frontend/src/config/api.js` sean correctas
- En Docker, usar nombres de servicios: `http://users-api:8080`
- En desarrollo local, usar: `http://localhost:8080`

---

## 📝 Variables de Entorno

Cada microservicio tiene un archivo `.env.example`. Copiar y modificar según necesidad:

```bash
cd backend/users-api
cp .env.example .env
# Editar .env con tus configuraciones
```

**Variables importantes:**

### RabbitMQ
```bash
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=gym_events
```

### MySQL
```bash
DB_USER=root
DB_PASS=root
DB_HOST=localhost
DB_PORT=3306
DB_SCHEMA=proyecto_integrador
```

### MongoDB
```bash
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=gym_subscriptions  # o payments, según el servicio
```

### Mercado Pago
```bash
MERCADOPAGO_ACCESS_TOKEN=your_token
MERCADOPAGO_PUBLIC_KEY=your_public_key
```

---

## 🧪 Testing

### Crear un usuario
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Juan",
    "apellido": "Pérez",
    "username": "juanperez",
    "email": "juan@example.com",
    "password": "password123",
    "tipo": "cliente"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@gym.com",
    "password": "admin123"
  }'
```

### Listar actividades
```bash
curl http://localhost:8082/actividades
```

### Crear pago en efectivo
```bash
curl -X POST http://localhost:8083/api/payments \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "amount": 5000,
    "currency": "ARS",
    "entity_type": "subscription",
    "entity_id": "sub123",
    "payment_gateway": "cash"
  }'
```

---

## 📚 Documentación Adicional

- [Arquitectura de Microservicios](./documentacion/ARQUITECTURA_MICROSERVICIOS.md)
- [Diagrama de Entidades](./documentacion/DIAGRAMA_ENTIDADES.md)
- [Guía Completa](./documentacion/GUIA_COMPLETA_MICROSERVICIOS.md)

---

## 🆘 Soporte

Si encuentras problemas:

1. Revisa los logs: `docker-compose logs -f`
2. Verifica health checks: `curl http://localhost:{port}/healthz`
3. Revisa la documentación en `/documentacion`
4. Asegúrate de que todos los puertos estén disponibles

---

**¡Sistema listo para usar! 🎉**
