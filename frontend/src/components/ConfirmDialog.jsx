import React from 'react';
import '../styles/ConfirmDialog.css';

/**
 * Componente de diálogo de confirmación reutilizable
 * @param {boolean} isOpen - Si el diálogo está abierto
 * @param {string} title - Título del diálogo
 * @param {string} message - Mensaje del diálogo
 * @param {function} onConfirm - Callback al confirmar
 * @param {function} onCancel - Callback al cancelar
 * @param {string} confirmText - Texto del botón de confirmar (default: "Confirmar")
 * @param {string} cancelText - Texto del botón de cancelar (default: "Cancelar")
 * @param {string} variant - Variante de color: 'danger', 'warning', 'info' (default: 'danger')
 */
const ConfirmDialog = ({
    isOpen,
    title = '¿Estás seguro?',
    message,
    onConfirm,
    onCancel,
    confirmText = 'Confirmar',
    cancelText = 'Cancelar',
    variant = 'danger'
}) => {
    if (!isOpen) return null;

    return (
        <div className="confirm-dialog-overlay" onClick={onCancel}>
            <div className="confirm-dialog-container" onClick={(e) => e.stopPropagation()}>
                <div className="confirm-dialog-header">
                    <h3 className="confirm-dialog-title">{title}</h3>
                </div>
                <div className="confirm-dialog-body">
                    <p className="confirm-dialog-message">{message}</p>
                </div>
                <div className="confirm-dialog-footer">
                    <button
                        onClick={onCancel}
                        className="confirm-dialog-button confirm-dialog-button-cancel"
                    >
                        {cancelText}
                    </button>
                    <button
                        onClick={onConfirm}
                        className={`confirm-dialog-button confirm-dialog-button-confirm confirm-dialog-button-${variant}`}
                    >
                        {confirmText}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default ConfirmDialog;
