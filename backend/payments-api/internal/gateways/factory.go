package gateways

import (
	"fmt"

	"github.com/yourusername/payments-api/internal/config"
)

// GatewayFactory - Patrón Factory para crear instancias de gateways
// Encapsula la lógica de creación y configuración de cada gateway
// Facilita agregar nuevos gateways sin modificar código existente (Open/Closed Principle)
type GatewayFactory struct {
	config *config.Config
}

// NewGatewayFactory - Constructor del factory
func NewGatewayFactory(cfg *config.Config) *GatewayFactory {
	return &GatewayFactory{
		config: cfg,
	}
}

// CreateGateway - Crea una instancia del gateway solicitado
// Parámetros:
//   - gatewayName: Identificador del gateway ("mercadopago", "stripe", "cash", etc.)
// Retorna:
//   - PaymentGateway: Instancia del gateway configurado
//   - error: Si el gateway no está soportado o falta configuración
func (f *GatewayFactory) CreateGateway(gatewayName string) (PaymentGateway, error) {
	switch gatewayName {
	case "mercadopago":
		return f.createMercadoPagoGateway()
	case "cash", "efectivo":
		// Gateway para pagos en efectivo (manual, en sucursal)
		return NewCashGateway(), nil
	case "stripe":
		// TODO: Implementar Stripe en el futuro
		return nil, fmt.Errorf("gateway 'stripe' aún no implementado")
	case "paypal":
		// TODO: Implementar PayPal en el futuro
		return nil, fmt.Errorf("gateway 'paypal' aún no implementado")
	default:
		return nil, fmt.Errorf("gateway no soportado: %s", gatewayName)
	}
}

// createMercadoPagoGateway - Crea y configura una instancia de MercadoPagoGateway
func (f *GatewayFactory) createMercadoPagoGateway() (PaymentGateway, error) {
	// Validar que existan las credenciales
	if f.config.MercadoPago.AccessToken == "" {
		return nil, fmt.Errorf("falta configuración: MERCADOPAGO_ACCESS_TOKEN")
	}

	if f.config.MercadoPago.PublicKey == "" {
		return nil, fmt.Errorf("falta configuración: MERCADOPAGO_PUBLIC_KEY")
	}

	// Crear instancia con credenciales inyectadas
	gateway := NewMercadoPagoGateway(
		f.config.MercadoPago.AccessToken,
		f.config.MercadoPago.PublicKey,
		f.config.MercadoPago.WebhookSecret,
	)

	return gateway, nil
}

// CreateSubscriptionGateway - Crea una instancia del gateway de suscripciones
// Parámetros:
//   - gatewayName: Identificador del gateway ("mercadopago", "stripe", etc.)
// Retorna:
//   - SubscriptionGateway: Instancia del gateway configurado
//   - error: Si el gateway no está soportado o falta configuración
func (f *GatewayFactory) CreateSubscriptionGateway(gatewayName string) (SubscriptionGateway, error) {
	switch gatewayName {
	case "mercadopago":
		return f.createMercadoPagoSubscriptionGateway()
	case "stripe":
		// TODO: Implementar Stripe Subscriptions en el futuro
		return nil, fmt.Errorf("subscription gateway 'stripe' aún no implementado")
	default:
		return nil, fmt.Errorf("subscription gateway no soportado: %s", gatewayName)
	}
}

// createMercadoPagoSubscriptionGateway - Crea y configura una instancia de MercadoPagoSubscriptionGateway
func (f *GatewayFactory) createMercadoPagoSubscriptionGateway() (SubscriptionGateway, error) {
	// Validar que existan las credenciales
	if f.config.MercadoPago.AccessToken == "" {
		return nil, fmt.Errorf("falta configuración: MERCADOPAGO_ACCESS_TOKEN")
	}

	if f.config.MercadoPago.PublicKey == "" {
		return nil, fmt.Errorf("falta configuración: MERCADOPAGO_PUBLIC_KEY")
	}

	// Crear instancia con credenciales inyectadas
	gateway := NewMercadoPagoSubscriptionGateway(
		f.config.MercadoPago.AccessToken,
		f.config.MercadoPago.PublicKey,
		f.config.MercadoPago.WebhookSecret,
	)

	return gateway, nil
}

// GetSupportedGateways - Retorna la lista de gateways soportados
func (f *GatewayFactory) GetSupportedGateways() []string {
	return []string{
		"mercadopago",
		"cash",     // Pagos en efectivo (manual, en sucursal)
		"efectivo", // Alias para cash
		// "stripe",  // TODO: Futuro
		// "paypal",  // TODO: Futuro
	}
}

// ValidateGatewayName - Valida si un nombre de gateway es soportado
func (f *GatewayFactory) ValidateGatewayName(gatewayName string) bool {
	supported := f.GetSupportedGateways()
	for _, name := range supported {
		if name == gatewayName {
			return true
		}
	}
	return false
}
