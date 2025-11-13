#!/bin/bash

set -e  

echo "==========================================="
echo "  SETUP Y EJECUCIÓN DE ACTIVITIES-API"
echo "==========================================="
echo ""

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Función para imprimir mensajes
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 1. VERIFICAR SI GO ESTÁ INSTALADO
echo ""
print_info "Paso 1: Verificando Go..."
if ! command -v go &> /dev/null; then
    print_warning "Go no está instalado. Instalando Go..."

    sudo apt update
    sudo apt install -y golang-go

    print_info "Go instalado correctamente"
    go version
else
    print_info "Go ya está instalado: $(go version)"
fi

# 2. VERIFICAR VARIABLES DE ENTORNO
echo ""
print_info "Paso 2: Configurando variables de entorno..."

if [ ! -f ".env" ]; then
    print_warning "Archivo .env no encontrado. Creando desde .env.example..."

    if [ -f ".env.example" ]; then
        cp .env.example .env
        print_info "Archivo .env creado. IMPORTANTE: Edita el archivo .env con tus credenciales antes de continuar."
        print_warning "Presiona Enter cuando hayas configurado el archivo .env..."
        read
    else
        print_error "Archivo .env.example no encontrado."
        exit 1
    fi
else
    print_info "Archivo .env encontrado"
fi

# 3. DESCARGAR DEPENDENCIAS DE GO
echo ""
print_info "Paso 3: Descargando dependencias de Go..."
go mod download
go mod tidy

print_info "Dependencias descargadas correctamente"

# 4. VERIFICAR MYSQL
echo ""
print_info "Paso 4: Verificando MySQL..."
print_warning "Asegúrate de que MySQL esté corriendo y las credenciales en .env sean correctas"
print_warning "Si usas Docker Compose, ejecuta: docker-compose -f ../docker-compose.new.yml up -d mysql"
print_warning "Presiona Enter para continuar..."
read

# 5. COMPILAR EL PROYECTO
echo ""
print_info "Paso 5: Compilando el proyecto..."
go build -o activities-api cmd/api/main.go

if [ $? -eq 0 ]; then
    print_info "Compilación exitosa! ✓"
else
    print_error "Error en la compilación"
    exit 1
fi

# 6. EJECUTAR EL MICROSERVICIO
echo ""
print_info "Paso 6: Ejecutando activities-api..."
print_info "El servidor estará disponible en http://localhost:8082"
print_warning "Presiona Ctrl+C para detener el servidor"
echo ""

./activities-api
