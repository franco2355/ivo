import React from 'react';
import '../styles/Spinner.css';

/**
 * Componente Spinner reutilizable
 * @param {string} size - Tamaño del spinner: 'small', 'medium', 'large' (default: 'medium')
 * @param {string} message - Mensaje opcional a mostrar debajo del spinner
 */
const Spinner = ({ size = 'medium', message = '' }) => {
    return (
        <div className="spinner-container">
            <div className={`spinner spinner-${size}`}></div>
            {message && <p className="spinner-message">{message}</p>}
        </div>
    );
};

export default Spinner;
