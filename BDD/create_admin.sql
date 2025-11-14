-- Script para crear usuario administrador
-- Usuario: admin
-- Password: admin123

USE gimnasio;

-- Insertar usuario administrador
INSERT INTO usuarios (nombre, apellido, username, password, is_admin)
VALUES ('Admin', 'Sistema', 'admin', SHA2('admin123', 256), 1)
ON DUPLICATE KEY UPDATE
  password = SHA2('admin123', 256),
  is_admin = 1;

-- Verificar que se creó correctamente
SELECT id_usuario, nombre, apellido, username, is_admin, fecha_registro
FROM usuarios
WHERE username = 'admin';
