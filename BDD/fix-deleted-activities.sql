-- Arreglar actividades: marcar como inactivas las que están soft deleted
-- Las actividades con deleted_at deben tener activa=false
-- Base de datos: gym_activities

-- EJECUTADO EXITOSAMENTE - Este script ya fue aplicado
-- Comando usado: docker exec gym-mysql sh -c 'mysql -u root -p"${MYSQL_ROOT_PASSWORD}" gym_activities -e "UPDATE actividades SET activa = false WHERE deleted_at IS NOT NULL;"'

UPDATE actividades
SET activa = false
WHERE deleted_at IS NOT NULL;

-- Verificar que solo las 12 actividades buenas estén activas
SELECT COUNT(*) as total_activas
FROM actividades
WHERE activa = true AND deleted_at IS NULL;

-- RESULTADO: 12 actividades activas (correcto)
