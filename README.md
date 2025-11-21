# Sistema de Gestión de Gimnasio - Arquitectura de Microservicios

Sistema de gestión de gimnasio implementado con **arquitectura de microservicios** en Go y frontend en React con Tailwind CSS.

## 🚀 Inicio Rápido

### Opción 1: Docker (Recomendado - TODO el sistema)

```bash
# 1. Configurar variables de entorno
cp .env.example .env
# Editar .env si es necesario (por defecto funciona con root123)

# 2. Levantar todo el sistema
docker-compose up -d

# 3. Verificar que todo esté corriendo
docker-compose ps
```

**Servicios disponibles:**
- Frontend: http://localhost:5173
- Users API: http://localhost:8080
- Subscriptions API: http://localhost:8081
- Activities API: http://localhost:8082
- Payments API: http://localhost:8083
- Search API: http://localhost:8084
- RabbitMQ Admin: http://localhost:15672 (guest/guest)
- Solr Admin: http://localhost:8983

### Opción 2: Desarrollo Local (Un microservicio)

```bash
# 1. Levantar solo la infraestructura (bases de datos, colas, etc.)
docker-compose up -d mysql mongo rabbitmq memcached solr

# 2. Configurar variables de entorno del microservicio
cd backend/users-api
cp .env.example .env
# Editar .env con configuración local

# 3. Ejecutar el microservicio
go run cmd/api/main.go  # Puerto 8080
```

### Opción 3: Frontend en desarrollo

```bash
cd frontend

# Instalar dependencias (incluye Tailwind CSS)
npm install

# Ejecutar en modo desarrollo
npm run dev
```

### Verificar Health Checks

```bash
curl http://localhost:8080/healthz  # users-api
curl http://localhost:8081/healthz  # subscriptions-api
curl http://localhost:8082/healthz  # activities-api
curl http://localhost:8083/healthz  # payments-api
curl http://localhost:8084/healthz  # search-api
```

---

## 🏗️ Arquitectura

```
Frontend (React)
     │
     ├─→ users-api (8080)         MySQL
     ├─→ subscriptions-api (8081)  MongoDB + RabbitMQ
     ├─→ activities-api (8082)    MySQL + RabbitMQ
     ├─→ payments-api (8083)      MongoDB
     └─→ search-api (8084)        In-Memory + RabbitMQ + Memcached
```

### Microservicios

| Servicio              | Puerto | Base de Datos | Estado       | Descripción                            |
| --------------------- | ------ | ------------- | ------------ | -------------------------------------- |
| **users-api**         | 8080   | MySQL         | ✅ Funcional | Autenticación, JWT, CRUD usuarios      |
| **subscriptions-api** | 8081   | MongoDB       | ✅ Funcional | Planes y suscripciones + eventos       |
| **activities-api**    | 8082   | MySQL         | ✅ Funcional | Actividades, sucursales, inscripciones |
| **payments-api**      | 8083   | MongoDB       | ✅ Funcional | Pagos genéricos, gateways múltiples    |
| **search-api**        | 8084   | In-Memory     | ✅ Funcional | Búsqueda con caché de 2 niveles        |

---

## 🔐 Configuración de Variables de Entorno

El proyecto usa un sistema centralizado de variables de entorno para máxima seguridad.

### Estructura de archivos .env

```
ivo/
├── .env                    # Variables para Docker Compose (NO en git)
├── .env.example            # Plantilla con valores de ejemplo (SÍ en git)
│
└── backend/
    ├── users-api/
    │   └── .env.example    # Para desarrollo local sin Docker
    ├── subscriptions-api/
    │   └── .env.example
    └── ...
```

### ¿Cuándo se usa cada .env?

**Con Docker (`docker-compose up`):**
- Lee **SOLO** el archivo `.env` de la raíz
- Las variables se pasan a los contenedores via `environment:` en docker-compose.yml
- Base de datos: `DB_HOST=mysql` (nombre del contenedor)

**Desarrollo local (`go run main.go`):**
- Cada microservicio lee su propio `.env` local
- Base de datos: `DB_HOST=localhost` y `DB_PORT=3307`
- Útil para debugging y desarrollo rápido

### Ejemplo: Configurar nuevo entorno

```bash
# 1. Copiar plantilla
cp .env.example .env

# 2. Editar credenciales (si es necesario)
nano .env

# 3. Levantar sistema
docker-compose up -d
```

**Variables importantes:**
- `MYSQL_ROOT_PASSWORD` y `DB_PASS`: Deben coincidir con la BD existente
- `JWT_SECRET`: Cambiarlo en producción
- `RABBITMQ_DEFAULT_PASS`: Credenciales de RabbitMQ

---

## 📁 Estructura del Proyecto

```
ivo/
│
├── .env                         # Variables de entorno (Docker)
├── .env.example                 # Plantilla de variables
├── docker-compose.yml           # Infraestructura completa
│
├── backend/
│   ├── users-api/              # Autenticación y gestión de usuarios
│   ├── subscriptions-api/      # Planes y suscripciones (⭐ Ejemplo)
│   ├── activities-api/         # Actividades e inscripciones
│   ├── payments-api/           # Sistema de pagos con gateways
│   └── search-api/             # Búsqueda y caché
│
├── frontend/                   # Aplicación React + Tailwind CSS
│   ├── src/
│   │   ├── components/        # Componentes React
│   │   ├── pages/             # Páginas principales
│   │   ├── styles/            # CSS (+ Tailwind)
│   │   ├── context/           # Context API
│   │   └── hooks/             # Custom hooks
│   ├── tailwind.config.js     # Configuración de Tailwind
│   ├── postcss.config.cjs     # PostCSS para Tailwind
│   └── package.json           # Dependencias (incluye Tailwind)
│
└── documentacion/              # Documentación del proyecto
    ├── ARQUITECTURA_MICROSERVICIOS.md
    ├── DIAGRAMA_ENTIDADES.md
    ├── GUIA_IMPLEMENTAR_MICROSERVICIO.md
    ├── GUIA_COMPLETA_MICROSERVICIOS.md
    └── INSTRUCCIONES_DOCKER.md
```

---

## 📚 Documentación

### Documentación General

- **[ARQUITECTURA_MICROSERVICIOS.md](documentacion/ARQUITECTURA_MICROSERVICIOS.md)** - Patrones de diseño y decisiones arquitectónicas
- **[DIAGRAMA_ENTIDADES.md](documentacion/DIAGRAMA_ENTIDADES.md)** - Modelo de datos completo con relaciones
- **[GUIA_IMPLEMENTAR_MICROSERVICIO.md](documentacion/GUIA_IMPLEMENTAR_MICROSERVICIO.md)** - Guía para crear nuevos microservicios
- **[GUIA_COMPLETA_MICROSERVICIOS.md](documentacion/GUIA_COMPLETA_MICROSERVICIOS.md)** - Guía de uso del sistema completo
- **[INSTRUCCIONES_DOCKER.md](documentacion/INSTRUCCIONES_DOCKER.md)** - Instrucciones para Docker

### Documentación por Microservicio

Cada microservicio tiene su propio README con detalles específicos:

- [users-api/README.md](users-api/README.md) - API de usuarios y autenticación
- [subscriptions-api/README.md](subscriptions-api/README.md) - ⭐ **Ejemplo de referencia con arquitectura limpia**
- [activities-api/README.md](activities-api/README.md) - API de actividades
- [payments-api/README.md](payments-api/README.md) - API de pagos con gateways
  - [ARQUITECTURA_GATEWAYS_PAGOS.md](payments-api/ARQUITECTURA_GATEWAYS_PAGOS.md) - Arquitectura de gateways
  - [GUIA_IMPLEMENTACION_GATEWAYS.md](payments-api/GUIA_IMPLEMENTACION_GATEWAYS.md) - Guía de implementación
- [search-api/README.md](search-api/README.md) - API de búsqueda

---

## 🎯 Características Destacadas

### Patrones Implementados

- **Arquitectura Limpia** (Clean Architecture)

  - Separación de capas: Domain, Repository, Services, Controllers
  - Dependency Injection manual
  - DTOs separados de Entities

- **Event-Driven Architecture**

  - RabbitMQ para comunicación asíncrona
  - Eventos: subscription.created, inscription.created, etc.

- **Cache-Aside Pattern**

  - Caché de dos niveles (CCache local + Memcached distribuido)
  - TTL configurables

- **Repository Pattern**

  - Abstracción de acceso a datos
  - Interfaces + implementaciones (MongoDB, MySQL)

- **Gateway Pattern** (en payments-api)
  - Integración con múltiples pasarelas de pago
  - Strategy Pattern para intercambiar gateways
  - Factory Pattern para creación de instancias

### Seguridad

- **JWT Authentication** (users-api)
- **Password Hashing** (SHA-256)
- **Validación de Contraseñas Fuertes**
- **CORS Configurado**

### Observabilidad

- **Health Checks** en todos los servicios
- **Logs Estructurados**
- **Headers de Caché** (`X-Cache: HIT/MISS`)

---

## 🔄 Flujos de Datos

### Flujo 1: Crear Suscripción

```
1. Usuario → POST /subscriptions → subscriptions-api
2. subscriptions-api valida usuario con users-api (HTTP)
3. subscriptions-api crea suscripción con estado "pendiente_pago"
4. Publica evento a RabbitMQ: subscription.created
5. search-api consume evento y indexa
```

### Flujo 2: Crear Inscripción

```
1. Usuario → POST /inscripciones → activities-api
2. activities-api valida usuario y suscripción activa
3. activities-api crea inscripción
4. Publica evento a RabbitMQ: inscription.created
5. search-api actualiza cupo disponible
```

### Flujo 3: Búsqueda con Caché

```
1. Usuario → GET /search?q=yoga → search-api
2. Busca en CCache local (30s TTL)
   ├─ HIT → Return + Header "X-Cache: HIT"
   └─ MISS → Busca en Memcached (60s TTL)
       ├─ HIT → Guarda en CCache → Return
       └─ MISS → Ejecuta búsqueda → Guarda en ambos → Return
```

---

## 🛠️ Tecnologías

### Backend

- **Go 1.23** - Todos los microservicios
- **Gin** - Framework web HTTP

### Frontend

- **React 19** - Biblioteca de UI
- **React Router 7** - Navegación SPA
- **Vite 6** - Build tool y dev server
- **Tailwind CSS 3.4** - Framework CSS utility-first
- **Vitest** - Testing framework

### Bases de Datos

- **MySQL 9.3** - users-api, activities-api
- **MongoDB 7.0** - subscriptions-api, payments-api

### Mensajería y Caché

- **RabbitMQ 3.12** - Comunicación asíncrona
- **Memcached 1.6** - Caché distribuido
- **CCache** - Caché local in-memory

### Infraestructura

- **Docker & Docker Compose**
- **Apache Solr 9** (opcional para search-api)

---

## 🎨 Tailwind CSS - Guía de Instalación y Uso

El frontend ya tiene Tailwind CSS configurado. Si necesitas instalarlo en un proyecto nuevo:

### Instalación desde cero

```bash
cd frontend

# 1. Instalar Tailwind CSS y dependencias
npm install -D tailwindcss postcss autoprefixer

# 2. Generar archivos de configuración
npx tailwindcss init -p
```

### Configuración

**tailwind.config.js:**
```javascript
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

**postcss.config.cjs:**
```javascript
module.exports = {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

**src/index.css:**
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

### Uso en componentes

```jsx
// Ejemplo de componente con Tailwind
export default function Button({ children, onClick }) {
  return (
    <button
      onClick={onClick}
      className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
    >
      {children}
    </button>
  );
}
```

### Scripts disponibles

```bash
# Desarrollo con hot-reload
npm run dev

# Build para producción (optimiza Tailwind)
npm run build

# Preview del build
npm run preview
```

**Nota:** En producción, Tailwind automáticamente elimina clases no utilizadas (tree-shaking) para minimizar el CSS.

---

## 📊 Arquitectura Limpia (subscriptions-api)

**subscriptions-api es el ejemplo de referencia** que implementa correctamente todos los patrones:

```
subscriptions-api/
├── cmd/api/main.go                    # ✅ DI manual completa
├── internal/
│   ├── domain/
│   │   ├── entities/                  # ✅ Entidades de BD
│   │   └── dtos/                      # ✅ DTOs Request/Response
│   ├── repository/                    # ✅ Interfaces + MongoDB
│   ├── services/                      # ✅ Lógica de negocio con DI
│   ├── infrastructure/                # ✅ Servicios externos
│   ├── controllers/                   # ✅ Capa HTTP
│   ├── middleware/
│   ├── database/
│   └── config/
```

**Ver [subscriptions-api/README.md](subscriptions-api/README.md) para detalles completos.**

---

## 🧪 Testing Rápido

### Registrar Usuario

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Juan",
    "apellido": "Pérez",
    "username": "juanp",
    "email": "juan@example.com",
    "password": "Password123"
  }'
```

### Crear Plan

```bash
curl -X POST http://localhost:8081/plans \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Plan Premium",
    "descripcion": "Acceso completo",
    "precio_mensual": 100.00,
    "tipo_acceso": "completo",
    "duracion_dias": 30,
    "activo": true
  }'
```

### Buscar Actividades

```bash
curl "http://localhost:8084/search?q=yoga&type=activity"
```

---

## 🚧 Próximos Pasos

### Corto Plazo

- [ ] Implementar frontend completo (React)
- [ ] Agregar tests unitarios y de integración
- [ ] Migrar search-api a Apache Solr
- [ ] Implementar métricas (Prometheus + Grafana)

### Mediano Plazo

- [ ] API Gateway (Kong/Traefik)
- [ ] Service Discovery (Consul)
- [ ] Distributed Tracing (Jaeger)
- [ ] Autenticación OAuth2

### Largo Plazo

- [ ] Migrar a Kubernetes
- [ ] CI/CD completo (GitHub Actions)
- [ ] Monitoreo avanzado (ELK Stack)

---

## 🆘 Soporte

Para preguntas o problemas:

1. Revisar la documentación del microservicio específico
2. Consultar [ARQUITECTURA_MICROSERVICIOS.md](documentacion/ARQUITECTURA_MICROSERVICIOS.md)
3. Verificar logs: `docker-compose logs <servicio>`

---

## 👥 Equipo

Proyecto desarrollado como parte de **Arquitectura de Software II** - Universidad Católica de Córdoba

---

## 📄 Licencia

Proyecto académico - Universidad Católica de Córdoba

---

## 🔧 Comandos Útiles

### Docker

```bash
# Ver logs de un servicio
docker-compose logs -f users-api

# Reiniciar un servicio
docker-compose restart users-api

# Detener todo
docker-compose down

# Detener y eliminar volúmenes (BORRA DATOS)
docker-compose down -v

# Reconstruir imágenes
docker-compose up -d --build
```

### Frontend

```bash
# Instalar dependencias
npm install

# Desarrollo
npm run dev

# Tests
npm run test
npm run test:ui
npm run test:coverage

# Linting
npm run lint

# Build producción
npm run build
```

### Base de datos

```bash
# Conectar a MySQL del contenedor
mysql -h 127.0.0.1 -P 3307 -u root -proot123

# Conectar a MongoDB del contenedor
docker exec -it gym-mongo mongosh
```

---

**Última actualización**: 2025-01-20
