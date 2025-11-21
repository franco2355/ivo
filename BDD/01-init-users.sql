-- =====================================================
-- SCRIPT DE INICIALIZACIÓN - BASE DE DATOS USUARIOS
-- Proyecto: Sistema de Gestión de Gimnasio
-- Base de Datos: gym_users
-- =====================================================

CREATE DATABASE IF NOT EXISTS gym_users;
USE gym_users;

-- =====================================================
-- TABLA: usuarios
-- Gestiona los usuarios del sistema (clientes y admins)
-- =====================================================
CREATE TABLE IF NOT EXISTS usuarios (
    id_usuario INT AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    apellido VARCHAR(100) NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL COMMENT 'SHA-256 hash',
    is_admin TINYINT(1) NOT NULL DEFAULT 0,
    tipo VARCHAR(20) NOT NULL DEFAULT 'cliente' COMMENT 'cliente o admin',
    sucursal_origen_id INT NULL,
    fecha_registro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_email (email),
    INDEX idx_username (username),
    INDEX idx_tipo (tipo),
    INDEX idx_sucursal_origen (sucursal_origen_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- DATOS INICIALES: Usuario Admin
-- Password: admin123 (SHA-256)
-- =====================================================
INSERT INTO usuarios (nombre, apellido, username, email, password, tipo)
VALUES
    ('Admin', 'Sistema', 'admin', 'admin@gym.com',
     '240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9',
     'admin')
ON DUPLICATE KEY UPDATE username=username;

-- =====================================================
-- MENSAJE DE CONFIRMACIÓN
-- =====================================================
SELECT '✅ Base de datos gym_users inicializada correctamente' AS Status;
SELECT CONCAT('✅ Usuarios creados: ', COUNT(*)) AS Usuarios FROM usuarios;
