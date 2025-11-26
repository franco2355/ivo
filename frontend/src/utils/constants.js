// Constantes de la aplicación

// HTTP Timeouts (en milisegundos)
export const HTTP_TIMEOUT = {
    DEFAULT: 5000,
    LONG: 10000,
    SHORT: 2000
};

// Debounce delays (en milisegundos)
export const DEBOUNCE_DELAY = {
    SEARCH: 500,
    INPUT: 300,
    FAST: 150
};

// Paginación
export const PAGINATION = {
    DEFAULT_PAGE_SIZE: 10,
    MAX_PAGE_SIZE: 100,
    SEARCH_PAGE_SIZE: 100
};

// Validación de contraseñas
export const PASSWORD_VALIDATION = {
    MIN_LENGTH: 8,
    REQUIRE_UPPERCASE: true,
    REQUIRE_LOWERCASE: true,
    REQUIRE_NUMBER: true,
    REQUIRE_SPECIAL: true
};

// Tamaños de spinners
export const SPINNER_SIZES = {
    SMALL: 'small',
    MEDIUM: 'medium',
    LARGE: 'large'
};

// Estados de suscripción
export const SUBSCRIPTION_STATUS = {
    ACTIVE: 'activa',
    INACTIVE: 'inactiva',
    PENDING: 'pendiente',
    CANCELLED: 'cancelada'
};

// Estados de pago
export const PAYMENT_STATUS = {
    APPROVED: 'approved',
    PENDING: 'pending',
    REJECTED: 'rejected',
    CANCELLED: 'cancelled'
};

// Mensajes de error comunes
export const ERROR_MESSAGES = {
    SESSION_EXPIRED: 'Tu sesión ha expirado. Por favor, inicia sesión nuevamente.',
    NETWORK_ERROR: 'Error de conexión. Por favor, verifica tu conexión a internet.',
    GENERIC_ERROR: 'Ocurrió un error inesperado. Por favor, intenta nuevamente.',
    UNAUTHORIZED: 'No tienes permisos para realizar esta acción.',
    NOT_FOUND: 'El recurso solicitado no fue encontrado.',
    VALIDATION_ERROR: 'Por favor, verifica los datos ingresados.',
    SERVER_ERROR: 'Error del servidor. Por favor, intenta más tarde.'
};

// Mensajes de éxito comunes
export const SUCCESS_MESSAGES = {
    SAVE_SUCCESS: 'Guardado exitosamente',
    DELETE_SUCCESS: 'Eliminado exitosamente',
    UPDATE_SUCCESS: 'Actualizado exitosamente',
    CREATE_SUCCESS: 'Creado exitosamente'
};

// Delays para animaciones (en milisegundos)
export const ANIMATION_DELAYS = {
    TOAST_DURATION: 3000,
    MODAL_FADE: 300,
    TOOLTIP_DELAY: 500
};

// Regex patterns
export const REGEX_PATTERNS = {
    EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
    PHONE: /^\d{10}$/,
    ALPHANUMERIC: /^[a-zA-Z0-9]+$/,
    NUMBERS_ONLY: /^\d+$/
};
