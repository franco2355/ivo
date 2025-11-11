import { useState } from 'react';
import { mockPlanes } from '../data/mockData';
import '../styles/AdminPlanes.css';

const AdminPlanes = () => {
    const [planes, setPlanes] = useState(mockPlanes);
    const [mostrarModal, setMostrarModal] = useState(false);
    const [planEditando, setPlanEditando] = useState(null);

    const handleAgregar = () => {
        setPlanEditando(null);
        setMostrarModal(true);
    };

    const handleEditar = (plan) => {
        setPlanEditando(plan);
        setMostrarModal(true);
    };

    const handleEliminar = (planId) => {
        if (window.confirm('¿Estás seguro de eliminar este plan? Esta acción afectará a las suscripciones existentes.')) {
            // Aquí iría la llamada a la API cuando esté lista
            alert('Funcionalidad disponible cuando subscriptions-api esté lista');
        }
    };

    const handleToggleActivo = (planId) => {
        // Aquí iría la llamada a la API cuando esté lista
        alert('Funcionalidad disponible cuando subscriptions-api esté lista');
    };

    return (
        <div className="admin-planes-container">
            <div className="admin-planes-header">
                <h2>Gestión de Planes</h2>
                <button className="btn-agregar-plan" onClick={handleAgregar}>
                    + Agregar Plan
                </button>
            </div>

            <div className="alert-info">
                ℹ️ Los planes se gestionarán con la API real cuando subscriptions-api esté disponible
            </div>

            <div className="planes-table-container">
                <table className="planes-table">
                    <thead>
                        <tr>
                            <th>Nombre</th>
                            <th>Precio</th>
                            <th>Tipo Acceso</th>
                            <th>Duración</th>
                            <th>Estado</th>
                            <th>Acciones</th>
                        </tr>
                    </thead>
                    <tbody>
                        {planes.map((plan) => (
                            <tr key={plan.id}>
                                <td>
                                    <div className="plan-nombre-cell">
                                        <span className="plan-color" style={{ backgroundColor: plan.color }}></span>
                                        {plan.nombre}
                                        {plan.popular && <span className="badge-popular">Popular</span>}
                                    </div>
                                </td>
                                <td className="precio-cell">${plan.precio_mensual.toFixed(2)}</td>
                                <td>
                                    <span className={`badge-acceso ${plan.tipo_acceso}`}>
                                        {plan.tipo_acceso}
                                    </span>
                                </td>
                                <td>{plan.duracion_dias} días</td>
                                <td>
                                    <label className="toggle-switch">
                                        <input
                                            type="checkbox"
                                            checked={plan.activo}
                                            onChange={() => handleToggleActivo(plan.id)}
                                        />
                                        <span className="toggle-slider-small"></span>
                                        <span className="toggle-label-small">
                                            {plan.activo ? 'Activo' : 'Inactivo'}
                                        </span>
                                    </label>
                                </td>
                                <td className="acciones-cell">
                                    <button
                                        className="btn-icon btn-editar"
                                        onClick={() => handleEditar(plan)}
                                        title="Editar"
                                    >
                                        ✏️
                                    </button>
                                    <button
                                        className="btn-icon btn-eliminar"
                                        onClick={() => handleEliminar(plan.id)}
                                        title="Eliminar"
                                    >
                                        🗑️
                                    </button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {mostrarModal && (
                <div className="modal-overlay" onClick={() => setMostrarModal(false)}>
                    <div className="modal-content-plan" onClick={(e) => e.stopPropagation()}>
                        <button className="modal-close" onClick={() => setMostrarModal(false)}>
                            ✕
                        </button>
                        <h3>{planEditando ? 'Editar Plan' : 'Nuevo Plan'}</h3>
                        <div className="alert-info">
                            Esta funcionalidad estará disponible cuando subscriptions-api esté lista
                        </div>
                        <p>Por ahora solo se pueden visualizar los planes mock.</p>
                    </div>
                </div>
            )}
        </div>
    );
};

export default AdminPlanes;
