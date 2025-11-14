package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"proyecto-integrador/clients/usuario"
	"proyecto-integrador/dto"
	"proyecto-integrador/model"
	"time"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

type usuarioService struct{}

type IUsuarioService interface {
	GenerateToken(username string, password string) (string, error)
	GetClaimsFromToken(tokenString string) (jwt.MapClaims, error)
	RegisterUser(datos dto.UsuarioMinDTO) error
}

var (
	IncorrectCredentialsError = errors.New("Credenciales incorrectas")

	UsuarioService IUsuarioService
	jwtSecret      string
)

func init() {
	UsuarioService = &usuarioService{}

	jwtSecret = os.Getenv("JWT_SECRET")
}

func calculateSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func (us *usuarioService) GenerateToken(username string, password string) (string, error) {
	var userdata model.Usuario = usuario.GetUsuarioByUsername(username)

	if calculateSHA256(password) != userdata.Password {
		log.Debugf("Contraseña incorrecta para el usuario %s@%s\n", username, password)
		return "", IncorrectCredentialsError
	}

	// Determinar el rol basado en is_admin
	role := "user"
	if userdata.IsAdmin {
		role = "admin"
	}

	claims := jwt.MapClaims{
		"iss":        "proyecto2025-morini-heredia",
		"exp":        time.Now().Add(30 * time.Minute).Unix(),
		"username":   userdata.Username,
		"id_usuario": userdata.Id,
		"user_id":    fmt.Sprintf("%d", userdata.Id), // Para compatibilidad con otras APIs
		"is_admin":   userdata.IsAdmin,
		"role":       role, // Para compatibilidad con otras APIs
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtSecret))
}

func (us *usuarioService) GetClaimsFromToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Error al obtener los claims")
	}

	return claims, nil
}

func (us *usuarioService) RegisterUser(datos dto.UsuarioMinDTO) error {
	var newUser model.Usuario = model.Usuario{
		Nombre:   datos.Nombre,
		Apellido: datos.Apellido,
		Username: datos.Username,
		Password: calculateSHA256(datos.Password),
	}
	return usuario.RegisterUser(newUser)
}
