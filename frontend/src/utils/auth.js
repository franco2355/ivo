import { ERROR_MESSAGES } from './constants';

/**
 * Maneja la sesión expirada
 * - Limpia el localStorage
 * - Muestra mensaje de error
 * - Redirecciona al login
 */
export const handleSessionExpired = (toast, navigate) => {
    // Limpiar localStorage
    localStorage.removeItem('access_token');
    localStorage.removeItem('idUsuario');
    localStorage.removeItem('isAdmin');
    localStorage.removeItem('isLoggedIn');
    localStorage.removeItem('nombre');

    // Mostrar mensaje de error
    if (toast) {
        toast.error(ERROR_MESSAGES.SESSION_EXPIRED);
    }

    // Redireccionar al login
    if (navigate) {
        navigate('/login');
    }
};

/**
 * Verifica si una respuesta HTTP es un error de autenticación
 * @param {Response} response - La respuesta HTTP
 * @returns {boolean} - true si es error 401
 */
export const isAuthError = (response) => {
    return response.status === 401;
};
