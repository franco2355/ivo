#!/usr/bin/env python3
import requests
import json

print("🔍 Iniciando indexación de actividades en Search API...")

# Obtener actividades desde Activities API
try:
    response = requests.get('http://localhost:8082/actividades')
    response.raise_for_status()
    actividades = response.json()
except Exception as e:
    print(f"❌ Error obteniendo actividades: {e}")
    exit(1)

if not actividades:
    print("❌ No se encontraron actividades para indexar")
    exit(1)

print(f"📊 Se encontraron {len(actividades)} actividades")

# Indexar cada actividad
indexed = 0
errors = 0

for act in actividades:
    sucursal_id = act.get('sucursal_id')
    doc = {
        "id": f"activity_{act['id']}",
        "type": "activity",
        "titulo": act.get('titulo', ''),
        "descripcion": act.get('descripcion', ''),
        "categoria": act.get('categoria', ''),
        "instructor": act.get('instructor', ''),
        "dia": act.get('dia', ''),
        "horario_inicio": act.get('horario_inicio', ''),
        "horario_final": act.get('horario_final', ''),
        "cupo_disponible": act.get('lugares', 0),
        "sucursal_id": str(sucursal_id) if sucursal_id is not None else ""
    }

    try:
        resp = requests.post(
            'http://localhost:8084/search/index',
            json=doc,
            headers={'Content-Type': 'application/json'}
        )

        if resp.status_code in [200, 201]:
            print(f"✅ Actividad indexada: {act['titulo']} (ID: {act['id']})")
            indexed += 1
        else:
            print(f"❌ Error indexando actividad {act['id']}: {resp.text}")
            errors += 1
    except Exception as e:
        print(f"❌ Error indexando actividad {act['id']}: {e}")
        errors += 1

print(f"\n🎉 Indexación completada!")
print(f"   ✅ Indexadas: {indexed}")
print(f"   ❌ Errores: {errors}")
