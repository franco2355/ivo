import { useState, useEffect } from 'react';
import { PAYMENTS_API, USERS_API } from '../config/api';
import '../styles/AdminPagos.css';
import { useToastContext } from '../context/ToastContext';

const AdminPagos = () => {
    const [pagos, setPagos] = useState([]);
    const [loading, setLoading] = useState(true);
    const [filtroEstado, setFiltroEstado] = useState('all');
    const [usuarios, setUsuarios] = useState({});
    const toast = useToastContext();
    const [estadisticas, setEstadisticas] = useState({
        total: 0,
        completados: 0,
        pendientes: 0,
        fallidos: 0,
        montoTotal: 0
    });
    const [paginaActual, setPaginaActual] = useState(1);
    const pagosPorPagina = 10;

    useEffect(() => {
        fetchPagos();
    }, []);

    const fetchPagos = async () => {
        try {
            setLoading(true);
            // Obtener todos los pagos (sin filtro de usuario)
            const response = await fetch(PAYMENTS_API.payments);

            if (!response.ok) {
                throw new Error("Error al cargar pagos");
            }

            const data = await response.json();
            setPagos(data);

            // Calcular estadísticas
            const stats = {
                total: data.length,
                completados: data.filter(p => p.status === 'completed').length,
                pendientes: data.filter(p => p.status === 'pending').length,
                fallidos: data.filter(p => p.status === 'failed').length,
                montoTotal: data
                    .filter(p => p.status === 'completed')
                    .reduce((sum, p) => sum + (p.amount || 0), 0)
            };
            setEstadisticas(stats);

            // Obtener nombres de usuarios únicos
            const userIds = [...new Set(data.map(p => p.user_id))];
            await fetchUsuarios(userIds);

        } catch (error) {
            console.error("Error al cargar pagos:", error);
            setPagos([]);
        } finally {
            setLoading(false);
        }
    };

    const fetchUsuarios = async (userIds) => {
        try {
            const usuariosMap = {};
            const token = localStorage.getItem('access_token');

            // Fetch cada usuario
            await Promise.all(
                userIds.map(async (userId) => {
                    try {
                        const response = await fetch(USERS_API.userById(userId), {
                            headers: {
                                'Authorization': `Bearer ${token}`
                            }
                        });
                        if (response.ok) {
                            const userData = await response.json();
                            usuariosMap[userId] = `${userData.nombre} ${userData.apellido}`;
                        } else {
                            usuariosMap[userId] = `Usuario #${userId}`;
                        }
                    } catch (error) {
                        console.error(`Error al cargar usuario ${userId}:`, error);
                        usuariosMap[userId] = `Usuario #${userId}`;
                    }
                })
            );

            setUsuarios(usuariosMap);
        } catch (error) {
            console.error("Error al cargar usuarios:", error);
        }
    };

    const handleAprobarPago = async (pago) => {
        // Validación: Solo permitir aprobar pagos en efectivo
        if (pago.payment_gateway !== 'cash') {
            toast.error('Solo se pueden aprobar manualmente los pagos en efectivo. Los pagos de MercadoPago se aprueban automáticamente.');
            return;
        }

        if (!window.confirm('¿Estás seguro de que deseas aprobar este pago en efectivo?')) {
            return;
        }

        try {
            const response = await fetch(PAYMENTS_API.approveCashPayment(pago.id), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });

            if (!response.ok) {
                throw new Error("Error al aprobar el pago");
            }

            toast.success('Pago en efectivo aprobado exitosamente');
            fetchPagos(); // Refrescar la lista
        } catch (error) {
            console.error("Error al aprobar pago:", error);
            toast.error(`Error al aprobar el pago: ${error.message}`);
        }
    };

    const handleRechazarPago = async (pago) => {
        // Validación: Solo permitir rechazar pagos en efectivo
        if (pago.payment_gateway !== 'cash') {
            toast.error('Solo se pueden rechazar manualmente los pagos en efectivo. Los pagos de MercadoPago se gestionan automáticamente.');
            return;
        }

        if (!window.confirm('¿Estás seguro de que deseas rechazar este pago en efectivo?')) {
            return;
        }

        const reason = prompt('¿Por qué rechazas este pago? (opcional)');

        try {
            const response = await fetch(PAYMENTS_API.rejectCashPayment(pago.id), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ reason: reason || 'Rechazado por administrador' })
            });

            if (!response.ok) {
                throw new Error("Error al rechazar el pago");
            }

            toast.info('Pago en efectivo rechazado');
            fetchPagos(); // Refrescar la lista
        } catch (error) {
            console.error("Error al rechazar pago:", error);
            toast.error(`Error al rechazar el pago: ${error.message}`);
        }
    };

    const getEstadoBadgeClass = (status) => {
        switch (status) {
            case 'completed':
                return 'estado-completado';
            case 'pending':
                return 'estado-pendiente';
            case 'failed':
                return 'estado-fallido';
            case 'refunded':
                return 'estado-reembolsado';
            default:
                return '';
        }
    };

    const getEstadoTexto = (status) => {
        switch (status) {
            case 'completed':
                return 'Completado';
            case 'pending':
                return 'Pendiente';
            case 'failed':
                return 'Fallido';
            case 'refunded':
                return 'Reembolsado';
            default:
                return status;
        }
    };

    // Filtrar y ordenar pagos (más recientes primero)
    const pagosFiltrados = (filtroEstado === 'all'
        ? pagos
        : pagos.filter(pago => pago.status === filtroEstado)
    ).sort((a, b) => {
        // Ordenar por fecha descendente (más reciente primero)
        return new Date(b.created_at) - new Date(a.created_at);
    });

    // Calcular paginación
    const totalPaginas = Math.ceil(pagosFiltrados.length / pagosPorPagina);
    const indiceInicio = (paginaActual - 1) * pagosPorPagina;
    const indiceFin = indiceInicio + pagosPorPagina;
    const pagosPaginados = pagosFiltrados.slice(indiceInicio, indiceFin);

    // Resetear página cuando cambia el filtro
    useEffect(() => {
        setPaginaActual(1);
    }, [filtroEstado]);

    if (loading) {
        return (
            <div className="admin-pagos-container">
                <div className="loading-message">Cargando pagos...</div>
            </div>
        );
    }

    return (
        <div className="admin-pagos-container">
            <div className="admin-pagos-header">
                <h2>Gestión de Pagos</h2>
                <button className="btn-refrescar" onClick={fetchPagos}>
                    Actualizar
                </button>
            </div>

            <div className="estadisticas-grid">
                <div className="stat-card-admin">
                    <div className="stat-icon"></div>
                    <div className="stat-info">
                        <span className="stat-value">{estadisticas.total}</span>
                        <span className="stat-label">Total Pagos</span>
                    </div>
                </div>
                <div className="stat-card-admin success">
                    <div className="stat-icon"></div>
                    <div className="stat-info">
                        <span className="stat-value">{estadisticas.completados}</span>
                        <span className="stat-label">Completados</span>
                    </div>
                </div>
                <div className="stat-card-admin warning">
                    <div className="stat-icon"></div>
                    <div className="stat-info">
                        <span className="stat-value">{estadisticas.pendientes}</span>
                        <span className="stat-label">Pendientes</span>
                    </div>
                </div>
                <div className="stat-card-admin money">
                    <div className="stat-icon"></div>
                    <div className="stat-info">
                        <span className="stat-value">${estadisticas.montoTotal.toFixed(2)}</span>
                        <span className="stat-label">Ingresos</span>
                    </div>
                </div>
            </div>

            <div className="filtros-pagos">
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
                    Completados ({estadisticas.completados})
                </button>
                <button
                    className={filtroEstado === 'pending' ? 'filtro-activo' : ''}
                    onClick={() => setFiltroEstado('pending')}
                >
                    Pendientes ({estadisticas.pendientes})
                </button>
                <button
                    className={filtroEstado === 'failed' ? 'filtro-activo' : ''}
                    onClick={() => setFiltroEstado('failed')}
                >
                    Fallidos ({estadisticas.fallidos})
                </button>
            </div>

            {pagosFiltrados.length === 0 ? (
                <div className="no-pagos-admin">
                    <p>No hay pagos para mostrar</p>
                </div>
            ) : (
                <>
                    <div className="pagos-table-container">
                        <table className="pagos-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Usuario</th>
                                    <th>Concepto</th>
                                    <th>Monto</th>
                                    <th>Gateway</th>
                                    <th>Estado</th>
                                    <th>Fecha</th>
                                    <th>Acciones</th>
                                </tr>
                            </thead>
                            <tbody>
                                {pagosPaginados.map((pago) => (
                                <tr key={pago.id}>
                                    <td className="id-cell">{pago.id.substring(0, 8)}...</td>
                                    <td>{usuarios[pago.user_id] || `Usuario #${pago.user_id}`}</td>
                                    <td>
                                        <div className="concepto-cell">
                                            <span className="entity-type">{pago.entity_type}</span>
                                            {pago.metadata?.plan_nombre && (
                                                <span className="entity-name">{pago.metadata.plan_nombre}</span>
                                            )}
                                        </div>
                                    </td>
                                    <td className="monto-cell">
                                        ${pago.amount?.toFixed(2)} {pago.currency}
                                    </td>
                                    <td>
                                        <span className={`badge-gateway ${pago.payment_gateway === 'cash' ? 'gateway-cash' : 'gateway-mp'}`}>
                                            {pago.payment_gateway === 'cash' ? '💵 Efectivo' : '💳 MercadoPago'}
                                        </span>
                                    </td>
                                    <td>
                                        <span className={`badge-estado ${getEstadoBadgeClass(pago.status)}`}>
                                            {getEstadoTexto(pago.status)}
                                        </span>
                                    </td>
                                    <td className="fecha-cell">
                                        {new Date(pago.created_at).toLocaleDateString('es-AR', {
                                            year: 'numeric',
                                            month: '2-digit',
                                            day: '2-digit',
                                            hour: '2-digit',
                                            minute: '2-digit'
                                        })}
                                    </td>
                                    <td className="acciones-cell">
                                        {pago.status === 'pending' && pago.payment_gateway === 'cash' && (
                                            <div className="acciones-buttons">
                                                <button
                                                    className="btn-aprobar"
                                                    onClick={() => handleAprobarPago(pago)}
                                                    title="Aprobar pago en efectivo"
                                                >
                                                    ✓ Aprobar
                                                </button>
                                                <button
                                                    className="btn-rechazar"
                                                    onClick={() => handleRechazarPago(pago)}
                                                    title="Rechazar pago en efectivo"
                                                >
                                                    ✗ Rechazar
                                                </button>
                                            </div>
                                        )}
                                        {pago.status === 'pending' && pago.payment_gateway !== 'cash' && (
                                            <span className="no-acciones" title="Este pago se procesa automáticamente">
                                                🔄 Automático
                                            </span>
                                        )}
                                        {pago.status !== 'pending' && (
                                            <span className="no-acciones">-</span>
                                        )}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

                {/* Controles de paginación */}
                {totalPaginas > 1 && (
                    <div className="paginacion-container">
                        <button
                            className="btn-paginacion"
                            onClick={() => setPaginaActual(prev => Math.max(prev - 1, 1))}
                            disabled={paginaActual === 1}
                        >
                            ← Anterior
                        </button>

                        <div className="paginacion-info">
                            Página {paginaActual} de {totalPaginas}
                            <span className="paginacion-total">
                                ({pagosFiltrados.length} pagos)
                            </span>
                        </div>

                        <button
                            className="btn-paginacion"
                            onClick={() => setPaginaActual(prev => Math.min(prev + 1, totalPaginas))}
                            disabled={paginaActual === totalPaginas}
                        >
                            Siguiente →
                        </button>
                    </div>
                )}
                </>
            )}
        </div>
    );
};

export default AdminPagos;
