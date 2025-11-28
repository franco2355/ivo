// Script para crear planes de ejemplo en MongoDB
// Ejecutar con: docker exec gym-mongo mongosh gym_subscriptions /scripts/create_plans_mongo.js

// La base de datos ya se selecciona al invocar:
// mongosh gym_subscriptions /ruta/al/script.js

// Limpiar planes existentes (opcional)
db.plans.deleteMany({});

// Crear planes de ejemplo
const planes = [
    {
        nombre: "Plan Básico",
        descripcion: "Acceso a clases básicas de lunes a viernes",
        precio_mensual: 15000,
        tipo_acceso: "limitado",
        duracion_dias: 30,
        activo: true,
        actividades_permitidas: ["yoga", "funcional"],
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        nombre: "Plan Premium",
        descripcion: "Acceso ilimitado a todas las clases y sucursales",
        precio_mensual: 25000,
        tipo_acceso: "completo",
        duracion_dias: 30,
        activo: true,
        actividades_permitidas: [],
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        nombre: "Plan Estudiante",
        descripcion: "Plan especial para estudiantes con descuento",
        precio_mensual: 12000,
        tipo_acceso: "limitado",
        duracion_dias: 30,
        activo: true,
        actividades_permitidas: ["yoga", "pilates", "funcional"],
        created_at: new Date(),
        updated_at: new Date()
    }
];

// Insertar planes
const resultado = db.plans.insertMany(planes);

print("✅ Planes creados exitosamente:");
print(`   Total: ${resultado.insertedIds.length}`);

// Mostrar planes creados
db.plans.find().forEach(plan => {
    print(`   - ${plan.nombre} (ID: ${plan._id})`);
});
