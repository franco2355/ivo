import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { SUBSCRIPTIONS_API } from '../config/api';
import '../styles/MiSuscripcion.css';
import { useToastContext } from '../context/ToastContext';
import { handleSessionExpired, isAuthError } from '../utils/auth';

const MiSuscripcion = () => {
    const [suscripcion, setSuscripcion] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const navigate = useNavigate();
    const toast = useToastContext();

    useEffect(() => {
        const controller = new AbortController();

        const loadData = async () => {
            try {
                setLoading(true);
                setError(null);
                const token = localStorage.getItem('access_token');
                const userId = localStorage.getItem('idUsuario');

                if (!userId) {
                    setError('Usuario no identificado');
                    toast.error('Debes iniciar sesión para ver tu suscripción');
                    navigate('/login');
                    return;
                }

                console.log('[MiSuscripcion] Cargando suscripción activa del usuario:', userId);

                const response = await fetch(SUBSCRIPTIONS_API.activeSubscription(userId), {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    },
                    signal: controller.signal
                });

                console.log('[MiSuscripcion] Response:', response.status);

                if (isAuthError(response)) {
                    handleSessionExpired(toast, navigate);
                    return;
                }

                if (response.status === 404) {
                    console.log('[MiSuscripcion] No se encontró suscripción activa');
                    setSuscripcion(null);
                    return;
                }

                if (!response.ok) {
                    throw new Error(`Error ${response.status}: ${response.statusText}`);
                }

                const data = await response.json();
                console.log('[MiSuscripcion] Suscripción activa recibida:', data);

                if (data && data.estado === 'activa') {
                    setSuscripcion(data);
                } else {
                    setSuscripcion(null);
                }
            } catch (error) {
                if (error.name === 'AbortError') {
                    console.log('[MiSuscripcion] Request cancelado');
                    return;
                }
                console.error("[MiSuscripcion] Error al cargar suscripción:", error);
                setError(error.message || 'Error al cargar la suscripción');
                toast.error('No se pudo cargar tu suscripción. Intenta nuevamente.');
                setSuscripcion(null);
            } finally {
                setLoading(false);
            }
        };

        loadData();

        // Cleanup: cancelar la request si el componente se desmonta
        return () => {
            controller.abort();
        };
    }, []);

    const fetchSuscripcionActiva = async () => {
        try {
            setLoading(true);
            setError(null);
            const token = localStorage.getItem('access_token');
            const userId = localStorage.getItem('idUsuario');

            if (!userId) {
                setError('Usuario no identificado');
                toast.error('Debes iniciar sesión para ver tu suscripción');
                navigate('/login');
                return;
            }

            console.log('[MiSuscripcion] Recargando suscripción activa del usuario:', userId);

            const response = await fetch(SUBSCRIPTIONS_API.activeSubscription(userId), {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            console.log('[MiSuscripcion] Response:', response.status);

            if (isAuthError(response)) {
                handleSessionExpired(toast, navigate);
                return;
            }

            if (response.status === 404) {
                // No tiene suscripción activa
                console.log('[MiSuscripcion] No se encontró suscripción activa');
                setSuscripcion(null);
                return;
            }

            if (!response.ok) {
                throw new Error(`Error ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            console.log('[MiSuscripcion] Suscripción activa recibida:', data);

            if (data && data.estado === 'activa') {
                setSuscripcion(data);
            } else {
                setSuscripcion(null);
            }
        } catch (error) {
            console.error("[MiSuscripcion] Error al cargar suscripción:", error);
            setError(error.message || 'Error al cargar la suscripción');
            toast.error('No se pudo cargar tu suscripción. Intenta nuevamente.');
            setSuscripcion(null);
        } finally {
            setLoading(false);
        }
    };

    const handleCancelar = async (subscriptionId) => {
        if (!window.confirm("¿Estás seguro de que deseas cancelar esta suscripción?")) {
            return;
        }

        try {
            const token = localStorage.getItem('access_token');
            const response = await fetch(SUBSCRIPTIONS_API.subscriptionById(subscriptionId), {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (isAuthError(response)) {
                handleSessionExpired(toast, navigate);
                return;
            }

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || 'Error al cancelar la suscripción');
            }

            toast.success("Suscripción cancelada exitosamente");
            // Recargar la suscripción activa
            await fetchSuscripcionActiva();
        } catch (error) {
            console.error("Error al cancelar suscripción:", error);
            toast.error(error.message || "Error al cancelar la suscripción");
        }
    };

    const handleRenovar = (planId) => {
        if (planId) {
            navigate(`/checkout/${planId}`);
        }
    };

    const getDiasRestantes = (fechaVencimiento) => {
        const hoy = new Date();
        const vencimiento = new Date(fechaVencimiento);
        const diffTime = vencimiento - hoy;
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
        return diffDays;
    };

    const getEstadoBadgeClass = (estado) => {
        switch (estado) {
            case 'activa':
                return 'estado-activa';
            case 'vencida':
                return 'estado-vencida';
            case 'cancelada':
                return 'estado-cancelada';
            case 'pendiente_pago':
                return 'estado-pendiente';
            default:
                return '';
        }
    };

    if (loading) {
        return (
            <div className="mi-suscripcion-container">
                <div className="loading-message">Cargando suscripción...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="mi-suscripcion-container">
                <div className="error-message">
                    <div className="error-icon">⚠️</div>
                    <h2>Error al cargar la suscripción</h2>
                    <p>{error}</p>
                    <button
                        className="btn-reintentar"
                        onClick={fetchSuscripcionActiva}
                    >
                        Reintentar
                    </button>
                </div>
            </div>
        );
    }

    if (!suscripcion) {
        return (
            <div className="mi-suscripcion-container">
                <div className="no-suscripcion">
                    <div className="no-suscripcion-icon">📋</div>
                    <h2>No tenés una suscripción activa</h2>
                    <p>Suscribite a uno de nuestros planes para acceder a todas las actividades</p>
                    <button
                        className="btn-ver-planes"
                        onClick={() => navigate('/planes')}
                    >
                        Ver Planes Disponibles
                    </button>
                </div>
            </div>
        );
    }

    const diasRestantes = getDiasRestantes(suscripcion.fecha_vencimiento);
    const proximoVencer = diasRestantes <= 7 && diasRestantes > 0;

    return (
        <div className="mi-suscripcion-container">
            <div className="suscripcion-header">
                <h1>Mi Suscripción</h1>
                <p>Suscripción Activa</p>
            </div>

            <div className="suscripcion-content">
                <div className="suscripcion-card">
                    <div className="suscripcion-plan-header" style={{ borderTopColor: suscripcion.plan?.color }}>
                        <div className="plan-info">
                            <h2>{suscripcion.plan?.nombre || "Plan"}</h2>
                            <p className="plan-descripcion">{suscripcion.plan?.descripcion}</p>
                        </div>
                        <div className={`estado-badge ${getEstadoBadgeClass(suscripcion.estado)}`}>
                            {suscripcion.estado.toUpperCase()}
                        </div>
                    </div>

                    <div className="suscripcion-detalles">
                        <div className="detalle-item">
                            <span className="detalle-label">Precio Mensual:</span>
                            <span className="detalle-valor precio">${suscripcion.plan?.precio_mensual.toFixed(2)}</span>
                        </div>
                        <div className="detalle-item">
                            <span className="detalle-label">Fecha de Inicio:</span>
                            <span className="detalle-valor">
                                {new Date(suscripcion.fecha_inicio).toLocaleDateString('es-AR')}
                            </span>
                        </div>
                        <div className="detalle-item">
                            <span className="detalle-label">Fecha de Vencimiento:</span>
                            <span className="detalle-valor">
                                {new Date(suscripcion.fecha_vencimiento).toLocaleDateString('es-AR')}
                            </span>
                        </div>
                        <div className="detalle-item">
                            <span className="detalle-label">Días Restantes:</span>
                            <span className={`detalle-valor ${proximoVencer ? 'dias-advertencia' : ''}`}>
                                {diasRestantes > 0 ? `${diasRestantes} días` : 'Vencida'}
                            </span>
                        </div>
                        {suscripcion.metadata?.auto_renovacion && (
                            <div className="detalle-item">
                                <span className="detalle-label">Auto-renovación:</span>
                                <span className="detalle-valor renovacion-activa">Activada ✓</span>
                            </div>
                        )}
                    </div>

                    {proximoVencer && (
                        <div className="alerta-vencimiento">
                            ⚠️ Tu suscripción vence pronto. Renovála para seguir disfrutando de los beneficios.
                        </div>
                    )}

                    <div className="suscripcion-beneficios">
                        <h3>Beneficios de tu plan:</h3>
                        <ul>
                            {suscripcion.plan?.beneficios?.map((beneficio, index) => (
                                <li key={index}>
                                    <span className="check-icon">✓</span>
                                    {beneficio}
                                </li>
                            ))}
                        </ul>
                    </div>

                    <div className="suscripcion-acciones">
                        {suscripcion.estado === 'activa' && (
                            <>
                                <button className="btn-renovar" onClick={() => handleRenovar(suscripcion.plan_id)}>
                                    Renovar Suscripción
                                </button>
                                <button className="btn-cancelar" onClick={() => handleCancelar(suscripcion.id)}>
                                    Cancelar Suscripción
                                </button>
                            </>
                        )}
                        {(suscripcion.estado === 'vencida' || suscripcion.estado === 'cancelada') && (
                            <button className="btn-renovar-principal" onClick={() => navigate('/planes')}>
                                Ver Planes Disponibles
                            </button>
                        )}
                        {suscripcion.estado === 'pendiente_pago' && (
                            <button className="btn-pagar" onClick={() => navigate('/pagos')}>
                                Completar Pago
                            </button>
                        )}
                    </div>

                    {suscripcion.historial_renovaciones && suscripcion.historial_renovaciones.length > 0 && (
                        <div className="historial-renovaciones">
                            <h3>Historial de Renovaciones</h3>
                            <div className="renovaciones-lista">
                                {suscripcion.historial_renovaciones.map((renovacion, index) => (
                                    <div key={index} className="renovacion-item">
                                        <div className="renovacion-fecha">
                                            {new Date(renovacion.fecha).toLocaleDateString('es-AR')}
                                        </div>
                                        <div className="renovacion-monto">${renovacion.monto.toFixed(2)}</div>
                                        <div className="renovacion-pago">ID: {renovacion.pago_id}</div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default MiSuscripcion;
