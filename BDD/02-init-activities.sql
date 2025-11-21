-- =====================================================
-- SCRIPT DE INICIALIZACIÓN - BASE DE DATOS ACTIVIDADES
-- Proyecto: Sistema de Gestión de Gimnasio
-- Base de Datos: gym_activities
-- =====================================================

CREATE DATABASE IF NOT EXISTS gym_activities;
USE gym_activities;

-- =====================================================
-- TABLA: sucursales
-- Gestiona las sucursales del gimnasio
-- =====================================================
CREATE TABLE IF NOT EXISTS sucursales (
    id_sucursal INT AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    direccion VARCHAR(255) NOT NULL,
    telefono VARCHAR(20),
    ciudad VARCHAR(100),
    codigo_postal VARCHAR(10),
    horario_apertura TIME,
    horario_cierre TIME,
    activa BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_activa (activa)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- TABLA: actividades
-- Gestiona las actividades ofrecidas en cada sucursal
-- =====================================================
CREATE TABLE IF NOT EXISTS actividades (
    id_actividad INT AUTO_INCREMENT PRIMARY KEY,
    titulo VARCHAR(100) NOT NULL,
    descripcion TEXT,
    cupo INT NOT NULL DEFAULT 20,
    dia VARCHAR(20) NOT NULL COMMENT 'Lunes, Martes, etc.',
    horario_inicio DATETIME NOT NULL,
    horario_final DATETIME NOT NULL,
    foto_url VARCHAR(255),
    instructor VARCHAR(100),
    categoria VARCHAR(50) COMMENT 'yoga, spinning, funcional, etc.',
    sucursal_id INT,
    activa BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (sucursal_id) REFERENCES sucursales(id_sucursal) ON DELETE SET NULL,
    INDEX idx_categoria (categoria),
    INDEX idx_dia (dia),
    INDEX idx_sucursal (sucursal_id),
    INDEX idx_activa (activa)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- TABLA: inscripciones
-- Gestiona las inscripciones de usuarios a actividades
-- =====================================================
CREATE TABLE IF NOT EXISTS inscripciones (
    id_inscripcion INT AUTO_INCREMENT PRIMARY KEY,
    usuario_id INT NOT NULL,
    actividad_id INT NOT NULL,
    suscripcion_id VARCHAR(50) NULL COMMENT 'ID de suscripción de MongoDB',
    is_activa BOOLEAN DEFAULT TRUE,
    fecha_inscripcion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT 'Soft delete timestamp',
    FOREIGN KEY (actividad_id) REFERENCES actividades(id_actividad) ON DELETE CASCADE,
    INDEX idx_usuario (usuario_id),
    INDEX idx_actividad (actividad_id),
    INDEX idx_activa (is_activa),
    INDEX idx_suscripcion (suscripcion_id),
    INDEX idx_deleted_at (deleted_at),
    UNIQUE KEY unique_usuario_actividad (usuario_id, actividad_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- DATOS INICIALES: Sucursales
-- =====================================================
INSERT INTO sucursales (nombre, direccion, telefono, ciudad, codigo_postal, horario_apertura, horario_cierre)
VALUES
    ('Sucursal Centro', 'Av. Corrientes 1234', '+54 11 4567-8900', 'Buenos Aires', '1043', '06:00:00', '23:00:00'),
    ('Sucursal Belgrano', 'Av. Cabildo 2500', '+54 11 4567-8901', 'Buenos Aires', '1428', '07:00:00', '22:00:00'),
    ('Sucursal Palermo', 'Av. Santa Fe 3500', '+54 11 4567-8902', 'Buenos Aires', '1425', '06:30:00', '22:30:00')
ON DUPLICATE KEY UPDATE nombre=nombre;

-- =====================================================
-- DATOS INICIALES: Actividades de ejemplo
-- =====================================================
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url)
VALUES
    ('Yoga Matutino', 'Clase de yoga para comenzar el día con energía', 20, 'Lunes', '2024-01-01 08:00:00', '2024-01-01 09:00:00', 'María González', 'yoga', 1, 'https://images.unsplash.com/photo-1544367567-0f2fcb009e0b'),
    ('Spinning Intenso', 'Entrenamiento cardiovascular de alta intensidad', 15, 'Martes', '2024-01-01 18:00:00', '2024-01-01 19:00:00', 'Carlos Pérez', 'spinning', 1, 'https://images.unsplash.com/photo-1534438327276-14e5300c3a48'),
    ('Funcional', 'Entrenamiento funcional para todo el cuerpo', 25, 'Miércoles', '2024-01-01 19:00:00', '2024-01-01 20:00:00', 'Laura Martínez', 'funcional', 2, 'https://images.unsplash.com/photo-1571019614242-c5c5dee9f50b'),
    ('Pilates', 'Fortalecimiento del core y flexibilidad', 18, 'Jueves', '2024-01-01 10:00:00', '2024-01-01 11:00:00', 'Ana Rodríguez', 'pilates', 2, 'https://images.unsplash.com/photo-1518611012118-696072aa579a'),
    ('CrossFit', 'Entrenamiento de alta intensidad variado', 20, 'Viernes', '2024-01-01 17:00:00', '2024-01-01 18:00:00', 'Javier López', 'crossfit', 3, 'https://images.unsplash.com/photo-1517836357463-d25dfeac3438'),
    ('Zumba', 'Baile fitness con ritmos latinos', 30, 'Sábado', '2024-01-01 11:00:00', '2024-01-01 12:00:00', 'Sofía Fernández', 'baile', 3, 'https://images.unsplash.com/photo-1518310383802-640c2de311b2')
ON DUPLICATE KEY UPDATE titulo=titulo;

-- =====================================================
-- MENSAJE DE CONFIRMACIÓN
-- =====================================================
SELECT '✅ Base de datos gym_activities inicializada correctamente' AS Status;
SELECT CONCAT('✅ Sucursales creadas: ', COUNT(*)) AS Sucursales FROM sucursales;
SELECT CONCAT('✅ Actividades creadas: ', COUNT(*)) AS Actividades FROM actividades;
