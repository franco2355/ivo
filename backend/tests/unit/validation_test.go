package unit

import (
	"testing"
	"time"
)

// TestEmailValidation verifica validación de emails
func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		isValid bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			isValid: true,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			isValid: true,
		},
		{
			name:    "valid email with plus",
			email:   "user+tag@example.com",
			isValid: true,
		},
		{
			name:    "invalid email - no @",
			email:   "userexample.com",
			isValid: false,
		},
		{
			name:    "invalid email - no domain",
			email:   "user@",
			isValid: false,
		},
		{
			name:    "invalid email - no local part",
			email:   "@example.com",
			isValid: false,
		},
		{
			name:    "invalid email - no TLD",
			email:   "user@example",
			isValid: false,
		},
		{
			name:    "invalid email - spaces",
			email:   "user @example.com",
			isValid: false,
		},
		{
			name:    "empty email",
			email:   "",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidEmail(tt.email)

			if isValid != tt.isValid {
				t.Errorf("Email '%s': expected valid=%v, got valid=%v",
					tt.email, tt.isValid, isValid)
			}
		})
	}
}

// TestUsernameValidation verifica validación de nombres de usuario
func TestUsernameValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		isValid  bool
		errorMsg string
	}{
		{
			name:     "valid username",
			username: "john_doe",
			isValid:  true,
		},
		{
			name:     "valid username with numbers",
			username: "user123",
			isValid:  true,
		},
		{
			name:     "valid username with hyphen",
			username: "john-doe",
			isValid:  true,
		},
		{
			name:     "too short",
			username: "ab",
			isValid:  false,
			errorMsg: "username must be between 3 and 30 characters",
		},
		{
			name:     "too long",
			username: "this_is_a_very_long_username_that_exceeds_limit",
			isValid:  false,
			errorMsg: "username must be between 3 and 30 characters",
		},
		{
			name:     "contains spaces",
			username: "john doe",
			isValid:  false,
			errorMsg: "username can only contain letters, numbers, hyphens and underscores",
		},
		{
			name:     "contains special chars",
			username: "john@doe",
			isValid:  false,
			errorMsg: "username can only contain letters, numbers, hyphens and underscores",
		},
		{
			name:     "starts with number",
			username: "123user",
			isValid:  true, // Depende de las reglas del negocio
		},
		{
			name:     "empty username",
			username: "",
			isValid:  false,
			errorMsg: "username is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.username)

			if tt.isValid {
				if err != nil {
					t.Errorf("Expected valid username, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid username, got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got: '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

// TestPriceValidation verifica validación de precios
func TestPriceValidation(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		isValid  bool
		errorMsg string
	}{
		{
			name:    "valid price",
			price:   1000.50,
			isValid: true,
		},
		{
			name:    "valid price - integer",
			price:   5000,
			isValid: true,
		},
		{
			name:     "negative price",
			price:    -100,
			isValid:  false,
			errorMsg: "price must be positive",
		},
		{
			name:     "zero price",
			price:    0,
			isValid:  false,
			errorMsg: "price must be positive",
		},
		{
			name:    "very small price",
			price:   0.01,
			isValid: true,
		},
		{
			name:    "very large price",
			price:   999999.99,
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePrice(tt.price)

			if tt.isValid {
				if err != nil {
					t.Errorf("Expected valid price, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid price, got nil")
				}
			}
		})
	}
}

// TestDateValidation verifica validación de fechas
func TestDateValidation(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)
	nextWeek := now.Add(7 * 24 * time.Hour)

	tests := []struct {
		name     string
		date     time.Time
		minDate  time.Time
		maxDate  time.Time
		isValid  bool
		errorMsg string
	}{
		{
			name:    "date within range",
			date:    tomorrow,
			minDate: now,
			maxDate: nextWeek,
			isValid: true,
		},
		{
			name:     "date before minimum",
			date:     yesterday,
			minDate:  now,
			maxDate:  nextWeek,
			isValid:  false,
			errorMsg: "date is before minimum allowed",
		},
		{
			name:     "date after maximum",
			date:     nextWeek.Add(24 * time.Hour),
			minDate:  now,
			maxDate:  nextWeek,
			isValid:  false,
			errorMsg: "date is after maximum allowed",
		},
		{
			name:    "date equals minimum",
			date:    now,
			minDate: now,
			maxDate: nextWeek,
			isValid: true,
		},
		{
			name:    "date equals maximum",
			date:    nextWeek,
			minDate: now,
			maxDate: nextWeek,
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateRange(tt.date, tt.minDate, tt.maxDate)

			if tt.isValid {
				if err != nil {
					t.Errorf("Expected valid date, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid date, got nil")
				}
			}
		})
	}
}

// TestTimeSlotValidation verifica validación de horarios
func TestTimeSlotValidation(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		end      string
		isValid  bool
		errorMsg string
	}{
		{
			name:    "valid time slot",
			start:   "09:00",
			end:     "10:00",
			isValid: true,
		},
		{
			name:    "valid time slot - afternoon",
			start:   "14:00",
			end:     "15:30",
			isValid: true,
		},
		{
			name:     "end before start",
			start:    "10:00",
			end:      "09:00",
			isValid:  false,
			errorMsg: "end time must be after start time",
		},
		{
			name:     "same start and end",
			start:    "10:00",
			end:      "10:00",
			isValid:  false,
			errorMsg: "end time must be after start time",
		},
		{
			name:     "invalid start format",
			start:    "25:00",
			end:      "10:00",
			isValid:  false,
			errorMsg: "invalid time format",
		},
		{
			name:     "invalid end format",
			start:    "09:00",
			end:      "10:70",
			isValid:  false,
			errorMsg: "invalid time format",
		},
		{
			name:    "crossing midnight",
			start:   "23:00",
			end:     "01:00",
			isValid: true, // Depende de las reglas del negocio
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTimeSlot(tt.start, tt.end)

			if tt.isValid {
				if err != nil {
					t.Errorf("Expected valid time slot, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid time slot, got nil")
				}
			}
		})
	}
}

// TestCapacityValidation verifica validación de cupos
func TestCapacityValidation(t *testing.T) {
	tests := []struct {
		name         string
		capacity     int
		currentCount int
		isAvailable  bool
	}{
		{
			name:         "has available spots",
			capacity:     20,
			currentCount: 15,
			isAvailable:  true,
		},
		{
			name:         "one spot left",
			capacity:     20,
			currentCount: 19,
			isAvailable:  true,
		},
		{
			name:         "fully booked",
			capacity:     20,
			currentCount: 20,
			isAvailable:  false,
		},
		{
			name:         "overbooked",
			capacity:     20,
			currentCount: 21,
			isAvailable:  false,
		},
		{
			name:         "empty",
			capacity:     20,
			currentCount: 0,
			isAvailable:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAvailable := hasAvailableCapacity(tt.capacity, tt.currentCount)

			if isAvailable != tt.isAvailable {
				t.Errorf("Expected available=%v, got available=%v",
					tt.isAvailable, isAvailable)
			}
		})
	}
}

// TestPhoneValidation verifica validación de números de teléfono
func TestPhoneValidation(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		isValid bool
	}{
		{
			name:    "valid phone - 10 digits",
			phone:   "1234567890",
			isValid: true,
		},
		{
			name:    "valid phone with country code",
			phone:   "+541234567890",
			isValid: true,
		},
		{
			name:    "valid phone with spaces",
			phone:   "123 456 7890",
			isValid: true,
		},
		{
			name:    "valid phone with hyphens",
			phone:   "123-456-7890",
			isValid: true,
		},
		{
			name:    "too short",
			phone:   "123456",
			isValid: false,
		},
		{
			name:    "contains letters",
			phone:   "123abc7890",
			isValid: false,
		},
		{
			name:    "empty phone",
			phone:   "",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidPhone(tt.phone)

			if isValid != tt.isValid {
				t.Errorf("Phone '%s': expected valid=%v, got valid=%v",
					tt.phone, tt.isValid, isValid)
			}
		})
	}
}

// ============================================
// Helper/Stub Functions (implementaciones básicas para testing)
// En producción, estas estarían en el código real
// ============================================

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	// Validación simplificada
	hasAt := false
	hasDot := false
	for i, c := range email {
		if c == '@' {
			if i == 0 || i == len(email)-1 {
				return false
			}
			hasAt = true
		}
		if c == '.' && hasAt {
			hasDot = true
		}
		if c == ' ' {
			return false
		}
	}
	return hasAt && hasDot
}

func validateUsername(username string) error {
	if username == "" {
		return &ValidationError{"username is required"}
	}
	if len(username) < 3 || len(username) > 30 {
		return &ValidationError{"username must be between 3 and 30 characters"}
	}
	// Verificar caracteres válidos
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-') {
			return &ValidationError{"username can only contain letters, numbers, hyphens and underscores"}
		}
	}
	return nil
}

func validatePrice(price float64) error {
	if price <= 0 {
		return &ValidationError{"price must be positive"}
	}
	return nil
}

func validateDateRange(date, minDate, maxDate time.Time) error {
	if date.Before(minDate) {
		return &ValidationError{"date is before minimum allowed"}
	}
	if date.After(maxDate) {
		return &ValidationError{"date is after maximum allowed"}
	}
	return nil
}

func validateTimeSlot(start, end string) error {
	// Validación simplificada
	if len(start) != 5 || len(end) != 5 {
		return &ValidationError{"invalid time format"}
	}

	// Parsear horas y minutos
	var startH, startM, endH, endM int
	_, err := parseTime(start, &startH, &startM)
	if err != nil {
		return &ValidationError{"invalid time format"}
	}
	_, err = parseTime(end, &endH, &endM)
	if err != nil {
		return &ValidationError{"invalid time format"}
	}

	// Comparar
	startMinutes := startH*60 + startM
	endMinutes := endH*60 + endM

	if endMinutes <= startMinutes {
		return &ValidationError{"end time must be after start time"}
	}

	return nil
}

func parseTime(timeStr string, h, m *int) (bool, error) {
	if len(timeStr) != 5 || timeStr[2] != ':' {
		return false, &ValidationError{"invalid format"}
	}

	var hour, min int
	for i := 0; i < 2; i++ {
		if timeStr[i] < '0' || timeStr[i] > '9' {
			return false, &ValidationError{"invalid format"}
		}
		hour = hour*10 + int(timeStr[i]-'0')
	}
	for i := 3; i < 5; i++ {
		if timeStr[i] < '0' || timeStr[i] > '9' {
			return false, &ValidationError{"invalid format"}
		}
		min = min*10 + int(timeStr[i]-'0')
	}

	if hour >= 24 || min >= 60 {
		return false, &ValidationError{"invalid time values"}
	}

	*h = hour
	*m = min
	return true, nil
}

func hasAvailableCapacity(capacity, currentCount int) bool {
	return currentCount < capacity
}

func isValidPhone(phone string) bool {
	if phone == "" {
		return false
	}

	digitCount := 0
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			digitCount++
		} else if c != ' ' && c != '-' && c != '+' && c != '(' && c != ')' {
			return false
		}
	}

	return digitCount >= 10
}

// ValidationError es un tipo de error personalizado
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
