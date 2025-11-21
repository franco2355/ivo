#!/usr/bin/env python3
import requests
import json
from datetime import datetime

print("🔍 Iniciando creación e indexación de planes en Search API...")

# Definir planes de ejemplo
planes_ejemplo = [
    {
        "nombre": "Plan Básico",
        "descripcion": "Acceso a clases básicas de lunes a viernes",
        "precio_mensual": 15000,
        "tipo_acceso": "limitado",
        "duracion_dias": 30,
        "activo": True,
        "actividades_permitidas": ["yoga", "funcional"]
    },
    {
        "nombre": "Plan Premium",
        "descripcion": "Acceso ilimitado a todas las clases y sucursales",
        "precio_mensual": 25000,
        "tipo_acceso": "completo",
        "duracion_dias": 30,
        "activo": True,
        "actividades_permitidas": []
    },
    {
        "nombre": "Plan Estudiante",
        "descripcion": "Plan especial para estudiantes con descuento",
        "precio_mensual": 12000,
        "tipo_acceso": "limitado",
        "duracion_dias": 30,
        "activo": True,
        "actividades_permitidas": ["yoga", "pilates", "funcional"]
    }
]

# Token de admin para crear planes (si es necesario)
# Nota: Ajusta esto según tu configuración de autenticación
ADMIN_TOKEN = None  # Cambiar si se requiere autenticación

headers = {'Content-Type': 'application/json'}
if ADMIN_TOKEN:
    headers['Authorization'] = f'Bearer {ADMIN_TOKEN}'

created_plans = []
indexed = 0
errors = 0

# Crear planes vía API
print("\n📝 Creando planes en subscriptions-api...")
for plan_data in planes_ejemplo:
    try:
        # Crear plan en subscriptions-api
        resp = requests.post(
            'http://localhost:8081/plans',
            json=plan_data,
            headers=headers
        )

        if resp.status_code in [200, 201]:
            plan = resp.json()
            created_plans.append(plan)
            print(f"✅ Plan creado: {plan_data['nombre']} (ID: {plan.get('id', 'N/A')})")
        else:
            print(f"❌ Error creando plan {plan_data['nombre']}: {resp.text}")
            errors += 1
    except Exception as e:
        print(f"❌ Error creando plan {plan_data['nombre']}: {e}")
        errors += 1

# Indexar planes en search-api
print(f"\n🔍 Indexando {len(created_plans)} planes en search-api...")
for plan in created_plans:
    # Construir documento para search-api
    doc = {
        "id": f"plan_{plan.get('id', '')}",
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
            print(f"✅ Plan indexado: {plan['nombre']}")
            indexed += 1
        else:
            print(f"❌ Error indexando plan {plan.get('nombre')}: {resp.text}")
            errors += 1
    except Exception as e:
        print(f"❌ Error indexando plan {plan.get('nombre')}: {e}")
        errors += 1

print(f"\n🎉 Proceso completado!")
print(f"   ✅ Planes creados: {len(created_plans)}")
print(f"   ✅ Planes indexados: {indexed}")
print(f"   ❌ Errores: {errors}")
