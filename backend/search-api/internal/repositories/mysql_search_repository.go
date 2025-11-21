package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/yourusername/gym-management/search-api/internal/domain/dtos"
)

// MySQLSearchRepository - Repositorio para búsqueda fulltext en MySQL (fallback)
type MySQLSearchRepository struct {
	db *sql.DB
}

// NewMySQLSearchRepository - Constructor
func NewMySQLSearchRepository(dbUser, dbPass, dbHost, dbPort, dbSchema string) (*MySQLSearchRepository, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbSchema)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MySQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging MySQL: %w", err)
	}

	return &MySQLSearchRepository{db: db}, nil
}

// SearchActivities - Busca actividades usando MySQL FULLTEXT
func (r *MySQLSearchRepository) SearchActivities(req dtos.SearchRequest) ([]dtos.SearchDocument, int, error) {
	// Base query with FULLTEXT search
	baseQuery := `
		SELECT
			a.id_actividad,
			a.titulo,
			a.descripcion,
			a.categoria,
			a.instructor,
			a.dia,
			a.horario_inicio,
			a.horario_final,
			a.cupo,
			COALESCE(s.nombre, '') as sucursal_nombre,
			a.sucursal_id,
			COUNT(*) OVER() as total_count
		FROM actividades a
		LEFT JOIN sucursales s ON a.sucursal_id = s.id_sucursal
	`

	var whereClauses []string
	var args []interface{}

	// Búsqueda fulltext si hay query
	if req.Query != "" {
		whereClauses = append(whereClauses, `(
			a.titulo LIKE ? OR
			a.descripcion LIKE ? OR
			a.categoria LIKE ? OR
			a.instructor LIKE ?
		)`)
		searchTerm := "%" + req.Query + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Filtros adicionales
	if categoria, ok := req.Filters["categoria"]; ok && categoria != "" {
		whereClauses = append(whereClauses, "a.categoria = ?")
		args = append(args, categoria)
	}

	if dia, ok := req.Filters["dia"]; ok && dia != "" {
		whereClauses = append(whereClauses, "a.dia = ?")
		args = append(args, dia)
	}

	if instructor, ok := req.Filters["instructor"]; ok && instructor != "" {
		whereClauses = append(whereClauses, "a.instructor LIKE ?")
		args = append(args, "%"+instructor+"%")
	}

	if sucursalID, ok := req.Filters["sucursal_id"]; ok && sucursalID != "" {
		whereClauses = append(whereClauses, "a.id_sucursal = ?")
		args = append(args, sucursalID)
	}

	// Construir query completa
	if len(whereClauses) > 0 {
		baseQuery += " AND " + strings.Join(whereClauses, " AND ")
	}

	// Ordenamiento
	baseQuery += " ORDER BY a.titulo ASC"

	// Paginación
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, req.PageSize, offset)

	// Ejecutar query
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing search query: %w", err)
	}
	defer rows.Close()

	var results []dtos.SearchDocument
	var totalCount int

	for rows.Next() {
		var doc dtos.SearchDocument
		var horarioInicio, horarioFinal string
		var sucursalID sql.NullString

		err := rows.Scan(
			&doc.ID,
			&doc.Titulo,
			&doc.Descripcion,
			&doc.Categoria,
			&doc.Instructor,
			&doc.Dia,
			&horarioInicio,
			&horarioFinal,
			&doc.CupoDisponible,
			&doc.SucursalNombre,
			&sucursalID,
			&totalCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning row: %w", err)
		}

		doc.Type = "activity"
		doc.HorarioInicio = horarioInicio
		doc.HorarioFinal = horarioFinal
		if sucursalID.Valid {
			doc.SucursalID = sucursalID.String
		}

		results = append(results, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, totalCount, nil
}

// GetAllActivities - Obtiene todas las actividades para indexación inicial
func (r *MySQLSearchRepository) GetAllActivities() ([]dtos.SearchDocument, error) {
	query := `
		SELECT
			a.id_actividad,
			a.titulo,
			a.descripcion,
			a.categoria,
			a.instructor,
			a.dia,
			a.horario_inicio,
			a.horario_final,
			a.cupo,
			COALESCE(s.nombre, '') as sucursal_nombre,
			a.sucursal_id
		FROM actividades a
		LEFT JOIN sucursales s ON a.sucursal_id = s.id_sucursal
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting all activities: %w", err)
	}
	defer rows.Close()

	var results []dtos.SearchDocument

	for rows.Next() {
		var doc dtos.SearchDocument
		var horarioInicio, horarioFinal string
		var sucursalID sql.NullString

		err := rows.Scan(
			&doc.ID,
			&doc.Titulo,
			&doc.Descripcion,
			&doc.Categoria,
			&doc.Instructor,
			&doc.Dia,
			&horarioInicio,
			&horarioFinal,
			&doc.CupoDisponible,
			&doc.SucursalNombre,
			&sucursalID,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning activity: %w", err)
		}

		doc.Type = "activity"
		doc.HorarioInicio = horarioInicio
		doc.HorarioFinal = horarioFinal
		if sucursalID.Valid {
			doc.SucursalID = sucursalID.String
		}

		results = append(results, doc)
	}

	return results, nil
}

// Close - Cierra la conexión a la base de datos
func (r *MySQLSearchRepository) Close() error {
	return r.db.Close()
}
