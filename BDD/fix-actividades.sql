-- =====================================================
-- SCRIPT DE CORRECCIÓN DE ACTIVIDADES
-- Elimina actividades de test y restablece datos correctos
-- =====================================================

USE gym_activities;

-- =====================================================
-- PASO 1: Eliminar actividades de test (soft delete)
-- =====================================================
UPDATE actividades
SET deleted_at = NOW()
WHERE titulo IN (
    'Clase Especial con Cupo Limitado',
    'Actividad No Autorizada',
    'Funcional Test',
    'Yoga Test',
    'Spinning Test',
    'Yoga Unsub Test'
) AND deleted_at IS NULL;

-- =====================================================
-- PASO 2: Eliminar actividades duplicadas con encoding roto
-- =====================================================
UPDATE actividades
SET deleted_at = NOW()
WHERE (
    instructor LIKE '%�%'
    OR titulo LIKE '%�%'
    OR descripcion LIKE '%�%'
) AND deleted_at IS NULL;

-- =====================================================
-- PASO 3: Limpiar actividades viejas (mantener solo las iniciales)
-- =====================================================
UPDATE actividades
SET deleted_at = NOW()
WHERE id_actividad > 6 AND deleted_at IS NULL;

-- =====================================================
-- PASO 4: Actualizar actividades iniciales con datos correctos
-- =====================================================

-- Yoga Matutino (ID 1)
UPDATE actividades
SET
    titulo = 'Yoga Matutino',
    descripcion = 'Clase de yoga para comenzar el día con energía y flexibilidad. Ideal para todos los niveles.',
    cupo = 20,
    dia = 'Lunes',
    horario_inicio = '2024-01-01 08:00:00',
    horario_final = '2024-01-01 09:00:00',
    instructor = 'María González',
    categoria = 'yoga',
    sucursal_id = 1,
    foto_url = 'https://images.unsplash.com/photo-1544367567-0f2fcb009e0b',
    activa = TRUE
WHERE id_actividad = 1;

-- Zumba (ID 6)
UPDATE actividades
SET
    titulo = 'Zumba',
    descripcion = 'Baile fitness con ritmos latinos. Quema calorías mientras te diviertes.',
    cupo = 30,
    dia = 'Sábado',
    horario_inicio = '2024-01-01 11:00:00',
    horario_final = '2024-01-01 12:00:00',
    instructor = 'Sofía Fernández',
    categoria = 'baile',
    sucursal_id = 3,
    foto_url = 'https://images.unsplash.com/photo-1518310383802-640c2de311b2',
    activa = TRUE
WHERE id_actividad = 6;

-- =====================================================
-- PASO 5: Insertar actividades adicionales (si no existen)
-- =====================================================

-- Verificar y agregar Spinning Intenso
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url, activa)
SELECT 'Spinning Intenso',
       'Entrenamiento cardiovascular de alta intensidad en bicicleta estática.',
       15,
       'Martes',
       '2024-01-01 18:00:00',
       '2024-01-01 19:00:00',
       'Carlos Pérez',
       'spinning',
       1,
       'https://images.unsplash.com/photo-1534438327276-14e5300c3a48',
       TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM actividades
    WHERE titulo = 'Spinning Intenso'
    AND deleted_at IS NULL
    AND id_actividad <= 6
);

-- Verificar y agregar Funcional
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url, activa)
SELECT 'Funcional',
       'Entrenamiento funcional para mejorar fuerza, resistencia y movilidad.',
       25,
       'Miércoles',
       '2024-01-01 19:00:00',
       '2024-01-01 20:00:00',
       'Laura Martínez',
       'funcional',
       2,
       'https://images.unsplash.com/photo-1571019614242-c5c5dee9f50b',
       TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM actividades
    WHERE titulo = 'Funcional'
    AND deleted_at IS NULL
    AND id_actividad <= 6
);

-- Verificar y agregar Pilates
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url, activa)
SELECT 'Pilates',
       'Fortalecimiento del core y mejora de la postura. Trabaja mente y cuerpo.',
       18,
       'Jueves',
       '2024-01-01 10:00:00',
       '2024-01-01 11:00:00',
       'Ana Rodríguez',
       'pilates',
       2,
       'https://images.unsplash.com/photo-1518611012118-696072aa579a',
       TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM actividades
    WHERE titulo = 'Pilates'
    AND deleted_at IS NULL
    AND id_actividad <= 6
);

-- Verificar y agregar CrossFit
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url, activa)
SELECT 'CrossFit',
       'Entrenamiento de alta intensidad variado. Desarrolla fuerza y acondicionamiento.',
       20,
       'Viernes',
       '2024-01-01 17:00:00',
       '2024-01-01 18:00:00',
       'Javier López',
       'crossfit',
       3,
       'https://images.unsplash.com/photo-1517836357463-d25dfeac3438',
       TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM actividades
    WHERE titulo = 'CrossFit'
    AND deleted_at IS NULL
    AND id_actividad <= 6
);

-- =====================================================
-- PASO 6: Agregar actividades adicionales variadas
-- =====================================================

-- Yoga Vespertino
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url, activa)
VALUES
('Yoga Vespertino', 'Clase de yoga relajante para terminar el día. Ideal para liberar tensiones.', 20, 'Lunes', '2024-01-01 19:00:00', '2024-01-01 20:00:00', 'María González', 'yoga', 1, 'https://images.unsplash.com/photo-1506126613408-eca07ce68773', TRUE)
ON DUPLICATE KEY UPDATE titulo=titulo;

-- Spinning Matutino
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url, activa)
VALUES
('Spinning Matutino', 'Activa tu metabolismo con spinning a primera hora del día.', 15, 'Miércoles', '2024-01-01 07:00:00', '2024-01-01 08:00:00', 'Carlos Pérez', 'spinning', 2, 'https://images.unsplash.com/photo-1534438327276-14e5300c3a48', TRUE)
ON DUPLICATE KEY UPDATE titulo=titulo;

-- Funcional Avanzado
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url, activa)
VALUES
('Funcional Avanzado', 'Entrenamiento funcional de alta intensidad para nivel avanzado.', 20, 'Jueves', '2024-01-01 18:00:00', '2024-01-01 19:00:00', 'Laura Martínez', 'funcional', 1, 'https://images.unsplash.com/photo-1571019613454-1cb2f99b2d8b', TRUE)
ON DUPLICATE KEY UPDATE titulo=titulo;

-- Pilates Reformer
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url, activa)
VALUES
('Pilates Reformer', 'Pilates con equipo especializado. Mejora fuerza y flexibilidad.', 12, 'Viernes', '2024-01-01 10:00:00', '2024-01-01 11:00:00', 'Ana Rodríguez', 'pilates', 3, 'https://images.unsplash.com/photo-1518310383802-640c2de311b2', TRUE)
ON DUPLICATE KEY UPDATE titulo=titulo;

-- Boxeo
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url, activa)
VALUES
('Boxeo', 'Entrenamiento de boxeo. Mejora coordinación, fuerza y resistencia cardiovascular.', 16, 'Martes', '2024-01-01 19:00:00', '2024-01-01 20:00:00', 'Roberto Sánchez', 'boxeo', 2, 'https://images.unsplash.com/photo-1549719386-74dfcbf7dbed', TRUE)
ON DUPLICATE KEY UPDATE titulo=titulo;

-- Stretching
INSERT INTO actividades (titulo, descripcion, cupo, dia, horario_inicio, horario_final, instructor, categoria, sucursal_id, foto_url, activa)
VALUES
('Stretching', 'Clase de elongación y flexibilidad. Previene lesiones y mejora movilidad.', 25, 'Miércoles', '2024-01-01 20:00:00', '2024-01-01 21:00:00', 'María González', 'stretching', 1, 'https://images.unsplash.com/photo-1599901860904-17e6ed7083a0', TRUE)
ON DUPLICATE KEY UPDATE titulo=titulo;

-- =====================================================
-- PASO 7: Verificar resultado
-- =====================================================
SELECT '✅ Actividades corregidas exitosamente' AS Status;
SELECT CONCAT('📊 Total de actividades activas: ', COUNT(*)) AS Total
FROM actividades
WHERE deleted_at IS NULL AND activa = TRUE;

SELECT id_actividad, titulo, dia,
       TIME(horario_inicio) as hora_inicio,
       TIME(horario_final) as hora_fin,
       cupo, categoria, instructor, sucursal_id
FROM actividades
WHERE deleted_at IS NULL AND activa = TRUE
ORDER BY
    FIELD(dia, 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo'),
    horario_inicio;
