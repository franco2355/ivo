package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/yourusername/gym-management/search-api/internal/domain/dtos"
)

// SolrClient - Cliente para interactuar con Apache Solr
type SolrClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewSolrClient - Constructor
func NewSolrClient(solrURL string) *SolrClient {
	return &SolrClient{
		baseURL:    strings.TrimSuffix(solrURL, "/"),
		httpClient: &http.Client{},
	}
}

// SolrDocument - Estructura de documento Solr
// Nota: Solr devuelve campos multivaluados como arrays, por eso usamos []string
type SolrDocument struct {
	ID              string    `json:"id"`
	Type            []string  `json:"type"` // Solr devuelve arrays para campos multivaluados
	Titulo          []string  `json:"titulo,omitempty"`
	Descripcion     []string  `json:"descripcion,omitempty"`
	Categoria       []string  `json:"categoria,omitempty"`
	Instructor      []string  `json:"instructor,omitempty"`
	Dia             []string  `json:"dia,omitempty"`
	HorarioInicio   []string  `json:"horario_inicio,omitempty"`
	HorarioFinal    []string  `json:"horario_final,omitempty"`
	SucursalID      []int     `json:"sucursal_id,omitempty"`
	SucursalNombre  []string  `json:"sucursal_nombre,omitempty"`
	RequierePremium []bool    `json:"requiere_premium,omitempty"`
	CupoDisponible  []int     `json:"cupo_disponible,omitempty"`
	PlanNombre      []string  `json:"plan_nombre,omitempty"`
	PlanPrecio      []float64 `json:"plan_precio,omitempty"`
	PlanTipoAcceso  []string  `json:"plan_tipo_acceso,omitempty"`
}

// SolrResponse - Respuesta de búsqueda de Solr
type SolrResponse struct {
	Response struct {
		NumFound int                      `json:"numFound"`
		Start    int                      `json:"start"`
		Docs     []map[string]interface{} `json:"docs"`
	} `json:"response"`
}

// IndexDocument - Indexa un documento en Solr
func (c *SolrClient) IndexDocument(doc dtos.SearchDocument) error {
	solrDoc := c.convertToSolrDocument(doc)

	payload := []SolrDocument{solrDoc}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	url := fmt.Sprintf("%s/update?commit=true", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request to Solr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("solr returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// IndexDocuments - Indexa múltiples documentos en batch
func (c *SolrClient) IndexDocuments(docs []dtos.SearchDocument) error {
	if len(docs) == 0 {
		return nil
	}

	solrDocs := make([]SolrDocument, len(docs))
	for i, doc := range docs {
		solrDocs[i] = c.convertToSolrDocument(doc)
	}

	jsonData, err := json.Marshal(solrDocs)
	if err != nil {
		return fmt.Errorf("error marshaling documents: %w", err)
	}

	url := fmt.Sprintf("%s/update?commit=true", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request to Solr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("solr returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Search - Realiza una búsqueda en Solr
func (c *SolrClient) Search(req dtos.SearchRequest) ([]dtos.SearchDocument, int, error) {
	// Construir query Solr
	params := url.Values{}

	// Query principal
	if req.Query != "" {
		// Búsqueda en múltiples campos incluyendo día
		queryStr := fmt.Sprintf("titulo:%s^2 OR descripcion:%s OR categoria:%s OR instructor:%s OR dia:%s",
			req.Query, req.Query, req.Query, req.Query, req.Query)
		params.Set("q", queryStr)
	} else {
		params.Set("q", "*:*")
	}

	// Filtro por tipo
	if req.Type != "" {
		params.Set("fq", fmt.Sprintf("type:%s", req.Type))
	}

	// Filtros adicionales
	var filterQueries []string
	for key, value := range req.Filters {
		if value != "" {
			filterQueries = append(filterQueries, fmt.Sprintf("%s:%s", key, value))
		}
	}
	if len(filterQueries) > 0 {
		existingFq := params.Get("fq")
		if existingFq != "" {
			filterQueries = append([]string{existingFq}, filterQueries...)
		}
		params["fq"] = filterQueries
	}

	// Paginación
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	start := (req.Page - 1) * req.PageSize
	params.Set("start", fmt.Sprintf("%d", start))
	params.Set("rows", fmt.Sprintf("%d", req.PageSize))

	// Formato de respuesta
	params.Set("wt", "json")

	// Construir URL completa
	searchURL := fmt.Sprintf("%s/select?%s", c.baseURL, params.Encode())

	// Ejecutar búsqueda
	resp, err := c.httpClient.Get(searchURL)
	if err != nil {
		return nil, 0, fmt.Errorf("error sending search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("solr returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var solrResp SolrResponse
	if err := json.NewDecoder(resp.Body).Decode(&solrResp); err != nil {
		return nil, 0, fmt.Errorf("error decoding Solr response: %w", err)
	}

	// Convertir resultados
	results := make([]dtos.SearchDocument, len(solrResp.Response.Docs))
	for i, solrDoc := range solrResp.Response.Docs {
		results[i] = c.convertFromSolrDocument(solrDoc)
	}

	return results, solrResp.Response.NumFound, nil
}

// DeleteDocument - Elimina un documento de Solr
func (c *SolrClient) DeleteDocument(id string) error {
	deleteQuery := map[string]interface{}{
		"delete": map[string]string{"id": id},
		"commit": map[string]interface{}{},
	}

	jsonData, err := json.Marshal(deleteQuery)
	if err != nil {
		return fmt.Errorf("error marshaling delete query: %w", err)
	}

	url := fmt.Sprintf("%s/update?commit=true", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating delete request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("solr delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Ping - Verifica la conectividad con Solr
func (c *SolrClient) Ping() error {
	url := fmt.Sprintf("%s/admin/ping", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("error pinging Solr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("solr ping failed with status %d", resp.StatusCode)
	}

	return nil
}

// convertToSolrDocument - Convierte SearchDocument a SolrDocument
func (c *SolrClient) convertToSolrDocument(doc dtos.SearchDocument) SolrDocument {
	solrDoc := SolrDocument{
		ID: doc.ID,
	}

	// Convertir strings a arrays para Solr
	if doc.Type != "" {
		solrDoc.Type = []string{doc.Type}
	}
	if doc.Titulo != "" {
		solrDoc.Titulo = []string{doc.Titulo}
	}
	if doc.Descripcion != "" {
		solrDoc.Descripcion = []string{doc.Descripcion}
	}
	if doc.Categoria != "" {
		solrDoc.Categoria = []string{doc.Categoria}
	}
	if doc.Instructor != "" {
		solrDoc.Instructor = []string{doc.Instructor}
	}
	if doc.Dia != "" {
		solrDoc.Dia = []string{doc.Dia}
	}
	if doc.HorarioInicio != "" {
		solrDoc.HorarioInicio = []string{doc.HorarioInicio}
	}
	if doc.HorarioFinal != "" {
		solrDoc.HorarioFinal = []string{doc.HorarioFinal}
	}
	if doc.SucursalID != "" {
		// Convertir string a int si es posible
		var sucursalID int
		fmt.Sscanf(doc.SucursalID, "%d", &sucursalID)
		if sucursalID > 0 {
			solrDoc.SucursalID = []int{sucursalID}
		}
	}
	if doc.SucursalNombre != "" {
		solrDoc.SucursalNombre = []string{doc.SucursalNombre}
	}
	if doc.RequierePremium {
		solrDoc.RequierePremium = []bool{doc.RequierePremium}
	}
	if doc.CupoDisponible > 0 {
		solrDoc.CupoDisponible = []int{doc.CupoDisponible}
	}
	if doc.PlanNombre != "" {
		solrDoc.PlanNombre = []string{doc.PlanNombre}
	}
	if doc.PlanPrecio > 0 {
		solrDoc.PlanPrecio = []float64{doc.PlanPrecio}
	}
	if doc.PlanTipoAcceso != "" {
		solrDoc.PlanTipoAcceso = []string{doc.PlanTipoAcceso}
	}

	return solrDoc
}

// Helper function to extract string value from interface{} (handles both string and []interface{})
func getStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []interface{}:
		if len(val) > 0 {
			if str, ok := val[0].(string); ok {
				return str
			}
		}
	}
	return ""
}

// Helper function to extract int value from interface{}
func getIntValue(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case []interface{}:
		if len(val) > 0 {
			if num, ok := val[0].(float64); ok {
				return int(num)
			}
		}
	}
	return 0
}

// Helper function to extract float64 value from interface{}
func getFloat64Value(v interface{}) float64 {
	if v == nil {
		return 0.0
	}
	switch val := v.(type) {
	case float64:
		return val
	case []interface{}:
		if len(val) > 0 {
			if num, ok := val[0].(float64); ok {
				return num
			}
		}
	}
	return 0.0
}

// Helper function to extract bool value from interface{}
func getBoolValue(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case []interface{}:
		if len(val) > 0 {
			if b, ok := val[0].(bool); ok {
				return b
			}
		}
	}
	return false
}

// convertFromSolrDocument - Convierte map[string]interface{} de Solr a SearchDocument
func (c *SolrClient) convertFromSolrDocument(solrDoc map[string]interface{}) dtos.SearchDocument {
	doc := dtos.SearchDocument{}

	if id, ok := solrDoc["id"].(string); ok {
		doc.ID = id
	}

	doc.Type = getStringValue(solrDoc["type"])
	doc.Titulo = getStringValue(solrDoc["titulo"])
	doc.Descripcion = getStringValue(solrDoc["descripcion"])
	doc.Categoria = getStringValue(solrDoc["categoria"])
	doc.Instructor = getStringValue(solrDoc["instructor"])
	doc.Dia = getStringValue(solrDoc["dia"])
	doc.HorarioInicio = getStringValue(solrDoc["horario_inicio"])
	doc.HorarioFinal = getStringValue(solrDoc["horario_final"])
	doc.SucursalNombre = getStringValue(solrDoc["sucursal_nombre"])
	doc.PlanNombre = getStringValue(solrDoc["plan_nombre"])
	doc.PlanTipoAcceso = getStringValue(solrDoc["plan_tipo_acceso"])

	sucursalID := getIntValue(solrDoc["sucursal_id"])
	if sucursalID > 0 {
		doc.SucursalID = fmt.Sprintf("%d", sucursalID)
	}

	doc.CupoDisponible = getIntValue(solrDoc["cupo_disponible"])
	doc.RequierePremium = getBoolValue(solrDoc["requiere_premium"])
	doc.PlanPrecio = getFloat64Value(solrDoc["plan_precio"])

	return doc
}
