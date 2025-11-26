import { describe, it, expect } from 'vitest';
import {
    isValidEmail,
    validatePassword,
    isRequired,
    hasMinLength,
    hasMaxLength,
    isInRange,
    sanitizeString
} from './validation';

describe('Validation Utils', () => {
    describe('isValidEmail', () => {
        it('should validate correct email addresses', () => {
            expect(isValidEmail('test@example.com')).toBe(true);
            expect(isValidEmail('user.name+tag@example.co.uk')).toBe(true);
        });

        it('should reject invalid email addresses', () => {
            expect(isValidEmail('invalid')).toBe(false);
            expect(isValidEmail('invalid@')).toBe(false);
            expect(isValidEmail('@example.com')).toBe(false);
            expect(isValidEmail('')).toBe(false);
        });
    });

    describe('validatePassword', () => {
        it('should validate strong passwords', () => {
            const result = validatePassword('Password123!');
            expect(result.valid).toBe(true);
            expect(result.errors).toHaveLength(0);
        });

        it('should reject passwords that are too short', () => {
            const result = validatePassword('Pass1!');
            expect(result.valid).toBe(false);
            expect(result.errors).toContain('al menos 8 caracteres');
        });

        it('should reject passwords without uppercase', () => {
            const result = validatePassword('password123!');
            expect(result.valid).toBe(false);
            expect(result.errors).toContain('una letra mayúscula');
        });

        it('should reject passwords without lowercase', () => {
            const result = validatePassword('PASSWORD123!');
            expect(result.valid).toBe(false);
            expect(result.errors).toContain('una letra minúscula');
        });

        it('should reject passwords without numbers', () => {
            const result = validatePassword('Password!');
            expect(result.valid).toBe(false);
            expect(result.errors).toContain('un número');
        });

        it('should reject passwords without special characters', () => {
            const result = validatePassword('Password123');
            expect(result.valid).toBe(false);
            expect(result.errors).toContain('un carácter especial');
        });
    });

    describe('isRequired', () => {
        it('should accept non-empty values', () => {
            expect(isRequired('test')).toBe(true);
            expect(isRequired('0')).toBe(true);
            expect(isRequired(123)).toBe(true);
        });

        it('should reject empty values', () => {
            expect(isRequired('')).toBe(false);
            expect(isRequired('   ')).toBe(false);
            expect(isRequired(null)).toBe(false);
            expect(isRequired(undefined)).toBe(false);
        });
    });

    describe('hasMinLength', () => {
        it('should validate minimum length', () => {
            expect(hasMinLength('test', 3)).toBe(true);
            expect(hasMinLength('test', 4)).toBe(true);
            expect(hasMinLength('test', 5)).toBe(false);
        });
    });

    describe('hasMaxLength', () => {
        it('should validate maximum length', () => {
            expect(hasMaxLength('test', 5)).toBe(true);
            expect(hasMaxLength('test', 4)).toBe(true);
            expect(hasMaxLength('test', 3)).toBe(false);
        });
    });

    describe('isInRange', () => {
        it('should validate numbers in range', () => {
            expect(isInRange(5, 1, 10)).toBe(true);
            expect(isInRange(1, 1, 10)).toBe(true);
            expect(isInRange(10, 1, 10)).toBe(true);
            expect(isInRange(0, 1, 10)).toBe(false);
            expect(isInRange(11, 1, 10)).toBe(false);
        });
    });

    describe('sanitizeString', () => {
        it('should escape HTML characters', () => {
            expect(sanitizeString('<script>alert("XSS")</script>'))
                .toBe('&lt;script&gt;alert(&quot;XSS&quot;)&lt;&#x2F;script&gt;');
        });

        it('should handle empty strings', () => {
            expect(sanitizeString('')).toBe('');
            expect(sanitizeString(null)).toBe('');
            expect(sanitizeString(undefined)).toBe('');
        });
    });
});
