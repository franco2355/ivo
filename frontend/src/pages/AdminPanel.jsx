import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import EditarActividadModal from '../components/EditarActividadModal';
import AgregarActividadModal from '../components/AgregarActividadModal';
import AdminPlanes from '../components/AdminPlanes';
import AdminPagos from '../components/AdminPagos';
import '../styles/AdminPanel.css';
import { useToastContext } from '../context/ToastContext';
import { handleSessionExpired, isAuthError } from '../utils/auth';
import { ACTIVITIES_API } from '../config/api';

const AdminPanel = () => {
    const [tabActiva, setTabActiva] = useState('actividades');
    const [actividades, setActividades] = useState([]);
    const [actividadEditar, setActividadEditar] = useState(null);
    const [mostrarAgregarModal, setMostrarAgregarModal] = useState(false);
    const [validandoSesion, setValidandoSesion] = useState(true);
    const navigate = useNavigate();
    const toast = useToastContext();

    // Función para cargar actividades (con useCallback para evitar re-renders)
    const fetchActividades = useCallback(async () => {
        try {
            const response = await fetch(ACTIVITIES_API.actividades);
            if (response.ok) {
                const responseData = await response.json();
                // El API devuelve { data: [...], total, page, etc }
                const data = responseData.data || responseData || [];
                // Filtrar actividades que tengan id válido (el API devuelve "id" no "id_actividad")
                const actividadesValidas = Array.isArray(data) ? data.filter(act => act.id) : [];
                setActividades(actividadesValidas);
            }
        } catch (error) {
            console.error("Error al cargar actividades:", error);
            toast.error('Error al cargar las actividades');
        }
    }, [toast]);

    // Validar sesión con el backend al cargar
    useEffect(() => {
        const validarSesionAdmin = async () => {
            const token = localStorage.getItem('access_token');

            // Si no hay token, redirigir al login
            if (!token) {
                toast.error('Debes iniciar sesión para acceder al panel de administración');
                navigate('/login');
                return;
            }

            try {
                // Validar token con el backend haciendo una petición autenticada
                const response = await fetch(ACTIVITIES_API.actividades, {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });

                if (isAuthError(response)) {
                    handleSessionExpired(toast, navigate);
                    return;
                }

                // Verificar que el usuario sea admin desde el token decodificado
                const isAdmin = localStorage.getItem("isAdmin") === "true";
                if (!isAdmin) {
                    toast.error('No tienes permisos de administrador');
                    navigate('/');
                    return;
                }

                // Sesión válida, cargar datos
                setValidandoSesion(false);
                fetchActividades();
            } catch (error) {
                console.error('Error validando sesión:', error);
                toast.error('Error al validar la sesión');
                navigate('/login');
            }
        };

        validarSesionAdmin();
    }, [navigate, toast, fetchActividades]);

    const handleEditar = (actividad) => {
        setActividadEditar(actividad);
    };

    const handleCloseModal = () => {
        setActividadEditar(null);
        setMostrarAgregarModal(false);
    };

    const handleSaveEdit = () => {
        fetchActividades();
        handleCloseModal();
    };

    const handleEliminar = async (actividad) => {
        if (!actividad.id) {
            console.error("Error: La actividad no tiene ID", actividad);
            toast.error('Error: No se puede eliminar la actividad porque no tiene ID');
            return;
        }

        if (window.confirm('¿Estás seguro de que deseas eliminar esta actividad? Se eliminarán también todas las inscripciones asociadas.')) {
            try {
                const response = await fetch(ACTIVITIES_API.actividadById(actividad.id), {
                    method: 'DELETE',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                        'Content-Type': 'application/json'
                    }
                });

                if (isAuthError(response)) {
                    handleSessionExpired(toast, navigate);
                } else if (response.ok) {
                    fetchActividades();
                    toast.success('Actividad eliminada con éxito');
                } else {
                    const errorData = await response.json();
                    toast.error(errorData.error || 'Error al eliminar la actividad');
                }
            } catch (error) {
                console.error("Error al eliminar:", error);
                toast.error('Error al eliminar la actividad. Por favor, intenta de nuevo más tarde.');
            }
        }
    };

    // Mostrar loading mientras se valida la sesión
    if (validandoSesion) {
        return (
            <div className="admin-container">
                <div style={{ textAlign: 'center', padding: '50px' }}>
                    <p>Validando permisos de administrador...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="admin-container">
            <div className="admin-header">
                <h2>Panel de Administración</h2>
            </div>

            <div className="admin-tabs">
                <button
                    className={`tab ${tabActiva === 'actividades' ? 'tab-active' : ''}`}
                    onClick={() => setTabActiva('actividades')}
                >
                    🏋️ Actividades
                </button>
                <button
                    className={`tab ${tabActiva === 'planes' ? 'tab-active' : ''}`}
                    onClick={() => setTabActiva('planes')}
                >
                    📋 Planes
                </button>
                <button
                    className={`tab ${tabActiva === 'pagos' ? 'tab-active' : ''}`}
                    onClick={() => setTabActiva('pagos')}
                >
                    💳 Pagos
                </button>
            </div>

            <div className="admin-content">
                {tabActiva === 'actividades' && (
                    <div className="actividades-section">
                        <div className="section-header">
                            <h3>Gestión de Actividades</h3>
                            <button
                                className="btn-agregar"
                                onClick={() => setMostrarAgregarModal(true)}
                            >
                                <span>+</span>
                                Agregar Actividad
                            </button>
                        </div>

                        <div className="admin-table-container">
                            <table className="admin-table">
                                <thead>
                                    <tr>
                                        <th>Título</th>
                                        <th>Descripción</th>
                                        <th>Instructor</th>
                                        <th>Categoría</th>
                                        <th>Día</th>
                                        <th>Horario</th>
                                        <th>Cupo</th>
                                        <th>Acciones</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {actividades.length === 0 ? (
                                        <tr>
                                            <td colSpan="8" style={{ textAlign: 'center', padding: '20px' }}>
                                                No hay actividades disponibles
                                            </td>
                                        </tr>
                                    ) : (
                                        actividades.map((actividad, index) => (
                                            <tr key={actividad.id || `actividad-${index}`}>
                                                <td>{actividad.titulo}</td>
                                                <td>{actividad.descripcion}</td>
                                                <td>{actividad.instructor}</td>
                                                <td>{actividad.categoria}</td>
                                                <td>{actividad.dia}</td>
                                                <td>{actividad.horario_inicio} - {actividad.horario_final}</td>
                                                <td>{actividad.cupo - actividad.lugares} / {actividad.cupo}</td>
                                                <td className="acciones-column">
                                                    <button
                                                        className="action-button edit-button"
                                                        onClick={() => handleEditar(actividad)}
                                                        title="Editar"
                                                    >
                                                        ✏️
                                                    </button>
                                                    <button
                                                        className="action-button delete-button"
                                                        onClick={() => handleEliminar(actividad)}
                                                        title="Eliminar"
                                                    >
                                                        🗑️
                                                    </button>
                                                </td>
                                            </tr>
                                        ))
                                    )}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}

                {tabActiva === 'planes' && <AdminPlanes />}

                {tabActiva === 'pagos' && <AdminPagos />}
            </div>

            {actividadEditar && (
                <EditarActividadModal
                    actividad={actividadEditar}
                    onClose={handleCloseModal}
                    onSave={handleSaveEdit}
                />
            )}

            {mostrarAgregarModal && (
                <AgregarActividadModal
                    onClose={handleCloseModal}
                    onSave={handleSaveEdit}
                />
            )}
        </div>
    );
};

export default AdminPanel; 