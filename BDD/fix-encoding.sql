-- =====================================================
-- SCRIPT DE CORRECCIÓN DE ENCODING
-- Corrige los caracteres especiales en actividades
-- =====================================================

USE gym_activities;

-- Configurar encoding de la conexión
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- =====================================================
-- Actualizar actividades con caracteres correctos
-- =====================================================

UPDATE actividades
SET
    descripcion = 'Clase de yoga para comenzar el dia con energia y flexibilidad. Ideal para todos los niveles.',
    instructor = 'Maria Gonzalez'
WHERE id_actividad = 1;

UPDATE actividades
SET
    descripcion = 'Baile fitness con ritmos latinos. Quema calorias mientras te diviertes.',
    instructor = 'Sofia Fernandez',
    dia = 'Sabado'
WHERE id_actividad = 6;

UPDATE actividades
SET
    descripcion = 'Entrenamiento cardiovascular de alta intensidad en bicicleta estatica.',
    instructor = 'Carlos Perez'
WHERE id_actividad = 69;

UPDATE actividades
SET
    instructor = 'Laura Martinez'
WHERE id_actividad = 70;

UPDATE actividades
SET
    instructor = 'Ana Rodriguez'
WHERE id_actividad = 71;

UPDATE actividades
SET
    instructor = 'Javier Lopez'
WHERE id_actividad = 72;

UPDATE actividades
SET
    instructor = 'Maria Gonzalez'
WHERE id_actividad = 73;

UPDATE actividades
SET
    instructor = 'Carlos Perez',
    dia = 'Miercoles'
WHERE id_actividad = 74;

UPDATE actividades
SET
    instructor = 'Laura Martinez'
WHERE id_actividad = 75;

UPDATE actividades
SET
    instructor = 'Ana Rodriguez'
WHERE id_actividad = 76;

UPDATE actividades
SET
    instructor = 'Roberto Sanchez'
WHERE id_actividad = 77;

UPDATE actividades
SET
    instructor = 'Maria Gonzalez',
    dia = 'Miercoles'
WHERE id_actividad = 78;

SELECT '✅ Encoding corregido' AS Status;
SELECT id_actividad, titulo, dia, instructor
FROM actividades
WHERE deleted_at IS NULL AND activa = TRUE
ORDER BY id_actividad;
