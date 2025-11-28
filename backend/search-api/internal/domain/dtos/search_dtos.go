package dtos

// SearchDocument - DTO que representa un documento genérico indexado
// Puede ser una actividad, plan o suscripción
type SearchDocument struct {
	ID   string `json:"id"`
	Type string `json:"type"` // activity, plan, subscription

	// Campos de Actividad
	Titulo           string `json:"titulo,omitempty"`
	Descripcion      string `json:"descripcion,omitempty"`
	Categoria        string `json:"categoria,omitempty"`
	Instructor       string `json:"instructor,omitempty"`
	Dia              string `json:"dia,omitempty"`
	HorarioInicio    string `json:"horario_inicio,omitempty"`
	HorarioFinal     string `json:"horario_final,omitempty"`
	SucursalID       string `json:"sucursal_id,omitempty"`
	SucursalNombre   string `json:"sucursal_nombre,omitempty"`
	RequierePremium  bool   `json:"requiere_premium,omitempty"`
	Cupo             int    `json:"cupo,omitempty"`             // Cupo total
	CupoDisponible   int    `json:"cupo_disponible,omitempty"` // Lugares disponibles

	// Campos de Plan
	PlanNombre     string  `json:"plan_nombre,omitempty"`
	PlanPrecio     float64 `json:"plan_precio,omitempty"`
	PlanTipoAcceso string  `json:"plan_tipo_acceso,omitempty"`

	// Campos de Suscripción (admin)
	UsuarioID int    `json:"usuario_id,omitempty"`
	Estado    string `json:"estado,omitempty"`

	// Metadata
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// SearchRequest - DTO para peticiones de búsqueda
type SearchRequest struct {
	Query      string            `json:"query"`      // Término de búsqueda
	Filters    map[string]string `json:"filters"`    // Filtros: categoria, dia, etc.
	Type       string            `json:"type"`       // Filtrar por tipo: activity, plan, subscription
	Page       int               `json:"page"`       // Página (default: 1)
	PageSize   int               `json:"page_size"`  // Tamaño de página (default: 10)
	SortBy     string            `json:"sort_by"`    // Campo para ordenar
	SortOrder  string            `json:"sort_order"` // asc o desc
}

// SearchResponse - DTO de respuesta con resultados de búsqueda
type SearchResponse struct {
	Results    []SearchDocument `json:"results"`
	TotalCount int              `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// RabbitMQEvent - DTO para eventos recibidos de RabbitMQ
type RabbitMQEvent struct {
	Action    string                 `json:"action"` // create, update, delete
	Type      string                 `json:"type"`   // activity, plan, subscription
	ID        string                 `json:"id"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// StatsResponse - DTO de respuesta con estadísticas del índice
type StatsResponse struct {
	TotalDocuments int            `json:"total_documents"`
	DocumentsByType map[string]int `json:"documents_by_type"`
	CacheStats     *CacheStats    `json:"cache_stats,omitempty"`
}

// CacheStats - DTO con estadísticas del caché
type CacheStats struct {
	LocalCacheSize  int `json:"local_cache_size"`
	LocalCacheHits  int `json:"local_cache_hits"`
	LocalCacheMisses int `json:"local_cache_misses"`
}
