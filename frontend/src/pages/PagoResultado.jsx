import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useToastContext } from '../context/ToastContext';
import '../styles/PagoResultado.css';

const PagoResultado = () => {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const toast = useToastContext();
    const [status, setStatus] = useState('loading');
    const [paymentInfo, setPaymentInfo] = useState(null);

    useEffect(() => {
        const processPaymentResult = async () => {
            // Obtener parámetros de la URL
            const mpStatus = searchParams.get('status');
            const collection_id = searchParams.get('collection_id');
            const collection_status = searchParams.get('collection_status');
            const payment_id = searchParams.get('payment_id');
            const external_reference = searchParams.get('external_reference');
            const preference_id = searchParams.get('preference_id');

            console.log('[PagoResultado] Parámetros recibidos de Mercado Pago:', {
                mpStatus,
                collection_id,
                collection_status,
                payment_id,
                external_reference,
                preference_id
            });

            // Determinar el estado del pago
            let finalStatus = 'pending';

            if (mpStatus === 'approved' || collection_status === 'approved') {
                finalStatus = 'success';
            } else if (mpStatus === 'rejected' || collection_status === 'rejected') {
                finalStatus = 'rejected';
            } else if (mpStatus === 'pending' || collection_status === 'pending') {
                finalStatus = 'pending';
            }

            // Si tenemos un payment_id o collection_id, buscar info del pago
            const paymentIdToCheck = payment_id || collection_id;

            if (paymentIdToCheck) {
                try {
                    console.log('[PagoResultado] Sincronizando estado del pago:', paymentIdToCheck);

                    // Intentar sincronizar con nuestro backend
                    const syncResponse = await fetch(`http://localhost:8083/payments/${external_reference}/sync`);

                    if (syncResponse.ok) {
                        const syncedPayment = await syncResponse.json();
                        console.log('[PagoResultado] Pago sincronizado:', syncedPayment);

                        setPaymentInfo({
                            id: syncedPayment.id,
                            amount: syncedPayment.amount,
                            currency: syncedPayment.currency,
                            status: syncedPayment.status,
                            transaction_id: syncedPayment.transaction_id
                        });

                        // Actualizar estado según la respuesta
                        if (syncedPayment.status === 'completed' || syncedPayment.status === 'approved') {
                            finalStatus = 'success';
                        } else if (syncedPayment.status === 'rejected' || syncedPayment.status === 'cancelled') {
                            finalStatus = 'rejected';
                        }
                    }
                } catch (error) {
                    console.error('[PagoResultado] Error sincronizando pago:', error);
                }
            }

            setStatus(finalStatus);

            // Mostrar notificación según el resultado
            if (finalStatus === 'success') {
                toast.success('¡Pago exitoso! Tu suscripción ha sido activada.');
            } else if (finalStatus === 'rejected') {
                toast.error('El pago fue rechazado. Por favor, intenta nuevamente.');
            } else if (finalStatus === 'pending') {
                toast.info('Tu pago está siendo procesado. Te notificaremos cuando se complete.');
            }
        };

        processPaymentResult();
    }, [searchParams, toast]);

    const handleContinue = () => {
        if (status === 'success' || status === 'pending') {
            navigate('/mi-suscripcion');
        } else {
            navigate('/planes');
        }
    };

    const getStatusContent = () => {
        switch (status) {
            case 'loading':
                return {
                    icon: '⏳',
                    title: 'Procesando pago...',
                    description: 'Estamos verificando el estado de tu pago con Mercado Pago.',
                    color: '#009ee3'
                };
            case 'success':
                return {
                    icon: '✅',
                    title: '¡Pago exitoso!',
                    description: 'Tu suscripción ha sido activada correctamente. Ya podés disfrutar de todos los beneficios de tu plan.',
                    color: '#4CAF50'
                };
            case 'pending':
                return {
                    icon: '⏰',
                    title: 'Pago en proceso',
                    description: 'Tu pago está siendo procesado. Te notificaremos por email cuando se complete. Esto puede demorar unos minutos.',
                    color: '#ff9800'
                };
            case 'rejected':
                return {
                    icon: '❌',
                    title: 'Pago rechazado',
                    description: 'Lo sentimos, tu pago no pudo ser procesado. Por favor, verificá tus datos e intentá nuevamente.',
                    color: '#f44336'
                };
            default:
                return {
                    icon: '❓',
                    title: 'Estado desconocido',
                    description: 'No pudimos determinar el estado de tu pago. Por favor, contacta a soporte.',
                    color: '#757575'
                };
        }
    };

    const content = getStatusContent();

    return (
        <div className="pago-resultado-container">
            <div className="pago-resultado-card" style={{ borderTopColor: content.color }}>
                <div className="pago-resultado-icon" style={{ fontSize: '80px' }}>
                    {content.icon}
                </div>

                <h1 className="pago-resultado-title" style={{ color: content.color }}>
                    {content.title}
                </h1>

                <p className="pago-resultado-description">
                    {content.description}
                </p>

                {paymentInfo && (
                    <div className="pago-info-detalle">
                        <h3>Detalles del pago</h3>
                        <div className="info-row">
                            <span className="label">ID de Pago:</span>
                            <span className="value">{paymentInfo.id}</span>
                        </div>
                        <div className="info-row">
                            <span className="label">Monto:</span>
                            <span className="value">${paymentInfo.amount} {paymentInfo.currency}</span>
                        </div>
                        <div className="info-row">
                            <span className="label">Estado:</span>
                            <span className="value status" style={{ color: content.color }}>
                                {paymentInfo.status}
                            </span>
                        </div>
                        {paymentInfo.transaction_id && (
                            <div className="info-row">
                                <span className="label">ID de Transacción:</span>
                                <span className="value">{paymentInfo.transaction_id}</span>
                            </div>
                        )}
                    </div>
                )}

                <div className="pago-resultado-actions">
                    <button
                        className="btn-continuar"
                        onClick={handleContinue}
                        style={{ backgroundColor: content.color }}
                    >
                        {status === 'success' || status === 'pending'
                            ? 'Ver mi suscripción'
                            : 'Volver a planes'
                        }
                    </button>

                    {status === 'rejected' && (
                        <button
                            className="btn-secondary"
                            onClick={() => navigate('/planes')}
                        >
                            Elegir otro plan
                        </button>
                    )}
                </div>

                <div className="pago-resultado-soporte">
                    <p>¿Necesitás ayuda?</p>
                    <p>Contactanos al <strong>0351-123-4567</strong> o por email a <strong>soporte@gimnasio.com</strong></p>
                </div>
            </div>
        </div>
    );
};

export default PagoResultado;
