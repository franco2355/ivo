import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import '../styles/Planes.css';
import { useToastContext } from '../context/ToastContext';
import { SUBSCRIPTIONS_API } from '../config/api';

const Planes = () => {
    const [planes, setPlanes] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [activeSuscripcion, setActiveSuscripcion] = useState(null);
    const navigate = useNavigate();
    const isLoggedIn = localStorage.getItem("isLoggedIn") === "true";
    const isAdmin = localStorage.getItem("isAdmin") === "true";
    const userId = localStorage.getItem("idUsuario");
    const toast = useToastContext();

    useEffect(() => {
        console.log('[Planes] Componente montado, cargando datos...');
        cargarDatos();
    }, []);


    const cargarDatos = async () => {
        // Primero cargar la suscripción activa si el usuario está logueado
        if (isLoggedIn && userId) {
            await cargarSuscripcionActiva();
        }
        // Luego cargar los planes
        await cargarPlanes();
    };

    const cargarSuscripcionActiva = async () => {
        try {
            const token = localStorage.getItem('access_token');
            const headers = token ? { 'Authorization': `Bearer ${token}` } : {};

            // Primero intentar obtener suscripción activa
            const response = await fetch(SUBSCRIPTIONS_API.activeSubscription(userId), { headers });
            if (response.ok) {
                const data = await response.json();
                if (data && data.estado === 'activa') {
                    setActiveSuscripcion(data);
                    console.log('[Planes] Suscripción activa encontrada:', data);
                    return;
                }
            }

            // Si no hay activa, buscar si hay pendiente de pago
            const allResponse = await fetch(SUBSCRIPTIONS_API.subscriptionsByUser(userId), { headers });
            if (allResponse.ok) {
                const allData = await allResponse.json();
                if (allData && Array.isArray(allData)) {
                    const pendiente = allData.find(s => s.estado === 'pendiente_pago');
                    if (pendiente) {
                        setActiveSuscripcion(pendiente);
                        console.log('[Planes] Suscripción pendiente de pago encontrada:', pendiente);
                    }
                }
            }
        } catch (error) {
            console.error('[Planes] Error al cargar suscripción activa:', error);
        }
    };

    const cargarPlanes = async () => {
        try {
            console.log('[Planes] Cargando planes desde Subscriptions API...');
            setLoading(true);
            setError(null);

            const response = await fetch(SUBSCRIPTIONS_API.plans);
            console.log('[Planes] Response:', response.status);

            if (!response.ok) {
                throw new Error('Error al cargar planes');
            }

            const data = await response.json();
            console.log('[Planes] Data recibida:', data);

            // Subscriptions API devuelve { plans: [...], total, page, ... }
            if (data && Array.isArray(data.plans)) {
                console.log('[Planes] Planes cargados:', data.plans.length);

                // Filtrar solo planes activos (mostrar todos, incluso el que tiene activo)
                let planesActivos = data.plans.filter(plan => plan.activo === true);
                console.log('[Planes] Planes activos:', planesActivos.length);

                setPlanes(planesActivos);
            } else if (data.plans === null || data.total === 0) {
                // No hay planes, pero no es un error
                console.warn('[Planes] No hay planes disponibles');
                setPlanes([]);
            } else {
                throw new Error('Formato de respuesta inválido');
            }
        } catch (err) {
            console.error('[Planes] ❌ Error:', err);
            setError('No se pudieron cargar los planes. Por favor, intenta más tarde.');
            setPlanes([]);
        } finally {
            setLoading(false);
        }
    };

    const handleSelectPlan = async (planId) => {
        console.log('[Planes] Usuario seleccionó plan:', planId);

        if (isAdmin) {
            toast.info("Los administradores no pueden suscribirse a planes");
            return;
        }

        if (!isLoggedIn) {
            toast.warning("Debes iniciar sesión para suscribirte a un plan");
            navigate('/login');
            return;
        }

        // Verificar si ya tiene este plan activo o pendiente
        if (activeSuscripcion && activeSuscripcion.plan_id === planId) {
            if (activeSuscripcion.estado === 'pendiente_pago') {
                // Redirigir a Mi Suscripción para completar el pago
                navigate('/mi-suscripcion');
            } else {
                toast.info(`Ya tienes una suscripción activa al ${activeSuscripcion.plan_nombre}`);
            }
            return;
        }

        // Navegar al checkout con el plan seleccionado
        // La confirmación de cancelación del plan actual se hace en el Checkout al momento del pago
        navigate(`/checkout/${planId}`);
    };

    if (loading) {
        return (
            <div className="planes-container">
                <div className="loading-message">
                    <div className="spinner"></div>
                    <p>Cargando planes disponibles...</p>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="planes-container">
                <div className="planes-header">
                    <h1>Nuestros Planes</h1>
                </div>
                <div className="error-message">
                    <p>{error}</p>
                    <button className="btn-retry" onClick={cargarPlanes}>
                        Reintentar
                    </button>
                </div>
            </div>
        );
    }

    if (planes.length === 0) {
        return (
            <div className="planes-container">
                <div className="planes-header">
                    <h1>Nuestros Planes</h1>
                </div>
                <div className="no-planes-message">
                    <p>No hay planes disponibles en este momento.</p>
                    <p>Por favor, volvé más tarde o contactanos para más información.</p>
                </div>
            </div>
        );
    }

    return (
        <div className="planes-container">
            <div className="planes-header">
                <h1>Nuestros Planes</h1>
                <p>Elegí el plan que mejor se adapte a tus necesidades</p>
            </div>

            <div className="planes-grid">
                {planes.map((plan) => (
                    <div
                        key={plan.id}
                        className={`plan-card ${plan.popular ? 'popular' : ''} ${activeSuscripcion && activeSuscripcion.plan_id === plan.id ? 'plan-actual' : ''}`}
                    >
                        {activeSuscripcion && activeSuscripcion.plan_id === plan.id && (
                            <div className={`plan-actual-badge ${activeSuscripcion.estado === 'pendiente_pago' ? 'pendiente' : ''}`}>
                                {activeSuscripcion.estado === 'pendiente_pago' ? 'Pendiente de Pago' : 'Tu Plan Actual'}
                            </div>
                        )}
                        {plan.popular && <div className="popular-badge">Más Popular</div>}
                        {plan.ahorro && <div className="ahorro-badge">{plan.ahorro}</div>}

                        <div className="plan-header" style={{ borderTopColor: plan.color || '#4CAF50' }}>
                            <h2>{plan.nombre}</h2>
                            <p className="plan-descripcion">{plan.descripcion}</p>
                        </div>

                        <div className="plan-precio">
                            <span className="precio-signo">$</span>
                            <span className="precio-monto">{plan.precio_mensual?.toFixed(2)}</span>
                            <span className="precio-periodo">/ mes</span>
                        </div>

                        {plan.duracion_dias && plan.duracion_dias > 30 && (
                            <div className="plan-duracion">
                                Compromiso de {plan.duracion_dias} días
                            </div>
                        )}

                        {plan.tipo_acceso && (
                            <div className="plan-acceso">
                                <strong>Tipo de acceso:</strong> {plan.tipo_acceso === 'completo' ? 'Completo' : 'Limitado'}
                            </div>
                        )}

                        {plan.max_clases_semana && plan.max_clases_semana > 0 && (
                            <div className="plan-clases">
                                <strong>Clases por semana:</strong> {plan.max_clases_semana}
                            </div>
                        )}

                        <div className="plan-beneficios">
                            <h3>Beneficios incluidos:</h3>
                            <ul>
                                {plan.beneficios && plan.beneficios.length > 0 ? (
                                    plan.beneficios.map((beneficio, index) => (
                                        <li key={index}>
                                            <span className="check-icon">✓</span>
                                            {beneficio}
                                        </li>
                                    ))
                                ) : (
                                    <li>
                                        <span className="check-icon">✓</span>
                                        Acceso al gimnasio por {plan.duracion_dias} días
                                    </li>
                                )}
                            </ul>
                        </div>

                        {isAdmin ? (
                            <button
                                className="btn-seleccionar-plan"
                                style={{
                                    backgroundColor: '#999',
                                    cursor: 'not-allowed',
                                    opacity: 0.6
                                }}
                                disabled
                            >
                                Solo para usuarios
                            </button>
                        ) : activeSuscripcion && activeSuscripcion.plan_id === plan.id ? (
                            activeSuscripcion.estado === 'pendiente_pago' ? (
                                <button
                                    className="btn-seleccionar-plan btn-completar-pago"
                                    onClick={() => navigate('/mi-suscripcion')}
                                    title="Ir a completar el pago"
                                >
                                    Completar Pago
                                </button>
                            ) : (
                                <button
                                    className="btn-seleccionar-plan btn-renovar"
                                    onClick={() => navigate(`/checkout/${plan.id}`)}
                                    title="Renovar tu suscripción actual"
                                >
                                    Renovar Suscripción
                                </button>
                            )
                        ) : activeSuscripcion ? (
                            <button
                                className="btn-seleccionar-plan btn-cambiar-plan"
                                style={{
                                    backgroundColor: '#FF9800',
                                    border: '2px solid #F57C00'
                                }}
                                onClick={() => handleSelectPlan(plan.id)}
                                title="Cambiar a este plan"
                            >
                                Cambiar de plan
                            </button>
                        ) : (
                            <button
                                className="btn-seleccionar-plan"
                                style={{ backgroundColor: plan.color || '#4CAF50' }}
                                onClick={() => handleSelectPlan(plan.id)}
                            >
                                Seleccionar Plan
                            </button>
                        )}
                    </div>
                ))}
            </div>

            <div className="planes-footer">
                <div className="info-adicional">
                    <h3>¿Necesitás ayuda para elegir?</h3>
                    <p>Contactanos al <strong>0351-123-4567</strong> o visitá cualquiera de nuestras sucursales</p>
                    <button
                        className="btn-ver-sucursales"
                        onClick={() => navigate('/sucursales')}
                    >
                        Ver Sucursales
                    </button>
                </div>
            </div>
        </div>
    );
};

export default Planes;
