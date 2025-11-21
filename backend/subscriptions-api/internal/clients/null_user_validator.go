package clients

import (
	"context"
	"log"
)

// NullUserValidator - Validador que siempre retorna true
// Se usa cuando la validación del usuario ya se hace en el middleware JWT
type NullUserValidator struct{}

// NewNullUserValidator - Constructor
func NewNullUserValidator() *NullUserValidator {
	return &NullUserValidator{}
}

// ValidateUser - Siempre retorna true porque confiamos en el JWT
func (n *NullUserValidator) ValidateUser(ctx context.Context, userID string) (bool, error) {
	log.Printf("[NullUserValidator] ✅ Usuario %s validado (confiando en JWT)", userID)
	return true, nil
}
