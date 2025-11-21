import { PASSWORD_VALIDATION, REGEX_PATTERNS } from './constants';

/**
 * Valida un email
 * @param {string} email
 * @returns {boolean}
 */
export const isValidEmail = (email) => {
    return REGEX_PATTERNS.EMAIL.test(email);
};

/**
 * Valida una contraseña según los requisitos definidos
 * @param {string} password
 * @returns {{valid: boolean, errors: string[]}}
 */
export const validatePassword = (password) => {
    const errors = [];

    if (password.length < PASSWORD_VALIDATION.MIN_LENGTH) {
        errors.push(`al menos ${PASSWORD_VALIDATION.MIN_LENGTH} caracteres`);
    }

    if (PASSWORD_VALIDATION.REQUIRE_UPPERCASE && !/[A-Z]/.test(password)) {
        errors.push('una letra mayúscula');
    }

    if (PASSWORD_VALIDATION.REQUIRE_LOWERCASE && !/[a-z]/.test(password)) {
        errors.push('una letra minúscula');
    }

    if (PASSWORD_VALIDATION.REQUIRE_NUMBER && !/[0-9]/.test(password)) {
        errors.push('un número');
    }

    if (PASSWORD_VALIDATION.REQUIRE_SPECIAL && !/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
        errors.push('un carácter especial');
    }

    return {
        valid: errors.length === 0,
        errors
    };
};

/**
 * Valida que un campo no esté vacío
 * @param {string} value
 * @returns {boolean}
 */
export const isRequired = (value) => {
    return value !== null && value !== undefined && value.toString().trim() !== '';
};

/**
 * Valida la longitud mínima de un string
 * @param {string} value
 * @param {number} minLength
 * @returns {boolean}
 */
export const hasMinLength = (value, minLength) => {
    return value && value.length >= minLength;
};

/**
 * Valida la longitud máxima de un string
 * @param {string} value
 * @param {number} maxLength
 * @returns {boolean}
 */
export const hasMaxLength = (value, maxLength) => {
    return value && value.length <= maxLength;
};

/**
 * Valida que un número esté en un rango
 * @param {number} value
 * @param {number} min
 * @param {number} max
 * @returns {boolean}
 */
export const isInRange = (value, min, max) => {
    const num = Number(value);
    return !isNaN(num) && num >= min && num <= max;
};

/**
 * Valida múltiples campos de un formulario
 * @param {Object} formData
 * @param {Object} rules - Objeto con reglas de validación por campo
 * @returns {{valid: boolean, errors: Object}}
 *
 * Ejemplo:
 * validateForm(formData, {
 *   email: [
 *     { validator: isRequired, message: 'El email es requerido' },
 *     { validator: isValidEmail, message: 'Email inválido' }
 *   ],
 *   password: [
 *     { validator: isRequired, message: 'La contraseña es requerida' }
 *   ]
 * })
 */
export const validateForm = (formData, rules) => {
    const errors = {};
    let valid = true;

    for (const field in rules) {
        const fieldRules = rules[field];
        const value = formData[field];

        for (const rule of fieldRules) {
            if (typeof rule.validator === 'function') {
                const isValid = rule.validator(value);
                if (!isValid) {
                    errors[field] = rule.message;
                    valid = false;
                    break; // Solo mostrar el primer error por campo
                }
            }
        }
    }

    return { valid, errors };
};

/**
 * Sanitiza un string removiendo caracteres peligrosos
 * @param {string} str
 * @returns {string}
 */
export const sanitizeString = (str) => {
    if (!str) return '';
    return str
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#x27;')
        .replace(/\//g, '&#x2F;');
};
