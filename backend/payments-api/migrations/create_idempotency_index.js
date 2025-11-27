// Migración para crear índice único en idempotency_key
// Esto previene que se creen pagos duplicados con el mismo idempotency key
//
// Ejecutar con:
// mongosh gym_management < create_idempotency_index.js
//
// O desde mongosh:
// use gym_management
// load("create_idempotency_index.js")

print("📦 Creando índice único para idempotency_key en colección payments...");

db = db.getSiblingDB('gym_management');

// Crear índice único en idempotency_key
// - unique: true -> No permite duplicados
// - sparse: true -> Solo indexa documentos que tienen el campo (permite documentos sin idempotency_key)
// - name: Nombre descriptivo del índice
db.payments.createIndex(
  { idempotency_key: 1 },
  {
    unique: true,
    sparse: true,
    name: "idx_idempotency_key_unique"
  }
);

print("✅ Índice creado exitosamente!");
print("");
print("Verificación:");
const indexes = db.payments.getIndexes();
printjson(indexes.filter(idx => idx.name === "idx_idempotency_key_unique"));
