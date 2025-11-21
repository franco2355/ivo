import { useState, useCallback } from 'react';

export const useToast = () => {
    const [toasts, setToasts] = useState([]);

    const showToast = useCallback((message, type = 'info', duration = 4000) => {
        const id = Date.now();
        const newToast = { id, message, type, duration };

        setToasts(prev => [...prev, newToast]);

        setTimeout(() => {
            setToasts(prev => prev.filter(toast => toast.id !== id));
        }, duration + 300); // Extra time for animation
    }, []);

    const success = useCallback((message, duration) => {
        showToast(message, 'success', duration);
    }, [showToast]);

    const error = useCallback((message, duration) => {
        showToast(message, 'error', duration);
    }, [showToast]);

    const warning = useCallback((message, duration) => {
        showToast(message, 'warning', duration);
    }, [showToast]);

    const info = useCallback((message, duration) => {
        showToast(message, 'info', duration);
    }, [showToast]);

    const removeToast = useCallback((id) => {
        setToasts(prev => prev.filter(toast => toast.id !== id));
    }, []);

    return {
        toasts,
        showToast,
        success,
        error,
        warning,
        info,
        removeToast
    };
};
