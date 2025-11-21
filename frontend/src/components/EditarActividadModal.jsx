import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useToastContext } from '../context/ToastContext';
import { handleSessionExpired, isAuthError } from '../utils/auth';
import { USERS_API } from '../config/api';
import '../styles/EditarActividadModal.css';

const EditarActividadModal = ({ actividad, onClose, onSave }) => {
    const navigate = useNavigate();
    const toast = useToastContext();
    const [formData, setFormData] = useState({
        id_actividad: '',
        titulo: '',
        descripcion: '',
        cupo: 0,
        dia: '',
        hora_inicio: '',
        hora_fin: '',
        foto_url: '',
        instructor: '',
        categoria: ''
    });
    const [error, setError] = useState('');
    const [validationErrors, setValidationErrors] = useState({});

    useEffect(() => {
        if (actividad) {
            // Asegurarse de que todos los campos necesarios estén presentes y que el cupo sea un número
            const actividadData = {
                id_actividad: actividad.id_actividad,
                titulo: actividad.titulo || '',
                descripcion: actividad.descripcion || '',
                cupo: parseInt(actividad.cupo, 10) || 0,
                dia: actividad.dia || '',
                hora_inicio: actividad.hora_inicio || '',
                hora_fin: actividad.hora_fin || '',
                foto_url: actividad.foto_url || '',
                instructor: actividad.instructor || '',
                categoria: actividad.categoria || ''
            };
            setFormData(actividadData);
        }
    }, [actividad]);

    const validateForm = () => {
        const errors = {};
        
        if (!formData.titulo.trim()) {
            errors.titulo = 'El título es requerido';
        } else if (formData.titulo.length < 3) {
            errors.titulo = 'El título debe tener al menos 3 caracteres';
        }

        if (!formData.descripcion.trim()) {
            errors.descripcion = 'La descripción es requerida';
        }

        if (!formData.cupo || formData.cupo < 1) {
            errors.cupo = 'El cupo debe ser mayor a 0';
        }

        if (!formData.dia) {
            errors.dia = 'El día es requerido';
        }

        if (!formData.hora_inicio) {
            errors.hora_inicio = 'La hora de inicio es requerida';
        }

        if (!formData.hora_fin) {
            errors.hora_fin = 'La hora de fin es requerida';
        } else if (formData.hora_fin <= formData.hora_inicio) {
            errors.hora_fin = 'La hora de fin debe ser posterior a la hora de inicio';
        }

        if (!formData.instructor.trim()) {
            errors.instructor = 'El instructor es requerido';
        }

        if (!formData.categoria.trim()) {
            errors.categoria = 'La categoría es requerida';
        }

        setValidationErrors(errors);
        return Object.keys(errors).length === 0;
    };

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: name === 'cupo' ? parseInt(value, 10) || 0 : value
        }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!validateForm()) {
            return;
        }

        try {
            // Asegurarse de que el cupo sea un número antes de enviar
            const dataToSend = {
                ...formData,
                cupo: parseInt(formData.cupo, 10)
            };

            const response = await fetch(`${USERS_API.base}/actividades/${formData.id_actividad}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('access_token')}`
                },
                body: JSON.stringify(dataToSend)
            });

            if (isAuthError(response)) {
                handleSessionExpired(toast, navigate);
            } else if (response.ok) {
                onSave();
                onClose();
            } else {
                const errorData = await response.json();
                setError(errorData.error || 'Error al actualizar la actividad');
            }
        } catch (error) {
            setError('Error al conectar con el servidor');
        }
    };

    return (
        <div className="modal-overlay">
            <div className="modal-content">
                <h2>Editar Actividad</h2>
                {error && <div className="error-message">{error}</div>}
                <form onSubmit={handleSubmit}>
                    <div className="form-group">
                        <label htmlFor="titulo">Título:</label>
                        <input
                            type="text"
                            id="titulo"
                            name="titulo"
                            value={formData.titulo}
                            onChange={handleChange}
                            required
                        />
                        {validationErrors.titulo && <span className="error-text">{validationErrors.titulo}</span>}
                    </div>

                    <div className="form-group">
                        <label htmlFor="descripcion">Descripción:</label>
                        <textarea
                            id="descripcion"
                            name="descripcion"
                            value={formData.descripcion}
                            onChange={handleChange}
                            required
                        />
                        {validationErrors.descripcion && <span className="error-text">{validationErrors.descripcion}</span>}
                    </div>

                    <div className="form-group">
                        <label htmlFor="cupo">Cupo: ({actividad.cupo - actividad.lugares} inscriptos)</label>
                        <input
                            type="number"
                            id="cupo"
                            name="cupo"
                            value={formData.cupo}
                            onChange={handleChange}
                            required
                            min="1"
                        />
                        {validationErrors.cupo && <span className="error-text">{validationErrors.cupo}</span>}
                    </div>

                    <div className="form-group">
                        <label htmlFor="dia">Día:</label>
                        <select
                            id="dia"
                            name="dia"
                            value={formData.dia}
                            onChange={handleChange}
                            required
                        >
                            <option value="">Seleccione un día</option>
                            <option value="Lunes">Lunes</option>
                            <option value="Martes">Martes</option>
                            <option value="Miercoles">Miercoles</option>
                            <option value="Jueves">Jueves</option>
                            <option value="Viernes">Viernes</option>
                            <option value="Sabado">Sabado</option>
                            <option value="Domingo">Domingo</option>
                        </select>
                        {validationErrors.dia && <span className="error-text">{validationErrors.dia}</span>}
                    </div>

                    <div className="form-group">
                        <label htmlFor="hora_inicio">Hora de inicio:</label>
                        <input
                            type="time"
                            id="hora_inicio"
                            name="hora_inicio"
                            value={formData.hora_inicio}
                            onChange={handleChange}
                            required
                        />
                        {validationErrors.hora_inicio && <span className="error-text">{validationErrors.hora_inicio}</span>}
                    </div>

                    <div className="form-group">
                        <label htmlFor="hora_fin">Hora de fin:</label>
                        <input
                            type="time"
                            id="hora_fin"
                            name="hora_fin"
                            value={formData.hora_fin}
                            onChange={handleChange}
                            required
                        />
                        {validationErrors.hora_fin && <span className="error-text">{validationErrors.hora_fin}</span>}
                    </div>

                    <div className="form-group">
                        <label htmlFor="foto_url">URL de la foto:</label>
                        <input
                            type="text"
                            id="foto_url"
                            name="foto_url"
                            value={formData.foto_url}
                            onChange={handleChange}
                        />
                        {validationErrors.foto_url && <span className="error-text">{validationErrors.foto_url}</span>}
                    </div>

                    <div className="form-group">
                        <label htmlFor="instructor">Instructor:</label>
                        <input
                            type="text"
                            id="instructor"
                            name="instructor"
                            value={formData.instructor}
                            placeholder="Instructor..."
                            onChange={handleChange}
                            required
                        />
                        {validationErrors.instructor && <span className="error-text">{validationErrors.instructor}</span>}
                    </div>

                    <div className="form-group">
                        <label htmlFor="categoria">Categoría:</label>
                        <input
                            type="text"
                            id="categoria"
                            name="categoria"
                            value={formData.categoria}
                            onChange={handleChange}
                            placeholder="Ej: Musculación, Cardio, Yoga..."
                            required
                        />
                        {validationErrors.categoria && <span className="error-text">{validationErrors.categoria}</span>}
                    </div>

                    <div className="form-buttons">
                        <button type="submit" className="btn-guardar">Guardar Cambios</button>
                        <button type="button" className="btn-cancelar" onClick={onClose}>
                            Cancelar
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default EditarActividadModal; 