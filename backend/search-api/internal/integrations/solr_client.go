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
// Solr puede devolver campos como arrays o valores únicos, usamos interface{} y extraemos el primer valor
type SolrDocument struct {
	ID              interface{} `json:"id"`
	Type            interface{} `json:"type"`
	Titulo          interface{} `json:"titulo,omitempty"`
	Descripcion     interface{} `json:"descripcion,omitempty"`
	Categoria       interface{} `json:"categoria,omitempty"`
	Instructor      interface{} `json:"instructor,omitempty"`
	Dia             interface{} `json:"dia,omitempty"`
	HorarioInicio   interface{} `json:"horario_inicio,omitempty"`
	HorarioFinal    interface{} `json:"horario_final,omitempty"`
	SucursalID      interface{} `json:"sucursal_id,omitempty"`
	SucursalNombre  interface{} `json:"sucursal_nombre,omitempty"`
	RequierePremium interface{} `json:"requiere_premium,omitempty"`
	CupoDisponible  interface{} `json:"cupo_disponible,omitempty"`
	PlanNombre      interface{} `json:"plan_nombre,omitempty"`
	PlanPrecio      interface{} `json:"plan_precio,omitempty"`
	PlanTipoAcceso  interface{} `json:"plan_tipo_acceso,omitempty"`
}

// Helper function to extract string from Solr field (can be string or []string)
func getSolrString(field interface{}) string {
	if field == nil {
		return ""
	}
	switch v := field.(type) {
	case string:
		return v
	case []interface{}:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				return str
			}
		}
	}
	return fmt.Sprintf("%v", field)
}

// Helper function to extract int from Solr field
func getSolrInt(field interface{}) int {
	if field == nil {
		return 0
	}
	switch v := field.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case []interface{}:
		if len(v) > 0 {
			if num, ok := v[0].(float64); ok {
				return int(num)
			}
		}
	}
	return 0
}

// Helper function to extract float64 from Solr field
func getSolrFloat(field interface{}) float64 {
	if field == nil {
		return 0
	}
	switch v := field.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case []interface{}:
		if len(v) > 0 {
			if num, ok := v[0].(float64); ok {
				return num
			}
		}
	}
	return 0
}

// Helper function to extract bool from Solr field
func getSolrBool(field interface{}) bool {
	if field == nil {
		return false
	}
	switch v := field.(type) {
	case bool:
		return v
	case []interface{}:
		if len(v) > 0 {
			if b, ok := v[0].(bool); ok {
				return b
			}
		}
	}
	return false
}

// SolrResponse - Respuesta de búsqueda de Solr
type SolrResponse struct {
	Response struct {
		NumFound int              `json:"numFound"`
		Start    int              `json:"start"`
		Docs     []SolrDocument   `json:"docs"`
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
		// Búsqueda en múltiples campos
		queryStr := fmt.Sprintf("titulo:%s^2 OR descripcion:%s OR categoria:%s OR instructor:%s",
			req.Query, req.Query, req.Query, req.Query)
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
	return SolrDocument{
		ID:              doc.ID,
		Type:            doc.Type,
		Titulo:          doc.Titulo,
		Descripcion:     doc.Descripcion,
		Categoria:       doc.Categoria,
		Instructor:      doc.Instructor,
		Dia:             doc.Dia,
		HorarioInicio:   doc.HorarioInicio,
		HorarioFinal:    doc.HorarioFinal,
		SucursalID:      doc.SucursalID,
		SucursalNombre:  doc.SucursalNombre,
		RequierePremium: doc.RequierePremium,
		CupoDisponible:  doc.CupoDisponible,
		PlanNombre:      doc.PlanNombre,
		PlanPrecio:      doc.PlanPrecio,
		PlanTipoAcceso:  doc.PlanTipoAcceso,
	}
}

// convertFromSolrDocument - Convierte SolrDocument a SearchDocument
func (c *SolrClient) convertFromSolrDocument(solrDoc SolrDocument) dtos.SearchDocument {
	return dtos.SearchDocument{
		ID:              getSolrString(solrDoc.ID),
		Type:            getSolrString(solrDoc.Type),
		Titulo:          getSolrString(solrDoc.Titulo),
		Descripcion:     getSolrString(solrDoc.Descripcion),
		Categoria:       getSolrString(solrDoc.Categoria),
		Instructor:      getSolrString(solrDoc.Instructor),
		Dia:             getSolrString(solrDoc.Dia),
		HorarioInicio:   getSolrString(solrDoc.HorarioInicio),
		HorarioFinal:    getSolrString(solrDoc.HorarioFinal),
		SucursalID:      getSolrString(solrDoc.SucursalID),
		SucursalNombre:  getSolrString(solrDoc.SucursalNombre),
		RequierePremium: getSolrBool(solrDoc.RequierePremium),
		CupoDisponible:  getSolrInt(solrDoc.CupoDisponible),
		PlanNombre:      getSolrString(solrDoc.PlanNombre),
		PlanPrecio:      getSolrFloat(solrDoc.PlanPrecio),
		PlanTipoAcceso:  getSolrString(solrDoc.PlanTipoAcceso),
	}
}
