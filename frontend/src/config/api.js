// Configuración centralizada de endpoints de API

const API_BASE_URLS = {
  users: 'http://localhost:5005',
  subscriptions: 'http://localhost:5004',
  activities: 'http://localhost:5003',
  payments: 'http://localhost:5001',
  search: 'http://localhost:5002'
};

// Flag para indicar qué APIs están en desarrollo (usarán mock data)
export const USE_MOCK = {
  subscriptions: true,  // API en desarrollo por otro equipo
  search: true,         // API no terminada
  sucursales: true      // Funcionalidad nueva, backend pendiente
};

// Endpoints de Users API
export const USERS_API = {
  base: API_BASE_URLS.users,
  register: `${API_BASE_URLS.users}/register`,
  login: `${API_BASE_URLS.users}/login`,
  users: `${API_BASE_URLS.users}/users`,
  userById: (id) => `${API_BASE_URLS.users}/users/${id}`,
  healthz: `${API_BASE_URLS.users}/healthz`
};

// Endpoints de Subscriptions API (MOCK - Backend en desarrollo)
export const SUBSCRIPTIONS_API = {
  base: API_BASE_URLS.subscriptions,
  plans: `${API_BASE_URLS.subscriptions}/plans`,
  planById: (id) => `${API_BASE_URLS.subscriptions}/plans/${id}`,
  subscriptions: `${API_BASE_URLS.subscriptions}/subscriptions`,
  subscriptionById: (id) => `${API_BASE_URLS.subscriptions}/subscriptions/${id}`,
  activeSubscription: (userId) => `${API_BASE_URLS.subscriptions}/subscriptions/active/${userId}`,
  subscriptionsByUser: (userId) => `${API_BASE_URLS.subscriptions}/subscriptions/user/${userId}`,
  updateStatus: (id) => `${API_BASE_URLS.subscriptions}/subscriptions/${id}/status`
};

// Endpoints de Activities API
export const ACTIVITIES_API = {
  base: API_BASE_URLS.activities,
  actividades: `${API_BASE_URLS.activities}/actividades`,
  actividadById: (id) => `${API_BASE_URLS.activities}/actividades/${id}`,
  inscripciones: `${API_BASE_URLS.activities}/inscripciones`,
  inscripcionById: (id) => `${API_BASE_URLS.activities}/inscripciones/${id}`,
  inscripcionesByUsuario: (id) => `${API_BASE_URLS.activities}/inscripciones/usuario/${id}`,
  sucursales: `${API_BASE_URLS.activities}/sucursales`,
  sucursalById: (id) => `${API_BASE_URLS.activities}/sucursales/${id}`,
  healthz: `${API_BASE_URLS.activities}/healthz`
};

// Endpoints de Payments API (REAL - Operativa)
export const PAYMENTS_API = {
  base: API_BASE_URLS.payments,
  payments: `${API_BASE_URLS.payments}/payments`,
  paymentById: (id) => `${API_BASE_URLS.payments}/payments/${id}`,
  paymentsByUser: (userId) => `${API_BASE_URLS.payments}/payments/user/${userId}`,
  paymentsByEntity: (entityType, entityId) =>
    `${API_BASE_URLS.payments}/payments/entity?entity_type=${entityType}&entity_id=${entityId}`,
  paymentsByStatus: (status) => `${API_BASE_URLS.payments}/payments/status?status=${status}`,
  updateStatus: (id) => `${API_BASE_URLS.payments}/payments/${id}/status`,
  processPayment: (id) => `${API_BASE_URLS.payments}/payments/${id}/process`,
  healthz: `${API_BASE_URLS.payments}/healthz`
};

// Endpoints de Search API (No usar - Backend no terminado)
export const SEARCH_API = {
  base: API_BASE_URLS.search,
  search: `${API_BASE_URLS.search}/search`,
  searchById: (id) => `${API_BASE_URLS.search}/search/${id}`,
  stats: `${API_BASE_URLS.search}/search/stats`,
  healthz: `${API_BASE_URLS.search}/healthz`
};

export default {
  USERS_API,
  SUBSCRIPTIONS_API,
  ACTIVITIES_API,
  PAYMENTS_API,
  SEARCH_API,
  USE_MOCK
};
