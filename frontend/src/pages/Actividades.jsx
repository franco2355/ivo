import React, { useState, useEffect } from "react";
import EditarActividadModal from '../components/EditarActividadModal';
import Spinner from '../components/Spinner';
import "../styles/Actividades.css";
import { useNavigate } from "react-router-dom";
import { useToastContext } from '../context/ToastContext';
import { handleSessionExpired, isAuthError } from '../utils/auth';
import { ACTIVITIES_API, SEARCH_API } from '../config/api';
import { useDebounce } from '../hooks/useDebounce';

const Actividades = () => {
    const [actividades, setActividades] = useState([]);
    const [actividadesFiltradas, setActividadesFiltradas] = useState([]);
    const [inscripciones, setInscripciones] = useState([]);
    const [actividadEditar, setActividadEditar] = useState(null);
    const [expandedActividadId, setExpandedActividadId] = useState(null);
    const [loading, setLoading] = useState(true);
    const [categorias, setCategorias] = useState([]); // Lista de categorías disponibles
    const [suscripcionActiva, setSuscripcionActiva] = useState(null); // Nueva: info del plan del usuario
    const [filtros, setFiltros] = useState({
        busqueda: "",
        categoria: "",
        dia: "",
        soloInscripto: false
    });
    const isLoggedIn = localStorage.getItem("isLoggedIn") === "true";
    const isAdmin = localStorage.getItem("isAdmin") === "true";
    const userId = localStorage.getItem("idUsuario");
    const navigate = useNavigate();
    const toast = useToastContext();

    // Función helper para extraer solo la hora de timestamps ISO (ej: "2024-01-01T18:00:00Z" -> "18:00")
    const extractTime = (timeString) => {
        if (!timeString) return '';
        if (timeString.includes('T')) {
            const timePart = timeString.split('T')[1];
            return timePart ? timePart.substring(0, 5) : timeString;
        }
        return timeString;
    };

    // Debounce solo el campo de búsqueda para mejor performance
    const debouncedBusqueda = useDebounce(filtros.busqueda, 500);

    // Cargar categorías al montar el componente
    useEffect(() => {
        fetchCategorias();
        if (isLoggedIn && !isAdmin) {
            fetchInscripciones();
            fetchSuscripcionActiva();
        }
    }, []);

    const fetchCategorias = async () => {
        try {
            const response = await fetch(SEARCH_API.categories);
            if (response.ok) {
                const data = await response.json();
                setCategorias(data.categories || []);
                console.log("📂 Categorías cargadas:", data.categories);
            }
        } catch (error) {
            console.error("Error al cargar categorías:", error);
            // Fallback: categorías por defecto
            setCategorias(['yoga', 'spinning', 'funcional', 'pilates', 'crossfit', 'baile', 'boxeo', 'stretching']);
        }
    };

    // Ejecutar búsqueda cuando cambien los filtros (con debounce en búsqueda)
    useEffect(() => {
        searchActividades();
    }, [debouncedBusqueda, filtros.categoria, filtros.dia, filtros.soloInscripto, inscripciones]);

    const searchActividades = async () => {
        try {
            setLoading(true);

            // Usar Search API con Solr para búsquedas
            const searchParams = {
                q: debouncedBusqueda || undefined,
                type: 'activity', // Solo buscar actividades, no suscripciones
                categoria: filtros.categoria || undefined,
                dia: filtros.dia || undefined,
                page_size: 100 // Traer suficientes resultados
            };

            const searchUrl = SEARCH_API.buildSearchUrl(searchParams);
            console.log("🔍 Search URL:", searchUrl);

            const response = await fetch(searchUrl);

            if (response.ok) {
                const data = await response.json();
                console.log("✅ Search API response:", data);

                let results = data.results || [];

                // Filtrar solo inscripto (esto se hace en cliente porque depende del estado local)
                if (filtros.soloInscripto) {
                    const idsInscripto = inscripciones.filter(insc => insc.is_activa).map(insc => insc.id_actividad);
                    results = results.filter(actividad =>
                        idsInscripto.includes(parseInt(actividad.id))
                    );
                }

                // Mapear campos de Search API a formato esperado por el frontend
                const mappedResults = results.map(actividad => {
                    // Extraer solo la hora de los timestamps ISO (ej: "2024-01-01T18:00:00Z" -> "18:00")
                    const extractTime = (isoString) => {
                        if (!isoString) return '';
                        if (isoString.includes('T')) {
                            const timePart = isoString.split('T')[1];
                            return timePart ? timePart.substring(0, 5) : isoString;
                        }
                        return isoString;
                    };

                    return {
                        id_actividad: actividad.id,
                        titulo: actividad.titulo,
                        descripcion: actividad.descripcion,
                        categoria: actividad.categoria,
                        instructor: actividad.instructor,
                        dia: actividad.dia,
                        hora_inicio: extractTime(actividad.horario_inicio),
                        hora_fin: extractTime(actividad.horario_final),
                        cupo: actividad.cupo_disponible || actividad.cupo,
                        lugares: actividad.cupo_disponible || actividad.lugares,
                        foto_url: actividad.foto_url,
                        sucursal_id: actividad.sucursal_id
                    };
                });

                setActividadesFiltradas(mappedResults);
                setActividades(mappedResults);

                // Log de cache hit/miss
                const cacheStatus = response.headers.get('X-Cache');
                if (cacheStatus) {
                    console.log(`💾 Cache: ${cacheStatus}`);
                }
            } else {
                // Fallback a Activities API si Search falla
                console.warn("⚠️ Search API falló, usando fallback a Activities API");
                await searchActividadesFallback();
            }
        } catch (error) {
            console.error("❌ Error en Search API:", error);
            // Fallback a Activities API
            await searchActividadesFallback();
        } finally {
            setLoading(false);
        }
    };

    // Fallback: usar Activities API directamente si Search API no está disponible
    const searchActividadesFallback = async () => {
        try {
            const response = await fetch(ACTIVITIES_API.actividades);

            if (response.ok) {
                const data = await response.json();
                console.log("📋 Fallback - Activities loaded:", data);

                let results = data || [];

                // Filtrar por búsqueda (título, descripción, instructor)
                if (debouncedBusqueda) {
                    const searchLower = debouncedBusqueda.toLowerCase();
                    results = results.filter(actividad =>
                        (actividad.titulo && actividad.titulo.toLowerCase().includes(searchLower)) ||
                        (actividad.descripcion && actividad.descripcion.toLowerCase().includes(searchLower)) ||
                        (actividad.instructor && actividad.instructor.toLowerCase().includes(searchLower))
                    );
                }

                // Filtrar por categoría
                if (filtros.categoria) {
                    const categoriaLower = filtros.categoria.toLowerCase();
                    results = results.filter(actividad =>
                        actividad.categoria && actividad.categoria.toLowerCase().includes(categoriaLower)
                    );
                }

                // Filtrar por día
                if (filtros.dia) {
                    results = results.filter(actividad =>
                        actividad.dia && actividad.dia.toLowerCase() === filtros.dia.toLowerCase()
                    );
                }

                // Filtrar solo inscripto
                if (filtros.soloInscripto) {
                    const idsInscripto = inscripciones.filter(insc => insc.is_activa).map(insc => insc.actividad_id);
                    console.log("IDs inscritos:", idsInscripto);
                    console.log("Actividades antes del filtro:", results.map(a => a.id));
                    results = results.filter(actividad =>
                        idsInscripto.includes(actividad.id)
                    );
                    console.log("Actividades después del filtro:", results.map(a => a.id));
                }

                // Mapear campos
                const mappedResults = results.map(actividad => ({
                    id_actividad: actividad.id,
                    titulo: actividad.titulo,
                    descripcion: actividad.descripcion,
                    categoria: actividad.categoria,
                    instructor: actividad.instructor,
                    dia: actividad.dia,
                    hora_inicio: actividad.horario_inicio,
                    hora_fin: actividad.horario_final,
                    cupo: actividad.cupo,
                    lugares: actividad.lugares,
                    foto_url: actividad.foto_url,
                    sucursal_id: actividad.sucursal_id
                }));

                setActividadesFiltradas(mappedResults);
                setActividades(mappedResults);
            }
        } catch (error) {
            console.error("❌ Error en fallback:", error);
        }
    };
    
    const fetchInscripciones = async () => {
        if (!userId) {
            console.log("No hay usuario logueado, saltando inscripciones");
            return;
        }

        try {
            const response = await fetch(ACTIVITIES_API.inscripcionesByUsuario(userId), {
                headers: {'Authorization': `Bearer ${localStorage.getItem('access_token')}`}
            });
            if (response.ok) {
                const data = await response.json();
                const inscripcionesActivas = Array.isArray(data)
                    ? data.filter(insc => insc.is_activa)
                    : [];

                console.log("Inscripciones cargadas:", inscripcionesActivas);
                setInscripciones(inscripcionesActivas);
            } else if (isAuthError(response)) {
                handleSessionExpired(toast, navigate);
            }
        } catch (error) {
            console.error("Error al cargar inscripciones:", error);
        }
    };

    const fetchSuscripcionActiva = async () => {
        try {
            const response = await fetch(`http://localhost:8081/subscriptions/active/${userId}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('access_token')}`
                }
            });

            if (response.ok) {
                const suscripcion = await response.json();
                console.log("Suscripción activa:", suscripcion);

                // Obtener info del plan
                if (suscripcion.plan_id) {
                    const planResponse = await fetch(`http://localhost:8081/plans/${suscripcion.plan_id}`);
                    if (planResponse.ok) {
                        const plan = await planResponse.json();
                        console.log("Plan del usuario:", plan);
                        setSuscripcionActiva({
                            ...suscripcion,
                            plan_info: plan
                        });
                    }
                }
            } else if (response.status === 404) {
                console.log("Usuario no tiene suscripción activa");
                setSuscripcionActiva(null);
            } else if (isAuthError(response)) {
                handleSessionExpired(toast, navigate);
            }
        } catch (error) {
            console.error("Error al cargar suscripción activa:", error);
        }
    };

    // Helper: Verificar si una actividad está permitida por el plan del usuario
    const actividadPermitida = (actividad) => {
        if (!suscripcionActiva || !suscripcionActiva.plan_info) {
            return true; // Si no hay plan cargado, no mostramos restricciones aún
        }

        const plan = suscripcionActiva.plan_info;

        // Si el plan tiene acceso completo, todas las actividades están permitidas
        if (plan.tipo_acceso === "completo") {
            return true;
        }

        // Si el plan es limitado, verificar si la categoría está en las actividades permitidas
        if (plan.tipo_acceso === "limitado" && plan.actividades_permitidas) {
            return plan.actividades_permitidas.includes(actividad.categoria);
        }

        return true; // Por defecto, permitir
    };

    const handleFiltroChange = (e) => {
        const { name, value } = e.target;
        setFiltros(prev => ({
            ...prev,
            [name]: value
        }));
    };

    const handleEnroling = async (actividadId) => {
        if (!isLoggedIn) {
            navigate("/login");
            return;
        }

        try {
            const response = await fetch(ACTIVITIES_API.inscripciones, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${localStorage.getItem("access_token")}`,
                },
                body: JSON.stringify({
                    actividad_id: actividadId,
                }),
            });

            const data = await response.json();

            if (!response.ok) {
                // Manejar diferentes tipos de errores
                let errorMessage = "Error al inscribirse en la actividad";

                if (response.status === 400) {
                    if (data.error && data.error.toLowerCase().includes("subscription")) {
                        errorMessage = "No tienes un plan activo. Por favor, suscríbete a un plan para inscribirte en actividades.";
                    } else if (data.error && data.error.toLowerCase().includes("plan")) {
                        errorMessage = "No tienes un plan activo. Por favor, suscríbete a un plan para inscribirte en actividades.";
                    } else if (data.error && data.error.toLowerCase().includes("cupo")) {
                        errorMessage = "No hay cupos disponibles para esta actividad";
                    } else if (data.error && data.error.toLowerCase().includes("ya inscrito")) {
                        errorMessage = "Ya estás inscrito en esta actividad";
                    } else if (data.details) {
                        errorMessage = data.details;
                    } else if (data.error) {
                        errorMessage = data.error;
                    }
                } else if (response.status === 401) {
                    handleSessionExpired(toast, navigate);
                    return;
                } else if (response.status === 403) {
                    // Forbidden - problemas de suscripción o permisos
                    const errorLower = data.error ? data.error.toLowerCase() : '';
                    if (errorLower.includes("suscripción") || errorLower.includes("suscripcion")) {
                        errorMessage = "No tienes un plan activo. Por favor, suscríbete a un plan para inscribirte en actividades.";
                    } else if (errorLower.includes("premium")) {
                        errorMessage = "Esta actividad requiere un plan premium. Mejora tu suscripción para acceder.";
                    } else if (data.error) {
                        errorMessage = data.error;
                    } else {
                        errorMessage = "No tienes permisos para realizar esta acción";
                    }
                } else if (response.status === 404) {
                    // Not Found - actividad no existe
                    errorMessage = "La actividad no existe o ha sido eliminada";
                } else if (response.status === 409) {
                    // Conflict - ya inscrito
                    errorMessage = "Ya estás inscrito en esta actividad";
                } else if (response.status === 500) {
                    // Error del servidor
                    console.error("Error 500 del servidor:", data);
                    if (data.error && data.error.toLowerCase().includes("subscription")) {
                        errorMessage = "No tienes un plan activo. Por favor, suscríbete a un plan para inscribirte en actividades.";
                    } else if (data.error && data.error.toLowerCase().includes("plan")) {
                        errorMessage = "No tienes un plan activo. Por favor, suscríbete a un plan para inscribirte en actividades.";
                    } else if (data.details) {
                        errorMessage = `Error del servidor: ${data.details}`;
                    } else if (data.error) {
                        errorMessage = data.error;
                    } else {
                        errorMessage = "Error del servidor. Por favor, intenta más tarde o contacta al administrador.";
                    }
                } else if (data.error) {
                    errorMessage = data.error;
                } else if (data.details) {
                    errorMessage = data.details;
                }

                throw new Error(errorMessage);
            }

            // Actualizar la lista de inscripciones
            fetchInscripciones();
            // Actualizar la lista de actividades para reflejar el cambio en los cupos
            searchActividades();
            toast.success("¡Inscripción exitosa!");
        } catch (error) {
            console.error("Error al inscribirse:", error);
            toast.error(error.message);
        }
    };
    
    const handleUnenrolling = async (id_actividad) => {
        try {
            const response = await fetch(ACTIVITIES_API.inscripciones, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                },
                body: JSON.stringify({
                    actividad_id: parseInt(id_actividad)
                })
            });

            if (response.status == 204) {
                toast.success("Inscripción cancelada exitosamente");
                fetchInscripciones();
            } else if (isAuthError(response)) {
                handleSessionExpired(toast, navigate);
                return;
            } else {
                toast.error("Ups! algo salió mal, vuelve a intentarlo más tarde");
            }

            searchActividades();
        } catch (error) {
            toast.error("Ups! algo salió mal, vuelve a intentarlo más tarde");
            console.error("Error al desinscribir el usuario:", error);
        }
    };

    const handleEditar = (actividad) => {
        setExpandedActividadId(null); // Cerramos el detalle expandido
        setActividadEditar(actividad);
    };

    const handleCloseModal = () => {
        setActividadEditar(null);
    };

    const handleSaveEdit = () => {
        searchActividades();
    };

    const handleEliminar = async (actividad) => {
        if (!actividad.id_actividad) {
            console.error("Error: La actividad no tiene ID", actividad);
            toast.error('Error: No se puede eliminar la actividad porque no tiene ID');
            return;
        }

        if (window.confirm('¿Estás seguro de que deseas eliminar esta actividad?')) {
            try {
                console.log("Intentando eliminar actividad con ID:", actividad.id_actividad);
                const response = await fetch(ACTIVITIES_API.actividadById(actividad.id_actividad), {
                    method: 'DELETE',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                        'Content-Type': 'application/json'
                    }
                });

                if (response.ok) {
                    searchActividades();
                    toast.success('Actividad eliminada con éxito');
                } else if (isAuthError(response)) {
                    handleSessionExpired(toast, navigate);
                } else {
                    const errorData = await response.json().catch(() => ({}));
                    toast.error(errorData.message || 'Error al eliminar la actividad');
                }
            } catch (error) {
                console.error("Error al eliminar:", error);
                toast.error('Error al eliminar la actividad');
            }
        }
    };

    const estaInscripto = (id_actividad) => {
        return inscripciones.some(insc => 
            insc.id_actividad === id_actividad &&
            insc.is_activa
        )
    };

    const toggleExpand = (actividadId) => {
        setExpandedActividadId(expandedActividadId === actividadId ? null : actividadId);
    };

    return (
        <div className="actividades-container">
            {expandedActividadId && (
                <div className="actividades-modal-bg" onClick={() => setExpandedActividadId(null)} />
            )}
            <div className="filtros-container">
                <div className="search-wrapper">
                    <span className="search-icon">🔍</span>
                    <input
                        type="text"
                        name="busqueda"
                        placeholder="Buscar actividad..."
                        value={filtros.busqueda}
                        onChange={handleFiltroChange}
                        className="filtro-input"
                    />
                </div>
                <select
                    name="categoria"
                    value={filtros.categoria}
                    onChange={handleFiltroChange}
                    className="filtro-select"
                >
                    <option value="">Categoría</option>
                    {categorias.map(cat => (
                        <option key={cat} value={cat}>
                            {cat.charAt(0).toUpperCase() + cat.slice(1)}
                        </option>
                    ))}
                </select>
                <select
                    name="dia"
                    value={filtros.dia}
                    onChange={handleFiltroChange}
                    className="filtro-select"
                >
                    <option value="">Día</option>
                    <option value="Lunes">Lunes</option>
                    <option value="Martes">Martes</option>
                    <option value="Miercoles">Miercoles</option>
                    <option value="Jueves">Jueves</option>
                    <option value="Viernes">Viernes</option>
                    <option value="Sabado">Sabado</option>
                    <option value="Domingo">Domingo</option>
                </select>
                {isLoggedIn && !isAdmin && (
                    <div className="toggle-wrapper">
                        <label className="toggle-label">
                            <input
                                type="checkbox"
                                name="soloInscripto"
                                checked={filtros.soloInscripto}
                                onChange={(e) => setFiltros(prev => ({
                                    ...prev,
                                    soloInscripto: e.target.checked
                                }))}
                                className="toggle-input"
                            />
                            <span className="toggle-slider"></span>
                            <span className="toggle-text">Solo inscriptas</span>
                        </label>
                    </div>
                )}
            </div>

            <div className="actividades-grid">
                {loading ? (
                    <Spinner size="large" message="Cargando actividades..." />
                ) : actividadesFiltradas.length === 0 ? (
                    <div className="mensaje-no-actividades">
                        No se encontraron actividades.
                    </div>
                ) : (
                    actividadesFiltradas.map((actividad) => (
                        <div 
                            className={`actividad-card ${expandedActividadId === actividad.id_actividad ? 'expanded' : ''}`} 
                            key={actividad.id_actividad}
                        >
                            <h3>{actividad.titulo}</h3>
                            <div className="actividad-info-basic">
                                <p>Instructor: {actividad.instructor || "No especificado"}</p>
                                <p>
                                    Horario: {actividad.hora_inicio} a {actividad.hora_fin}
                                </p>
                            </div>

                            {expandedActividadId === actividad.id_actividad && (
                                <div className="actividad-info-expanded">
                                    <div className="actividad-imagen">
                                        <img 
                                            src={actividad.foto_url || "https://via.placeholder.com/300x200"} 
                                            alt={actividad.titulo}
                                        />
                                    </div>
                                    <div className="actividad-detalles">
                                        <p>{actividad.descripcion}</p>
                                        <p>Categoría: {actividad.categoria || "No especificada"}</p>
                                        <p>Día: {actividad.dia || "No especificado"}</p>
                                        <p><b>Horario:</b> {actividad.hora_inicio} a {actividad.hora_fin}</p>
                                        <p>Cupo total: {actividad.cupo} | Lugares disponibles: {actividad.lugares}</p>
                                    </div>
                                </div>
                            )}

                            <div className="card-actions">
                                {isLoggedIn && (
                                    <>
                                        {isAdmin ? (
                                            <>
                                                <button
                                                    className="edit-button"
                                                    onClick={() => handleEditar(actividad)}
                                                    title="Editar"
                                                >
                                                    <span>✏️</span>
                                                    Editar
                                                </button>
                                                <button
                                                    className="delete-button"
                                                    onClick={() => handleEliminar(actividad)}
                                                    title="Eliminar"
                                                >
                                                    <span>🗑️</span>
                                                    Eliminar
                                                </button>
                                            </>
                                        ) : (
                                            <>
                                                {!actividadPermitida(actividad) ? (
                                                    <div className="restriccion-plan-container">
                                                        <button
                                                            className="inscripcion-button disabled"
                                                            disabled
                                                            title={`Tu plan no incluye ${actividad.categoria}`}
                                                        >
                                                            🔒 No permitido
                                                        </button>
                                                        <span className="restriccion-mensaje">
                                                            ⚠️ Tu plan "{suscripcionActiva?.plan_info?.nombre}" no incluye {actividad.categoria}
                                                        </span>
                                                        <button
                                                            className="upgrade-plan-button"
                                                            onClick={() => navigate('/planes')}
                                                        >
                                                            Actualizar plan →
                                                        </button>
                                                    </div>
                                                ) : (
                                                    <button
                                                        className={`inscripcion-button ${estaInscripto(actividad.id_actividad) ? 'cancelar' : ''}`}
                                                        onClick={() =>
                                                            estaInscripto(actividad.id_actividad) ?
                                                                handleUnenrolling(actividad.id_actividad) :
                                                                handleEnroling(actividad.id_actividad)
                                                        }
                                                    >
                                                        {estaInscripto(actividad.id_actividad) ? "Cancelar Inscripción" : "Inscribirse"}
                                                    </button>
                                                )}
                                            </>
                                        )}
                                    </>
                                )}
                                <button
                                    className="ver-mas-button"
                                    onClick={() => toggleExpand(actividad.id_actividad)}
                                >
                                    {expandedActividadId === actividad.id_actividad ? "Ver menos 🔼" : "Ver más 🔽"}
                                </button>
                            </div>
                        </div>
                    ))
                )}
            </div>

            {actividadEditar && (
                <EditarActividadModal
                    actividad={actividadEditar}
                    onClose={handleCloseModal}
                    onSave={handleSaveEdit}
                />
            )}
        </div>
    );
};

export default Actividades;
