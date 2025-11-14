import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { PAYMENTS_API } from '../config/api';
import {
    getPaymentStatusBadgeClass,
    getPaymentStatusLabel,
    normalizePaymentStatus,
} from '../utils/paymentStatus';
import '../styles/Pagos.css';

const Pagos = () => {
    const [pagos, setPagos] = useState([]);
    const [loading, setLoading] = useState(true);
    const [filtroEstado, setFiltroEstado] = useState('all');
    const navigate = useNavigate();
    const userId = localStorage.getItem("idUsuario");

    useEffect(() => {
        fetchPagos();
    }, [userId]);

    const fetchPagos = async () => {
        try {
            setLoading(true);
            const response = await fetch(PAYMENTS_API.paymentsByUser(userId));

            if (!response.ok) {
                throw new Error("Error al cargar pagos");
            }

            const data = await response.json();
            console.log("Pagos cargados:", data);
            const pagosNormalizados = data.map((pago) => ({
                ...pago,
                status: normalizePaymentStatus(
                    pago.status ?? pago.estado ?? pago.payment_status
                )
            }));
            setPagos(pagosNormalizados);
        } catch (error) {
            console.error("Error al cargar pagos:", error);
            // Si falla la API, mostrar array vacío
            setPagos([]);
        } finally {
            setLoading(false);
        }
    };

    const getEstadoBadgeClass = (status) => getPaymentStatusBadgeClass(status);

    const getEstadoTexto = (status) => getPaymentStatusLabel(status);

    const getMetodoPagoIcono = (metodo) => {
        switch (metodo) {
            case 'credit_card':
                return '💳';
            case 'debit_card':
                return '💳';
            case 'cash':
                return '💵';
            case 'transfer':
                return '🏦';
            default:
                return '💰';
        }
    };

    const pagosFiltrados = filtroEstado === 'all'
        ? pagos
        : pagos.filter(pago => pago.status === filtroEstado);

    if (loading) {
        return (
            <div className="pagos-container">
                <div className="loading-message">Cargando historial de pagos...</div>
            </div>
        );
    }

    return (
        <div className="pagos-container">
            <div className="pagos-header">
                <h1>Historial de Pagos</h1>
                <p>Consultá todos tus pagos realizados</p>
            </div>

            <div className="pagos-filtros">
                <button
                    className={filtroEstado === 'all' ? 'filtro-activo' : ''}
                    onClick={() => setFiltroEstado('all')}
                >
                    Todos ({pagos.length})
                </button>
                <button
                    className={filtroEstado === 'completed' ? 'filtro-activo' : ''}
                    onClick={() => setFiltroEstado('completed')}
                >
                    Completados ({pagos.filter(p => p.status === 'completed').length})
                </button>
                <button
                    className={filtroEstado === 'pending' ? 'filtro-activo' : ''}
                    onClick={() => setFiltroEstado('pending')}
                >
                    Pendientes ({pagos.filter(p => p.status === 'pending').length})
                </button>
                <button
                    className={filtroEstado === 'failed' ? 'filtro-activo' : ''}
                    onClick={() => setFiltroEstado('failed')}
                >
                    Fallidos ({pagos.filter(p => p.status === 'failed').length})
                </button>
            </div>

            {pagosFiltrados.length === 0 ? (
                <div className="no-pagos">
                    <div className="no-pagos-icon">💳</div>
                    <h2>No hay pagos registrados</h2>
                    <p>Cuando realices un pago, aparecerá aquí</p>
                    <button
                        className="btn-ver-planes"
                        onClick={() => navigate('/planes')}
                    >
                        Ver Planes Disponibles
                    </button>
                </div>
            ) : (
                <div className="pagos-lista">
                    {pagosFiltrados.map((pago) => (
                        <div key={pago.id} className="pago-card">
                            <div className="pago-header">
                                <div className="pago-tipo">
                                    <span className="tipo-icono">
                                        {pago.entity_type === 'subscription' ? '📋' : '🎯'}
                                    </span>
                                    <div className="tipo-info">
                                        <span className="tipo-label">
                                            {pago.entity_type === 'subscription' ? 'Suscripción' : 'Actividad'}
                                        </span>
                                        {pago.metadata?.plan_nombre && (
                                            <span className="tipo-nombre">{pago.metadata.plan_nombre}</span>
                                        )}
                                    </div>
                                </div>
                                <div className={`pago-estado ${getEstadoBadgeClass(pago.status)}`}>
                                    {getEstadoTexto(pago.status)}
                                </div>
                            </div>

                            <div className="pago-detalles">
                                <div className="detalle-row">
                                    <span className="detalle-label">ID de Pago:</span>
                                    <span className="detalle-valor">{pago.id}</span>
                                </div>
                                <div className="detalle-row">
                                    <span className="detalle-label">Monto:</span>
                                    <span className="detalle-valor monto">
                                        ${pago.amount?.toFixed(2)} {pago.currency}
                                    </span>
                                </div>
                                <div className="detalle-row">
                                    <span className="detalle-label">Método:</span>
                                    <span className="detalle-valor">
                                        {getMetodoPagoIcono(pago.payment_method)} {pago.payment_method}
                                    </span>
                                </div>
                                <div className="detalle-row">
                                    <span className="detalle-label">Fecha:</span>
                                    <span className="detalle-valor">
                                        {new Date(pago.created_at).toLocaleString('es-AR')}
                                    </span>
                                </div>
                                {pago.processed_at && (
                                    <div className="detalle-row">
                                        <span className="detalle-label">Procesado:</span>
                                        <span className="detalle-valor">
                                            {new Date(pago.processed_at).toLocaleString('es-AR')}
                                        </span>
                                    </div>
                                )}
                                {pago.transaction_id && (
                                    <div className="detalle-row">
                                        <span className="detalle-label">ID Transacción:</span>
                                        <span className="detalle-valor">{pago.transaction_id}</span>
                                    </div>
                                )}
                            </div>

                            {pago.metadata && Object.keys(pago.metadata).length > 0 && (
                                <div className="pago-metadata">
                                    <span className="metadata-titulo">Detalles adicionales:</span>
                                    {pago.metadata.duracion_dias && (
                                        <span className="metadata-item">
                                            Duración: {pago.metadata.duracion_dias} días
                                        </span>
                                    )}
                                    {pago.metadata.card_last4 && (
                                        <span className="metadata-item">
                                            Tarjeta: **** {pago.metadata.card_last4}
                                        </span>
                                    )}
                                </div>
                            )}

                            <div className="pago-acciones">
                                {pago.status === 'completed' && (
                                    <button className="btn-descargar-recibo">
                                        📄 Descargar Recibo
                                    </button>
                                )}
                                {pago.status === 'pending' && (
                                    <button
                                        className="btn-completar-pago"
                                        onClick={() => navigate(`/checkout/${pago.entity_id}`)}
                                    >
                                        Completar Pago
                                    </button>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default Pagos;
