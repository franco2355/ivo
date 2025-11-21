-- Script para insertar datos de prueba en activities-api
-- Ejecutar este script para tener datos con los que probar

USE proyecto_integrador;

-- Insertar sucursales de prueba
INSERT INTO sucursales (nombre, direccion, telefono, created_at, updated_at) VALUES
('Sucursal Centro', 'Av. Colón 1234, Córdoba', '351-123-4567', NOW(), NOW()),
('Sucursal Nueva Córdoba', 'Av. Hipólito Yrigoyen 567, Córdoba', '351-234-5678', NOW(), NOW()),
('Sucursal Alta Córdoba', 'Av. Monseñor Pablo Cabrera 890, Córdoba', '351-345-6789', NOW(), NOW());

-- Insertar actividades de prueba
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, foto_url, instructor, categoria, sucursal_id, created_at, updated_at) VALUES
('Yoga Matutino', 'Clase de yoga para principiantes en la mañana', 20, 'Lunes', '2024-01-01 10:00:00', '2024-01-01 11:00:00', 'https://example.com/yoga.jpg', 'María González', 'Yoga', 1, NOW(), NOW()),
('CrossFit Intenso', 'Entrenamiento de alta intensidad', 15, 'Martes', '2024-01-01 18:00:00', '2024-01-01 19:30:00', 'https://example.com/crossfit.jpg', 'Carlos Pérez', 'CrossFit', 1, NOW(), NOW()),
('Pilates Avanzado', 'Clase de pilates para nivel avanzado', 12, 'Miercoles', '2024-01-01 16:00:00', '2024-01-01 17:00:00', 'https://example.com/pilates.jpg', 'Laura Martínez', 'Pilates', 2, NOW(), NOW()),
('Spinning', 'Clase de spinning con música motivadora', 25, 'Jueves', '2024-01-01 19:00:00', '2024-01-01 20:00:00', 'https://example.com/spinning.jpg', 'Roberto Silva', 'Cardio', 2, NOW(), NOW()),
('Zumba', 'Baile fitness con ritmos latinos', 30, 'Viernes', '2024-01-01 20:00:00', '2024-01-01 21:00:00', 'https://example.com/zumba.jpg', 'Ana Rodríguez', 'Baile', 3, NOW(), NOW()),
('Funcional', 'Entrenamiento funcional para todo el cuerpo', 18, 'Sabado', '2024-01-01 09:00:00', '2024-01-01 10:00:00', 'https://example.com/funcional.jpg', 'Diego López', 'Funcional', 3, NOW(), NOW()),
('Yoga Vespertino', 'Clase de yoga relajante por la tarde', 20, 'Lunes', '2024-01-01 18:00:00', '2024-01-01 19:00:00', 'https://example.com/yoga2.jpg', 'María González', 'Yoga', 1, NOW(), NOW()),
('Boxeo', 'Clase de boxeo recreativo', 15, 'Martes', '2024-01-01 20:00:00', '2024-01-01 21:00:00', 'https://example.com/boxeo.jpg', 'Juan Ramírez', 'Boxeo', 2, NOW(), NOW());

-- Verificar datos insertados
SELECT 'Sucursales insertadas:' as info;
SELECT * FROM sucursales;

SELECT 'Actividades insertadas:' as info;
SELECT * FROM actividades;

SELECT 'Actividades con lugares disponibles:' as info;
SELECT id_actividad, titulo, cupo, dia, instructor, categoria FROM actividades;
