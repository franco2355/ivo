import { useState, useEffect } from 'react';
import { DEBOUNCE_DELAY } from '../utils/constants';

/**
 * Hook personalizado para debouncing de valores
 * @param {*} value - Valor a debounce
 * @param {number} delay - Delay en milisegundos (default: DEBOUNCE_DELAY.SEARCH)
 * @returns {*} Valor debounced
 */
export function useDebounce(value, delay = DEBOUNCE_DELAY.SEARCH) {
    const [debouncedValue, setDebouncedValue] = useState(value);

    useEffect(() => {
        // Configurar el timeout
        const handler = setTimeout(() => {
            setDebouncedValue(value);
        }, delay);

        // Limpiar el timeout si el valor cambia antes de que el delay termine
        return () => {
            clearTimeout(handler);
        };
    }, [value, delay]);

    return debouncedValue;
}
