import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { PAYMENTS_API, SUBSCRIPTIONS_API } from '../config/api';
import '../styles/Checkout.css';
import { useToastContext } from '../context/ToastContext';
import { handleSessionExpired, isAuthError } from '../utils/auth';

const API_URL = SUBSCRIPTIONS_API.base;

const Checkout = () => {
    const { planId } = useParams();
    const navigate = useNavigate();
    const [plan, setPlan] = useState(null);
    const [loading, setLoading] = useState(true);
    const [processing, setProcessing] = useState(false);
    const [formData, setFormData] = useState({
        payment_method: 'mercadopago' // 'mercadopago' o 'cash'
    });
    const [showCheckout, setShowCheckout] = useState(false);
    const [currentPaymentId, setCurrentPaymentId] = useState(null);
    const [pollingInterval, setPollingInterval] = useState(null);

    const userId = localStorage.getItem("idUsuario");
    const toast = useToastContext();

    useEffect(() => {
        const fetchPlan = async () => {
            try {
                console.log('[Checkout] Cargando plan:', planId);

                // Cargar plan desde la API real
                const response = await fetch(`${API_URL}/plans/${planId}`);
                console.log('[Checkout] Response:', response.status);

                if (!response.ok) {
                    throw new Error('Plan no encontrado');
                }

                const planData = await response.json();
                console.log('[Checkout] Plan cargado:', planData);

                if (!planData || !planData.activo) {
                    toast.error("Este plan no está disponible");
                    navigate('/planes');
                    return;
                }

                setPlan(planData);
            } catch (error) {
                console.error("[Checkout] Error al cargar plan:", error);
                toast.error("Error al cargar el plan. Por favor, intenta nuevamente.");
                navigate('/planes');
            } finally {
                setLoading(false);
            }
        };

        fetchPlan();
    }, [planId, navigate]);

    const handleInputChange = (e) => {
        const { name, value, type, checked } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: type === 'checkbox' ? checked : value
        }));
    };

    // Cleanup polling on unmount
    useEffect(() => {
        return () => {
            if (pollingInterval) {
                clearInterval(pollingInterval);
            }
        };
    }, [pollingInterval]);

    // Auto-sync payment status con polling cada 3 segundos
    const startPaymentPolling = (paymentId) => {
        console.log('[Checkout] 🔄 Iniciando polling automático del pago...');

        let attempts = 0;
        const maxAttempts = 40; // 40 intentos × 3s = 2 minutos máximo

        const interval = setInterval(async () => {
            attempts++;
            console.log(`[Checkout] 🔄 Verificando pago (intento ${attempts}/${maxAttempts})...`);

            try {
                const syncResponse = await fetch(`http://localhost:8083/payments/${paymentId}/sync`);

                if (syncResponse.ok) {
                    const syncedPayment = await syncResponse.json();
                    console.log('[Checkout] ✅ Estado del pago:', syncedPayment.status);

                    if (syncedPayment.status === 'completed' || syncedPayment.status === 'approved') {
                        clearInterval(interval);
                        setPollingInterval(null);
                        toast.success("¡Pago completado exitosamente! Tu suscripción está activa.");
                        navigate('/mi-suscripcion');
                    } else if (syncedPayment.status === 'rejected' || syncedPayment.status === 'cancelled') {
                        clearInterval(interval);
                        setPollingInterval(null);
                        toast.error("El pago fue rechazado. Por favor, intenta nuevamente.");
                        setShowCheckout(false);
                        setProcessing(false);
                    } else if (attempts >= maxAttempts) {
                        clearInterval(interval);
                        setPollingInterval(null);
                        toast.info("Tu pago está siendo procesado. Te notificaremos cuando se complete.");
                        navigate('/mi-suscripcion');
                    }
                }
            } catch (error) {
                console.error('[Checkout] ❌ Error en polling:', error);
                if (attempts >= maxAttempts) {
                    clearInterval(interval);
                    setPollingInterval(null);
                }
            }
        }, 3000); // Cada 3 segundos

        setPollingInterval(interval);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setProcessing(true);

        try {
            const token = localStorage.getItem('access_token');
            if (!token) {
                toast.warning("Debes iniciar sesión para continuar");
                navigate('/login');
                return;
            }

            console.log('[Checkout] Iniciando proceso de suscripción...');

            // 1. Crear suscripción REAL en subscriptions-api
            const subscriptionData = {
                usuario_id: userId,
                plan_id: planId,
                metodo_pago: formData.payment_method,
                auto_renovacion: false,
                notas: formData.payment_method === 'mercadopago' ? 'Pago a través de Mercado Pago' : 'Pago en efectivo'
            };

            console.log('[Checkout] Creando suscripción:', subscriptionData);

            const subscriptionResponse = await fetch(`${API_URL}/subscriptions`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(subscriptionData)
            });

            if (isAuthError(subscriptionResponse)) {
                handleSessionExpired(toast, navigate);
                return;
            } else if (!subscriptionResponse.ok) {
                const errorData = await subscriptionResponse.json().catch(() => ({}));
                throw new Error(errorData.error || 'Error al crear la suscripción');
            }

            const suscripcion = await subscriptionResponse.json();
            console.log('[Checkout] ✅ Suscripción creada:', suscripcion);

            // 2. Procesar pago según el método seleccionado
            if (formData.payment_method === 'cash') {
                // PAGO EN EFECTIVO - Crear pago pendiente
                const paymentData = {
                    entity_type: "subscription",
                    entity_id: suscripcion.id,
                    user_id: userId,
                    amount: plan.precio_mensual,
                    currency: "ARS",
                    payment_method: "cash",
                    payment_gateway: "cash",
                    metadata: {
                        plan_nombre: plan.nombre,
                        duracion_dias: plan.duracion_dias,
                        usuario_id: userId,
                        suscripcion_id: suscripcion.id
                    }
                };

                console.log('[Checkout] Creando pago en efectivo:', paymentData);

                const paymentResponse = await fetch('http://localhost:8083/payments', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(paymentData)
                });

                if (!paymentResponse.ok) {
                    const errorData = await paymentResponse.json().catch(() => ({}));
                    throw new Error(errorData.error || "Error al crear el pago en efectivo");
                }

                const paymentResult = await paymentResponse.json();
                console.log('[Checkout] ✅ Pago en efectivo registrado:', paymentResult);

                toast.success(`Pago en efectivo registrado. Código: ${paymentResult.transaction_id}`);
                toast.info('Acércate a la sucursal dentro de las próximas 48 horas para completar el pago.');
                navigate('/mi-suscripcion');

            } else {
                // MERCADO PAGO - Procesar con gateway externo
                const paymentData = {
                    entity_type: "subscription",
                    entity_id: suscripcion.id,
                    user_id: userId,
                    amount: plan.precio_mensual,
                    currency: "ARS",
                    payment_method: "credit_card",
                    payment_gateway: "mercadopago",
                    callback_url: `${window.location.origin}/mi-suscripcion`,
                    webhook_url: `http://localhost:8083/webhooks/mercadopago`,
                    metadata: {
                        plan_nombre: plan.nombre,
                        duracion_dias: plan.duracion_dias,
                        usuario_id: userId,
                        suscripcion_id: suscripcion.id
                    }
                };

                console.log('[Checkout] Procesando pago con Mercado Pago:', paymentData);

                const paymentResponse = await fetch('http://localhost:8083/payments/process', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(paymentData)
                });

                if (!paymentResponse.ok) {
                    const errorData = await paymentResponse.json().catch(() => ({}));
                    throw new Error(errorData.error || "Error al procesar el pago con Mercado Pago");
                }

                const paymentResult = await paymentResponse.json();
                console.log('[Checkout] ✅ Respuesta de Mercado Pago:', paymentResult);

                // 3. Abrir checkout de Mercado Pago
                if (paymentResult.metadata && paymentResult.metadata.payment_url) {
                    const preferenceId = paymentResult.transaction_id;

                    console.log('[Checkout] 🎯 Preference ID:', preferenceId);
                    console.log('[Checkout] 🌐 Payment URL:', paymentResult.metadata.payment_url);

                    const paymentId = paymentResult.id;
                    setCurrentPaymentId(paymentId);

                    console.log('[Checkout] ✅ Abriendo checkout de Mercado Pago...');

                    // Iniciar polling automático
                    startPaymentPolling(paymentId);

                    // Redireccionar a Mercado Pago
                    window.location.href = paymentResult.metadata.payment_url;
                } else {
                    console.error('[Checkout] ❌ No se recibió payment_url de Mercado Pago');
                    toast.error("Error: No se pudo generar el link de pago. Por favor, intenta nuevamente.");
                }
            }

        } catch (error) {
            console.error("[Checkout] ❌ Error:", error);
            toast.error(`Error al procesar la suscripción: ${error.message}`);
        } finally {
            setProcessing(false);
        }
    };

    if (loading) {
        return (
            <div className="checkout-container">
                <div className="loading-message">Cargando información...</div>
            </div>
        );
    }

    if (!plan) {
        return null;
    }

    const totalPagar = plan.precio_mensual;

    return (
        <div className="checkout-container">
            <div className="checkout-header">
                <h1>Finalizar Suscripción</h1>
            </div>

            <div className="checkout-content">
                <div className="checkout-form-section">
                    <div className="resumen-plan">
                        <h2>Resumen de tu plan</h2>
                        <div className="plan-seleccionado" style={{ borderLeftColor: plan.color }}>
                            <h3>{plan.nombre}</h3>
                            <p>{plan.descripcion}</p>
                            <div className="plan-precio-resumen">
                                <span className="precio">${plan.precio_mensual.toFixed(2)}</span>
                                <span className="periodo">/ mes</span>
                            </div>
                            {plan.duracion_dias > 30 && (
                                <p className="plan-duracion">Duración: {plan.duracion_dias} días</p>
                            )}
                        </div>
                    </div>

                    <form className="payment-form" onSubmit={handleSubmit}>
                        <h2>Método de Pago</h2>

                        <div className="payment-methods" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px' }}>
                            <div
                                className={`payment-method ${formData.payment_method === 'mercadopago' ? 'selected' : ''}`}
                                onClick={() => setFormData({ ...formData, payment_method: 'mercadopago' })}
                                style={{
                                    border: formData.payment_method === 'mercadopago' ? '2px solid #009ee3' : '2px solid #ddd',
                                    padding: '20px',
                                    borderRadius: '8px',
                                    backgroundColor: formData.payment_method === 'mercadopago' ? '#f0f9ff' : '#fff',
                                    textAlign: 'center',
                                    cursor: 'pointer',
                                    transition: 'all 0.3s ease'
                                }}>
                                <div style={{
                                    fontSize: '48px',
                                    marginBottom: '10px'
                                }}>💳</div>
                                <h3 style={{ margin: '0 0 8px 0', color: formData.payment_method === 'mercadopago' ? '#009ee3' : '#333' }}>
                                    Mercado Pago
                                </h3>
                                <p style={{ margin: '0', fontSize: '14px', color: '#666' }}>
                                    Tarjetas, débito, transferencia
                                </p>
                            </div>

                            <div
                                className={`payment-method ${formData.payment_method === 'cash' ? 'selected' : ''}`}
                                onClick={() => setFormData({ ...formData, payment_method: 'cash' })}
                                style={{
                                    border: formData.payment_method === 'cash' ? '2px solid #4CAF50' : '2px solid #ddd',
                                    padding: '20px',
                                    borderRadius: '8px',
                                    backgroundColor: formData.payment_method === 'cash' ? '#f0fff4' : '#fff',
                                    textAlign: 'center',
                                    cursor: 'pointer',
                                    transition: 'all 0.3s ease'
                                }}>
                                <div style={{
                                    fontSize: '48px',
                                    marginBottom: '10px'
                                }}>💵</div>
                                <h3 style={{ margin: '0 0 8px 0', color: formData.payment_method === 'cash' ? '#4CAF50' : '#333' }}>
                                    Efectivo
                                </h3>
                                <p style={{ margin: '0', fontSize: '14px', color: '#666' }}>
                                    Pagar en sucursal (48hs)
                                </p>
                            </div>
                        </div>

                        {formData.payment_method === 'cash' && (
                            <div style={{
                                backgroundColor: '#fff3cd',
                                border: '1px solid #ffc107',
                                borderRadius: '8px',
                                padding: '15px',
                                marginTop: '20px'
                            }}>
                                <strong>⚠️ Importante:</strong> El pago en efectivo debe realizarse en sucursal dentro de las próximas 48 horas. Se te proporcionará un código de pago que deberás presentar en caja.
                            </div>
                        )}

                        <div className="total-section">
                            <div className="total-row">
                                <span>Subtotal:</span>
                                <span>${totalPagar.toFixed(2)}</span>
                            </div>
                            <div className="total-row total">
                                <span>Total a Pagar:</span>
                                <span>${totalPagar.toFixed(2)}</span>
                            </div>
                        </div>

                        <div className="form-actions">
                            <button
                                type="button"
                                className="btn-cancelar"
                                onClick={() => navigate('/planes')}
                                disabled={processing}
                            >
                                Cancelar
                            </button>
                            <button
                                type="submit"
                                className="btn-pagar"
                                disabled={processing}
                                style={{
                                    backgroundColor: formData.payment_method === 'cash' ? '#4CAF50' : '#009ee3',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    gap: '8px'
                                }}
                            >
                                {processing ? (
                                    formData.payment_method === 'cash' ? 'Registrando pago...' : 'Abriendo Mercado Pago...'
                                ) : (
                                    <>
                                        <span>
                                            {formData.payment_method === 'cash' ? 'Generar código de pago' : 'Pagar con Mercado Pago'}
                                        </span>
                                        <span>${totalPagar.toFixed(2)}</span>
                                    </>
                                )}
                            </button>
                        </div>
                    </form>
                </div>

                <div className="checkout-info-section">
                    <div className="info-box">
                        <h3>Pago Seguro</h3>
                        <p>Tu información está protegida con encriptación SSL</p>
                    </div>
                    <div className="info-box">
                        <h3>Garantía</h3>
                        <p>7 días de garantía. Si no estás satisfecho, te devolvemos tu dinero.</p>
                    </div>
                    <div className="info-box">
                        <h3>Soporte</h3>
                        <p>¿Necesitás ayuda? Contactanos al 0351-123-4567</p>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Checkout;
