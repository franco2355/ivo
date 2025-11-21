#!/usr/bin/env python3
import hashlib
import subprocess

# Datos del usuario admin
admin_data = {
    'nombre': 'Admin',
    'apellido': 'Sistema',
    'username': 'admin',
    'email': 'admin@gym.com',
    'password': 'admin123',  # Contraseña en texto plano
    'is_admin': 1
}

# Hashear la contraseña con SHA256 (igual que hace la API)
password_hash = hashlib.sha256(admin_data['password'].encode()).hexdigest()

print("🔐 Creando usuario administrador...")
print(f"   Username: {admin_data['username']}")
print(f"   Email: {admin_data['email']}")
print(f"   Password: {admin_data['password']}")
print(f"   Hash: {password_hash}")
print("")

# Crear usuario en MySQL
sql_command = f"""
INSERT INTO usuarios (nombre, apellido, username, email, password, is_admin, tipo)
VALUES (
    '{admin_data['nombre']}',
    '{admin_data['apellido']}',
    '{admin_data['username']}',
    '{admin_data['email']}',
    '{password_hash}',
    {admin_data['is_admin']},
    'admin'
)
ON DUPLICATE KEY UPDATE
    password = '{password_hash}',
    is_admin = {admin_data['is_admin']},
    tipo = 'admin';
"""

try:
    result = subprocess.run(
        [
            "docker", "exec", "gym-mysql",
            "mysql", "-uroot", "-proot123",
            "-e", f"USE gym_users; {sql_command}"
        ],
        capture_output=True,
        text=True,
        check=True
    )

    print("✅ Usuario administrador creado exitosamente!")
    print("")
    print("📋 Credenciales de acceso:")
    print(f"   Username: {admin_data['username']}")
    print(f"   Email: {admin_data['email']}")
    print(f"   Password: {admin_data['password']}")
    print("")
    print("🌐 Puedes iniciar sesión en: http://localhost:5173/login")

except subprocess.CalledProcessError as e:
    print(f"❌ Error al crear usuario: {e}")
    if e.stderr:
        print(f"   Detalles: {e.stderr}")
