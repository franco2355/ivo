import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { PAYMENTS_API, SUBSCRIPTIONS_API } from '../config/api';
import '../styles/Checkout.css';

const API_URL = 'http://localhost:8081';

const Checkout = () => {
    const { planId } = useParams();
    const navigate = useNavigate();
    const [plan, setPlan] = useState(null);
    const [loading, setLoading] = useState(true);
    const [processing, setProcessing] = useState(false);
    const [formData, setFormData] = useState({
        payment_method: 'credit_card',
        card_number: '',
        card_name: '',
        card_expiry: '',
        card_cvv: '',
        auto_renovacion: false
    });

    const userId = localStorage.getItem("idUsuario");

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
                    alert("Este plan no está disponible");
                    navigate('/planes');
                    return;
                }

                setPlan(planData);
            } catch (error) {
                console.error("[Checkout] Error al cargar plan:", error);
                alert("Error al cargar el plan. Por favor, intenta nuevamente.");
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

    const handleSubmit = async (e) => {
        e.preventDefault();
        setProcessing(true);

        try {
            const token = localStorage.getItem('access_token');
            if (!token) {
                alert('Debes iniciar sesión para continuar');
                navigate('/login');
                return;
            }

            console.log('[Checkout] Iniciando proceso de suscripción...');

            // 1. Crear suscripción REAL en subscriptions-api
            const subscriptionData = {
                usuario_id: userId,
                plan_id: planId,
                metodo_pago: formData.payment_method,
                auto_renovacion: formData.auto_renovacion,
                notas: formData.payment_method === 'cash'
                    ? 'Pago pendiente en sucursal'
                    : `Pago con tarjeta terminada en ${formData.card_number.slice(-4)}`
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

            if (!subscriptionResponse.ok) {
                const errorData = await subscriptionResponse.json().catch(() => ({}));
                throw new Error(errorData.error || 'Error al crear la suscripción');
            }

            const suscripcion = await subscriptionResponse.json();
            console.log('[Checkout] ✅ Suscripción creada:', suscripcion);

            // 2. Crear pago en payments-api
            const paymentData = {
                entity_type: "subscription",
                entity_id: suscripcion.id,
                user_id: userId,
                amount: plan.precio_mensual,
                currency: "ARS",
                payment_method: formData.payment_method,
                payment_gateway: "mock", // Usar "mock" para testing sin credenciales
                metadata: {
                    plan_nombre: plan.nombre,
                    duracion_dias: plan.duracion_dias,
                    auto_renovacion: formData.auto_renovacion,
                    card_last4: formData.payment_method === 'credit_card' ? formData.card_number.slice(-4) : null
                }
            };

            console.log('[Checkout] Creando pago:', paymentData);

            const paymentResponse = await fetch(PAYMENTS_API.payments, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(paymentData)
            });

            if (!paymentResponse.ok) {
                throw new Error("Error al registrar el pago");
            }

            const payment = await paymentResponse.json();
            console.log('[Checkout] ✅ Pago creado:', payment);

            // 3. El pago queda pendiente de aprobación por el administrador
            const mensajePago = formData.payment_method === 'cash'
                ? "Tu solicitud de pago en efectivo ha sido registrada. Por favor, acercate a la sucursal para completar el pago. El administrador aprobará tu pago una vez que se verifique."
                : "Tu pago con tarjeta ha sido registrado y está pendiente de verificación. El administrador lo revisará y aprobará en breve.";

            alert(`¡Suscripción creada exitosamente! ${mensajePago}`);
            navigate('/mi-suscripcion');

        } catch (error) {
            console.error("[Checkout] ❌ Error:", error);
            alert(`Error al procesar la suscripción: ${error.message}`);
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

                        <div className="payment-methods">
                            <label className={`payment-method ${formData.payment_method === 'credit_card' ? 'selected' : ''}`}>
                                <input
                                    type="radio"
                                    name="payment_method"
                                    value="credit_card"
                                    checked={formData.payment_method === 'credit_card'}
                                    onChange={handleInputChange}
                                />
                                <span>Tarjeta de Crédito/Débito</span>
                            </label>
                            <label className={`payment-method ${formData.payment_method === 'cash' ? 'selected' : ''}`}>
                                <input
                                    type="radio"
                                    name="payment_method"
                                    value="cash"
                                    checked={formData.payment_method === 'cash'}
                                    onChange={handleInputChange}
                                />
                                <span>Efectivo (en sucursal)</span>
                            </label>
                        </div>

                        {formData.payment_method === 'credit_card' && (
                            <div className="card-details">
                                <div className="form-group">
                                    <label>Número de Tarjeta</label>
                                    <input
                                        type="text"
                                        name="card_number"
                                        placeholder="1234 5678 9012 3456"
                                        value={formData.card_number}
                                        onChange={handleInputChange}
                                        maxLength="16"
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Nombre en la Tarjeta</label>
                                    <input
                                        type="text"
                                        name="card_name"
                                        placeholder="JUAN PEREZ"
                                        value={formData.card_name}
                                        onChange={handleInputChange}
                                        required
                                    />
                                </div>
                                <div className="form-row">
                                    <div className="form-group">
                                        <label>Vencimiento</label>
                                        <input
                                            type="text"
                                            name="card_expiry"
                                            placeholder="MM/AA"
                                            value={formData.card_expiry}
                                            onChange={handleInputChange}
                                            maxLength="5"
                                            required
                                        />
                                    </div>
                                    <div className="form-group">
                                        <label>CVV</label>
                                        <input
                                            type="text"
                                            name="card_cvv"
                                            placeholder="123"
                                            value={formData.card_cvv}
                                            onChange={handleInputChange}
                                            maxLength="4"
                                            required
                                        />
                                    </div>
                                </div>
                            </div>
                        )}

                        {formData.payment_method === 'cash' && (
                            <div className="cash-info">
                                <p>Podés abonar en cualquiera de nuestras sucursales.</p>
                                <p>Recordá llevar tu DNI y mencionar que estás abonando el plan <strong>{plan.nombre}</strong>.</p>
                            </div>
                        )}

                        <div className="form-group checkbox-group">
                            <label>
                                <input
                                    type="checkbox"
                                    name="auto_renovacion"
                                    checked={formData.auto_renovacion}
                                    onChange={handleInputChange}
                                />
                                <span>Activar renovación automática</span>
                            </label>
                        </div>

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
                            >
                                {processing ? 'Procesando...' : `Pagar $${totalPagar.toFixed(2)}`}
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
