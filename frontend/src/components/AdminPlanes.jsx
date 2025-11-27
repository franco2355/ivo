import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import '../styles/AdminPlanes.css';
import { useToastContext } from '../context/ToastContext';
import { handleSessionExpired, isAuthError } from '../utils/auth';
import { SUBSCRIPTIONS_API, SEARCH_API } from '../config/api';

const API_URL = SUBSCRIPTIONS_API.base;

const AdminPlanes = () => {
    const navigate = useNavigate();
    const toast = useToastContext();
    const [planes, setPlanes] = useState([]);
    const [mostrarModal, setMostrarModal] = useState(false);
    const [planEditando, setPlanEditando] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState(null);
    const [paginaActual, setPaginaActual] = useState(1);
    const planesPorPagina = 10;

    // Cargar planes desde la API cuando el componente se monta
    useEffect(() => {
        console.log('[AdminPlanes] Componente montado, cargando planes...');
        cargarPlanes();
    }, []);

    const cargarPlanes = async () => {
        try {
            console.log('[AdminPlanes] Iniciando fetch a:', `${API_URL}/plans`);
            setIsLoading(true);
            setError(null);

            const response = await fetch(`${API_URL}/plans`);
            console.log('[AdminPlanes] Response recibida:', response.status, response.statusText);

            if (!response.ok) {
                throw new Error(`Error HTTP: ${response.status}`);
            }

            const data = await response.json();
            console.log('[AdminPlanes] Data parseada:', data);

            if (data && data.plans && Array.isArray(data.plans)) {
                console.log('[AdminPlanes] ✅ Planes cargados:', data.plans.length);
                console.log('[AdminPlanes] IDs:', data.plans.map(p => p.id));
                setPlanes(data.plans);
            } else {
                console.error('[AdminPlanes] ❌ Formato inválido:', data);
                throw new Error('Formato de respuesta inválido');
            }
        } catch (err) {
            console.error('[AdminPlanes] ❌ Error cargando planes:', err);
            setError(`Error al cargar planes: ${err.message}`);
            setPlanes([]);
        } finally {
            setIsLoading(false);
        }
    };

    const handleAgregar = () => {
        setPlanEditando(null);
        setMostrarModal(true);
    };

    const handleEditar = (plan) => {
        console.log('[AdminPlanes] Editando plan:', plan.id);
        setPlanEditando(plan);
        setMostrarModal(true);
    };

    const handleEliminar = async (planId) => {
        console.log('[AdminPlanes] Intentando eliminar plan:', planId);

        if (!window.confirm('¿Estás seguro de eliminar este plan? Esta acción afectará a las suscripciones existentes.')) {
            return;
        }

        try {
            const token = localStorage.getItem('access_token');
            if (!token) {
                toast.warning('Debes iniciar sesión como administrador');
                return;
            }

            console.log('[AdminPlanes] Enviando DELETE a:', `${API_URL}/plans/${planId}`);
            console.log('[AdminPlanes] Token (primeros 50 chars):', token.substring(0, 50) + '...');

            const response = await fetch(`${API_URL}/plans/${planId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            console.log('[AdminPlanes] DELETE Response:', response.status, response.statusText);

            if (isAuthError(response)) {
                handleSessionExpired(toast, navigate);
                return;
            } else if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                console.error('[AdminPlanes] Error del servidor:', errorData);
                throw new Error(errorData.error || errorData.message || `Error HTTP: ${response.status}`);
            }

            console.log('[AdminPlanes] ✅ Plan eliminado exitosamente');
            toast.success('Plan eliminado exitosamente');
            await cargarPlanes(); // Recargar la lista
        } catch (err) {
            console.error('[AdminPlanes] ❌ Error eliminando plan:', err);
            toast.error(`Error al eliminar el plan: ${err.message}`);
        }
    };

    const handleToggleActivo = async (planId, activoActual) => {
        try {
            const token = localStorage.getItem('access_token');
            if (!token) {
                toast.warning('Debes iniciar sesión como administrador');
                return;
            }

            console.log('[AdminPlanes] Toggle activo para plan:', planId);

            const response = await fetch(`${API_URL}/plans/${planId}/status`, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ activo: !activoActual })
            });

            if (isAuthError(response)) {
                handleSessionExpired(toast, navigate);
                return;
            } else if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `Error HTTP: ${response.status}`);
            }

            console.log('[AdminPlanes] ✅ Estado cambiado');
            await cargarPlanes();
        } catch (err) {
            console.error('[AdminPlanes] ❌ Error cambiando estado:', err);
            toast.error(`Error al cambiar el estado: ${err.message}`);
        }
    };

    const handleGuardarPlan = async (planData, isEdit = false) => {
        try {
            const token = localStorage.getItem('access_token');
            if (!token) {
                toast.warning('Debes iniciar sesión como administrador');
                return;
            }

            const url = isEdit ? `${API_URL}/plans/${planData.id}` : `${API_URL}/plans`;
            const method = isEdit ? 'PUT' : 'POST';

            console.log('[AdminPlanes]', isEdit ? 'Actualizando' : 'Creando', 'plan:', planData);

            const response = await fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(planData)
            });

            if (isAuthError(response)) {
                handleSessionExpired(toast, navigate);
                return;
            } else if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `Error HTTP: ${response.status}`);
            }

            console.log('[AdminPlanes] ✅ Plan guardado');
            toast.success(`Plan ${isEdit ? 'actualizado' : 'creado'} exitosamente`);
            setMostrarModal(false);
            await cargarPlanes();
        } catch (err) {
            console.error('[AdminPlanes] ❌ Error guardando plan:', err);
            toast.error(`Error al ${isEdit ? 'actualizar' : 'crear'} el plan: ${err.message}`);
        }
    };

    // Debug en cada render
    console.log('[AdminPlanes] RENDER - Estado actual:', {
        planesCount: planes.length,
        planesIds: planes.map(p => p.id),
        isLoading,
        error
    });

    return (
        <div className="admin-planes-container">
            <div className="admin-planes-header">
                <h2>Gestión de Planes</h2>
                <button className="btn-agregar-plan" onClick={handleAgregar}>
                    + Agregar Plan
                </button>
                <button
                    className="btn-agregar-plan"
                    onClick={cargarPlanes}
                    style={{ marginLeft: '10px', background: '#2196F3' }}
                >
                    Recargar
                </button>
            </div>

            {error && (
                <div className="alert-warning">
                    {error}
                </div>
            )}

            {!error && planes.length > 0 && (
                <div className="alert-success">
                    {planes.length} plan{planes.length !== 1 ? 'es' : ''} cargado{planes.length !== 1 ? 's' : ''}
                </div>
            )}

            {!error && planes.length === 0 && !isLoading && (
                <div className="alert-warning">
                    No hay planes creados. Agrega un plan para comenzar.
                </div>
            )}

            {isLoading ? (
                <div className="loading-overlay">
                    <div className="spinner"></div>
                    <p>Cargando planes...</p>
                </div>
            ) : planes.length > 0 ? (
                <>
                    <div className="planes-table-container">
                        <table className="planes-table">
                            <thead>
                                <tr>
                                    <th>ID (MongoDB)</th>
                                    <th>Nombre</th>
                                    <th>Precio</th>
                                    <th>Tipo Acceso</th>
                                    <th>Duración</th>
                                    <th>Estado</th>
                                    <th>Acciones</th>
                                </tr>
                            </thead>
                            <tbody>
                                {(() => {
                                    const totalPaginas = Math.ceil(planes.length / planesPorPagina);
                                    const indiceInicio = (paginaActual - 1) * planesPorPagina;
                                    const indiceFin = indiceInicio + planesPorPagina;
                                    const planesPaginados = planes.slice(indiceInicio, indiceFin);

                                    return planesPaginados.map((plan) => (
                                        <tr key={plan.id}>
                                            <td><code style={{ fontSize: '10px' }}>{plan.id}</code></td>
                                            <td>
                                                <div className="plan-nombre-cell">
                                                    <span className="plan-color" style={{ backgroundColor: plan.color || '#4CAF50' }}></span>
                                                    {plan.nombre}
                                                    {plan.popular && <span className="badge-popular">Popular</span>}
                                                </div>
                                            </td>
                                            <td className="precio-cell">${plan.precio_mensual?.toFixed(2)}</td>
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
                                                        onChange={() => handleToggleActivo(plan.id, plan.activo)}
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
                                                    Editar
                                                </button>
                                                <button
                                                    className="btn-icon btn-eliminar"
                                                    onClick={() => handleEliminar(plan.id)}
                                                    title="Eliminar"
                                                >
                                                    Eliminar
                                                </button>
                                            </td>
                                        </tr>
                                    ));
                                })()}
                            </tbody>
                        </table>
                    </div>

                    {/* Controles de paginación */}
                    {Math.ceil(planes.length / planesPorPagina) > 1 && (
                        <div className="paginacion-container">
                            <button
                                className="btn-paginacion"
                                onClick={() => setPaginaActual(prev => Math.max(prev - 1, 1))}
                                disabled={paginaActual === 1}
                            >
                                ← Anterior
                            </button>

                            <div className="paginacion-info">
                                Página {paginaActual} de {Math.ceil(planes.length / planesPorPagina)}
                                <span className="paginacion-total">
                                    ({planes.length} planes)
                                </span>
                            </div>

                            <button
                                className="btn-paginacion"
                                onClick={() => setPaginaActual(prev => Math.min(prev + 1, Math.ceil(planes.length / planesPorPagina)))}
                                disabled={paginaActual === Math.ceil(planes.length / planesPorPagina)}
                            >
                                Siguiente →
                            </button>
                        </div>
                    )}
                </>
            ) : null}

            {mostrarModal && (
                <div className="modal-overlay" onClick={() => setMostrarModal(false)}>
                    <div className="modal-content-plan" onClick={(e) => e.stopPropagation()}>
                        <button className="modal-close" onClick={() => setMostrarModal(false)}>
                            ✕
                        </button>
                        <h3>{planEditando ? 'Editar Plan' : 'Nuevo Plan'}</h3>

                        <FormularioPlan
                            planInicial={planEditando}
                            onGuardar={handleGuardarPlan}
                            onCancelar={() => setMostrarModal(false)}
                            toast={toast}
                        />
                    </div>
                </div>
            )}
        </div>
    );
};

// Componente FormularioPlan
const FormularioPlan = ({ planInicial, onGuardar, onCancelar, toast }) => {
    const isEdit = !!planInicial;
    const [categorias, setCategorias] = useState([]);

    const [formData, setFormData] = useState({
        id: planInicial?.id || null,
        nombre: planInicial?.nombre || '',
        descripcion: planInicial?.descripcion || '',
        precio_mensual: planInicial?.precio_mensual || '',
        duracion_dias: planInicial?.duracion_dias || 30,
        tipo_acceso: planInicial?.tipo_acceso || 'limitado',
        max_clases_semana: planInicial?.max_clases_semana || 0,
        activo: planInicial?.activo !== undefined ? planInicial.activo : true,
        beneficios: planInicial?.beneficios || [],
        color: planInicial?.color || '#4CAF50',
        actividades_permitidas: planInicial?.actividades_permitidas || []
    });

    const [nuevoBeneficio, setNuevoBeneficio] = useState('');

    // Cargar categorías disponibles
    useEffect(() => {
        const fetchCategorias = async () => {
            try {
                const response = await fetch(SEARCH_API.categories);
                if (response.ok) {
                    const data = await response.json();
                    setCategorias(data.categories || []);
                }
            } catch (error) {
                console.error("Error al cargar categorías:", error);
                // Fallback: categorías por defecto
                setCategorias(['yoga', 'spinning', 'funcional', 'pilates', 'crossfit', 'baile', 'boxeo', 'stretching']);
            }
        };
        fetchCategorias();
    }, []);

    const handleSubmit = (e) => {
        e.preventDefault();

        // Validaciones
        if (!formData.nombre || !formData.descripcion) {
            toast.warning('Por favor completa todos los campos requeridos');
            return;
        }

        if (formData.precio_mensual <= 0) {
            toast.warning('El precio debe ser mayor a 0');
            return;
        }

        const planData = {
            ...formData,
            precio_mensual: parseFloat(formData.precio_mensual),
            duracion_dias: parseInt(formData.duracion_dias),
            max_clases_semana: parseInt(formData.max_clases_semana) || null
        };

        // Si estamos editando, incluir el id
        if (isEdit && formData.id) {
            planData.id = formData.id;
        }

        onGuardar(planData, isEdit);
    };

    const agregarBeneficio = () => {
        if (nuevoBeneficio.trim()) {
            setFormData({
                ...formData,
                beneficios: [...formData.beneficios, nuevoBeneficio.trim()]
            });
            setNuevoBeneficio('');
        }
    };

    const eliminarBeneficio = (index) => {
        setFormData({
            ...formData,
            beneficios: formData.beneficios.filter((_, i) => i !== index)
        });
    };

    return (
        <form className="formulario-plan" onSubmit={handleSubmit}>
            <div className="form-group">
                <label>Nombre del Plan *</label>
                <input
                    type="text"
                    value={formData.nombre}
                    onChange={(e) => setFormData({ ...formData, nombre: e.target.value })}
                    placeholder="Ej: Plan Premium"
                    required
                />
            </div>

            <div className="form-group">
                <label>Descripción *</label>
                <textarea
                    value={formData.descripcion}
                    onChange={(e) => setFormData({ ...formData, descripcion: e.target.value })}
                    placeholder="Describe las características del plan"
                    rows="3"
                    required
                />
            </div>

            <div className="form-row">
                <div className="form-group">
                    <label>Precio Mensual *</label>
                    <input
                        type="number"
                        step="0.01"
                        min="0"
                        value={formData.precio_mensual}
                        onChange={(e) => setFormData({ ...formData, precio_mensual: e.target.value })}
                        placeholder="0.00"
                        required
                    />
                </div>

                <div className="form-group">
                    <label>Duración (días)</label>
                    <input
                        type="number"
                        min="1"
                        value={formData.duracion_dias}
                        onChange={(e) => setFormData({ ...formData, duracion_dias: e.target.value })}
                    />
                </div>
            </div>

            <div className="form-row">
                <div className="form-group">
                    <label>Tipo de Acceso</label>
                    <select
                        value={formData.tipo_acceso}
                        onChange={(e) => setFormData({ ...formData, tipo_acceso: e.target.value })}
                    >
                        <option value="limitado">Limitado</option>
                        <option value="completo">Completo</option>
                    </select>
                </div>

                <div className="form-group">
                    <label>Clases por Semana (0 = ilimitado)</label>
                    <input
                        type="number"
                        min="0"
                        value={formData.max_clases_semana}
                        onChange={(e) => setFormData({ ...formData, max_clases_semana: e.target.value })}
                    />
                </div>
            </div>

            <div className="form-group">
                <label>Color del Plan</label>
                <input
                    type="color"
                    value={formData.color}
                    onChange={(e) => setFormData({ ...formData, color: e.target.value })}
                />
            </div>

            {/* Campo para actividades permitidas - solo visible si el plan es "limitado" */}
            {formData.tipo_acceso === 'limitado' && (
                <div className="form-group">
                    <label>
                        Actividades Permitidas
                        <span style={{ fontSize: '0.85em', color: '#666', marginLeft: '8px' }}>
                            (Selecciona las categorías que este plan puede acceder)
                        </span>
                    </label>
                    <div className="categorias-selector">
                        {categorias.map((categoria) => {
                            const categoriaLower = categoria.toLowerCase();
                            const isSelected = formData.actividades_permitidas.includes(categoriaLower);
                            return (
                                <label key={categoria} className={`categoria-checkbox ${isSelected ? 'selected' : ''}`}>
                                    <input
                                        type="checkbox"
                                        checked={isSelected}
                                        onChange={(e) => {
                                            const newActividades = e.target.checked
                                                ? [...formData.actividades_permitidas, categoriaLower]
                                                : formData.actividades_permitidas.filter(a => a !== categoriaLower);
                                            setFormData({ ...formData, actividades_permitidas: newActividades });
                                        }}
                                    />
                                    <span>{categoria.charAt(0).toUpperCase() + categoria.slice(1)}</span>
                                </label>
                            );
                        })}
                    </div>
                    {formData.actividades_permitidas.length === 0 && (
                        <p style={{ color: '#ff9800', fontSize: '0.9em', marginTop: '8px' }}>
                            ⚠️ Sin actividades seleccionadas, el usuario no podrá inscribirse en ninguna actividad
                        </p>
                    )}
                </div>
            )}

            <div className="form-group">
                <label>Beneficios</label>
                <div className="beneficios-input">
                    <input
                        type="text"
                        value={nuevoBeneficio}
                        onChange={(e) => setNuevoBeneficio(e.target.value)}
                        placeholder="Agregar beneficio"
                        onKeyPress={(e) => {
                            if (e.key === 'Enter') {
                                e.preventDefault();
                                agregarBeneficio();
                            }
                        }}
                    />
                    <button type="button" onClick={agregarBeneficio}>+</button>
                </div>
                <ul className="beneficios-lista">
                    {formData.beneficios.map((beneficio, index) => (
                        <li key={`beneficio-${index}-${beneficio.substring(0, 10)}`}>
                            {beneficio}
                            <button type="button" onClick={() => eliminarBeneficio(index)}>×</button>
                        </li>
                    ))}
                </ul>
            </div>

            <div className="form-group-checkbox">
                <label>
                    <input
                        type="checkbox"
                        checked={formData.activo}
                        onChange={(e) => setFormData({ ...formData, activo: e.target.checked })}
                    />
                    Plan activo
                </label>
            </div>

            <div className="form-actions">
                <button type="button" className="btn-cancelar" onClick={onCancelar}>
                    Cancelar
                </button>
                <button type="submit" className="btn-guardar">
                    {isEdit ? 'Actualizar Plan' : 'Crear Plan'}
                </button>
            </div>
        </form>
    );
};

export default AdminPlanes;
