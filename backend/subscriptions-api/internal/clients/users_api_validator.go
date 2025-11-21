package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UsersAPIValidator - Implementación de UserValidator que consulta users-api
type UsersAPIValidator struct {
	baseURL string
	client  *http.Client
}

// NewUsersAPIValidator - Constructor con DI
func NewUsersAPIValidator(baseURL string) *UsersAPIValidator {
	return &UsersAPIValidator{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type userResponse struct {
	ID       int    `json:"id_usuario"`
	Username string `json:"username"`
}

// ValidateUser - Implementa la interface UserValidator
func (u *UsersAPIValidator) ValidateUser(ctx context.Context, userID string) (bool, error) {
	url := fmt.Sprintf("%s/users/%s", u.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("error creando request: %w", err)
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error consultando users-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, fmt.Errorf("usuario no encontrado")
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("error en users-api: status %d", resp.StatusCode)
	}

	var user userResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return false, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	return user.ID > 0, nil
}
