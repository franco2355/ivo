#!/usr/bin/env python3
import requests
import subprocess
import json

print("🔍 Obteniendo planes desde MongoDB e indexando en Search API...")

# Primero copiar de 'plans' a 'planes' si es necesario
print("📋 Sincronizando colecciones de planes...")
sync_command = [
    "docker", "exec", "gym-mongo",
    "mongosh", "gym_subscriptions", "--quiet", "--eval",
    """
    const plansFromPlans = db.plans.find().toArray();
    if (plansFromPlans.length > 0) {
        db.planes.deleteMany({});
        db.planes.insertMany(plansFromPlans);
        print('✅ Sincronizados ' + plansFromPlans.length + ' planes');
    }
    """
]

try:
    subprocess.run(sync_command, check=True, capture_output=True)
except Exception as e:
    print(f"⚠️  Advertencia al sincronizar: {e}")

# Obtener planes desde MongoDB (de la colección 'planes' que usa la API)
mongo_command = [
    "docker", "exec", "gym-mongo",
    "mongosh", "gym_subscriptions", "--quiet", "--eval",
    "JSON.stringify(db.planes.find().toArray())"
]

try:
    result = subprocess.run(mongo_command, capture_output=True, text=True, check=True)
    planes_json = result.stdout.strip()
    planes = json.loads(planes_json)
    print(f"📊 Se encontraron {len(planes)} planes en MongoDB")
except Exception as e:
    print(f"❌ Error obteniendo planes de MongoDB: {e}")
    exit(1)

if not planes:
    print("❌ No se encontraron planes")
    exit(1)

# Indexar cada plan
indexed = 0
errors = 0

for plan in planes:
    # Construir documento para search-api
    plan_id = str(plan['_id']['$oid']) if isinstance(plan['_id'], dict) else str(plan['_id'])

    doc = {
        "id": f"plan_{plan_id}",
        "type": "plan",
        "nombre": plan.get('nombre', ''),
        "descripcion": plan.get('descripcion', ''),
        "precio_mensual": plan.get('precio_mensual', 0),
        "tipo_acceso": plan.get('tipo_acceso', ''),
        "duracion_dias": str(plan.get('duracion_dias', 30)),
        "activo": plan.get('activo', True)
    }

    try:
        resp = requests.post(
            'http://localhost:8084/search/index',
            json=doc,
            headers={'Content-Type': 'application/json'}
        )

        if resp.status_code in [200, 201]:
            print(f"✅ Plan indexado: {plan['nombre']} (ID: {plan_id})")
            indexed += 1
        else:
            print(f"❌ Error indexando plan {plan.get('nombre')}: {resp.text}")
            errors += 1
    except Exception as e:
        print(f"❌ Error indexando plan {plan.get('nombre')}: {e}")
        errors += 1

print(f"\n🎉 Indexación completada!")
print(f"   ✅ Planes indexados: {indexed}")
print(f"   ❌ Errores: {errors}")
