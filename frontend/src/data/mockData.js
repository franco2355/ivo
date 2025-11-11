// Mock data para desarrollo hasta que las APIs estén listas

// PLANES DE SUSCRIPCIÓN (Mock - subscriptions-api en desarrollo)
export const mockPlanes = [
  {
    id: "plan_001",
    nombre: "Plan Básico",
    descripcion: "Acceso limitado a actividades grupales",
    precio_mensual: 50.00,
    tipo_acceso: "limitado",
    duracion_dias: 30,
    activo: true,
    actividades_permitidas: ["Yoga", "Pilates", "Stretching"],
    beneficios: [
      "Acceso a actividades básicas",
      "3 clases por semana",
      "Vestuarios y duchas",
      "App móvil"
    ],
    color: "#4CAF50"
  },
  {
    id: "plan_002",
    nombre: "Plan Premium",
    descripcion: "Acceso completo a todas las actividades",
    precio_mensual: 100.00,
    tipo_acceso: "completo",
    duracion_dias: 30,
    activo: true,
    actividades_permitidas: [],
    beneficios: [
      "Acceso ilimitado a todas las actividades",
      "Clases personalizadas",
      "Nutricionista incluido",
      "Vestuarios premium",
      "Descuentos en productos",
      "Invitados gratuitos (2 por mes)"
    ],
    color: "#FF9800",
    popular: true
  },
  {
    id: "plan_003",
    nombre: "Plan Trimestral",
    descripcion: "Plan Premium por 3 meses con descuento",
    precio_mensual: 85.00,
    tipo_acceso: "completo",
    duracion_dias: 90,
    activo: true,
    actividades_permitidas: [],
    beneficios: [
      "Todo lo del Plan Premium",
      "15% de descuento",
      "Evaluación física gratuita",
      "Plan de entrenamiento personalizado"
    ],
    color: "#2196F3",
    ahorro: "15% OFF"
  },
  {
    id: "plan_004",
    nombre: "Plan Anual",
    descripcion: "Plan Premium por 12 meses - Mejor precio",
    precio_mensual: 70.00,
    tipo_acceso: "completo",
    duracion_dias: 365,
    activo: true,
    actividades_permitidas: [],
    beneficios: [
      "Todo lo del Plan Premium",
      "30% de descuento",
      "Sesiones con entrenador personal (4 por mes)",
      "Chequeo médico anual incluido",
      "Prioridad en reservas"
    ],
    color: "#9C27B0",
    ahorro: "30% OFF"
  }
];

// SUSCRIPCIONES DE USUARIOS (Mock)
export const mockSuscripciones = {
  // Usuario ID 5 (ejemplo)
  "5": {
    id: "sub_001",
    usuario_id: "5",
    plan_id: "plan_002",
    plan: mockPlanes[1], // Plan Premium
    sucursal_origen_id: "suc_001",
    fecha_inicio: "2025-01-01T00:00:00Z",
    fecha_vencimiento: "2025-02-01T00:00:00Z",
    estado: "activa",
    pago_id: "pay_001",
    metadata: {
      auto_renovacion: true,
      metodo_pago_preferido: "credit_card",
      notas: "Cliente premium desde enero 2025"
    },
    historial_renovaciones: [
      {
        fecha: "2025-01-01T00:00:00Z",
        pago_id: "pay_001",
        monto: 100.00
      }
    ],
    created_at: "2025-01-01T00:00:00Z",
    updated_at: "2025-01-01T00:00:00Z"
  }
};

// SUCURSALES (Mock - activities-api puede tenerlas pero las mostramos en frontend)
export const mockSucursales = [
  {
    id: "suc_001",
    nombre: "GymPro Centro",
    direccion: "Av. Colón 450, Centro, Córdoba",
    telefono: "+54 351 123-4567",
    email: "centro@gympro.com",
    horarios: "Lunes a Viernes: 6:00 - 23:00 | Sábados: 8:00 - 20:00 | Domingos: 9:00 - 14:00",
    servicios: ["Musculación", "Cardio", "Clases Grupales", "Spinning", "Vestuarios Premium", "Estacionamiento"],
    imagen: "https://images.unsplash.com/photo-1534438327276-14e5300c3a48?w=800",
    destacada: true
  },
  {
    id: "suc_002",
    nombre: "GymPro Nueva Córdoba",
    direccion: "Bv. Illia 580, Nueva Córdoba, Córdoba",
    telefono: "+54 351 234-5678",
    email: "nuevacordoba@gympro.com",
    horarios: "Lunes a Viernes: 7:00 - 22:00 | Sábados: 9:00 - 19:00 | Domingos: 10:00 - 14:00",
    servicios: ["Musculación", "Cardio", "Yoga", "Pilates", "Functional", "Vestuarios", "Wifi"],
    imagen: "https://images.unsplash.com/photo-1571902943202-507ec2618e8f?w=800"
  },
  {
    id: "suc_003",
    nombre: "GymPro Cerro",
    direccion: "Av. Rafael Núñez 4500, Cerro de las Rosas, Córdoba",
    telefono: "+54 351 345-6789",
    email: "cerro@gympro.com",
    horarios: "Lunes a Viernes: 6:00 - 23:00 | Sábados: 8:00 - 20:00 | Domingos: Cerrado",
    servicios: ["Musculación", "Cardio", "CrossFit", "Boxing", "Piscina", "Sauna", "Estacionamiento"],
    imagen: "https://images.unsplash.com/photo-1540497077202-7c8a3999166f?w=800"
  }
];

// Funciones helper para obtener mock data

export const getMockPlanById = (planId) => {
  return mockPlanes.find(plan => plan.id === planId);
};

export const getMockSuscripcionByUserId = (userId) => {
  return mockSuscripciones[userId] || null;
};

export const getMockSucursalById = (sucursalId) => {
  return mockSucursales.find(suc => suc.id === sucursalId);
};

// Función para simular delay de API
export const mockApiDelay = (ms = 500) => {
  return new Promise(resolve => setTimeout(resolve, ms));
};

// Función para crear una suscripción mock
export const createMockSuscripcion = async (userId, planId) => {
  await mockApiDelay(800);

  const plan = getMockPlanById(planId);
  if (!plan) {
    throw new Error("Plan no encontrado");
  }

  const now = new Date();
  const vencimiento = new Date();
  vencimiento.setDate(vencimiento.getDate() + plan.duracion_dias);

  const newSub = {
    id: `sub_${Date.now()}`,
    usuario_id: userId,
    plan_id: planId,
    plan: plan,
    fecha_inicio: now.toISOString(),
    fecha_vencimiento: vencimiento.toISOString(),
    estado: "pendiente_pago",
    metadata: {
      auto_renovacion: false,
      metodo_pago_preferido: "credit_card"
    },
    historial_renovaciones: [],
    created_at: now.toISOString(),
    updated_at: now.toISOString()
  };

  // Guardar en el mock storage
  mockSuscripciones[userId] = newSub;

  return newSub;
};

export default {
  mockPlanes,
  mockSuscripciones,
  mockSucursales,
  getMockPlanById,
  getMockSuscripcionByUserId,
  getMockSucursalById,
  mockApiDelay,
  createMockSuscripcion
};
