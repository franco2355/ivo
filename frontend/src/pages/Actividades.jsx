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
        soloInscripto: false,
        soloMiPlan: false
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
        if (isLoggedIn) {
            fetchInscripciones();
            if (!isAdmin) {
                fetchSuscripcionActiva();
            }
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
        // Nota: NO incluir inscripciones ni suscripcionActiva como dependencias
        // porque son objetos que cambian de referencia y causan loops de peticiones
    }, [debouncedBusqueda, filtros.categoria, filtros.dia, filtros.soloInscripto, filtros.soloMiPlan]);

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

            const response = await fetch(searchUrl, {
                headers: { 'Cache-Control': 'no-cache' }
            });

            if (response.ok) {
                const data = await response.json();
                console.log("✅ Search API response:", data);

                let results = data.results || [];

                // Filtrar solo inscripto (esto se hace en cliente porque depende del estado local)
                if (filtros.soloInscripto) {
                    const idsInscripto = inscripciones.filter(insc => insc.is_activa).map(insc => insc.actividad_id);
                    console.log("🔍 DEBUG Filtro inscriptas:");
                    console.log("  - IDs inscritas:", idsInscripto);
                    console.log("  - IDs de actividades (antes filtro):", results.map(a => a.id));
                    results = results.filter(actividad => {
                        const actividadId = parseInt(actividad.id);
                        const match = idsInscripto.includes(actividadId);
                        console.log(`  - Actividad ${actividad.id} (${actividadId}): ${match ? '✅ incluida' : '❌ no incluida'}`);
                        return match;
                    });
                    console.log("  - Actividades filtradas:", results.length);
                }

                // Filtrar solo actividades permitidas por el plan
                if (filtros.soloMiPlan && suscripcionActiva?.plan_info) {
                    const plan = suscripcionActiva.plan_info;
                    if (plan.tipo_acceso === "limitado" && plan.actividades_permitidas) {
                        results = results.filter(actividad =>
                            plan.actividades_permitidas.includes(actividad.categoria)
                        );
                    }
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
                        cupo: actividad.cupo || 20, // Cupo total (default 20 si no viene)
                        lugares: actividad.cupo_disponible ?? actividad.lugares, // Lugares disponibles
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

                // La API devuelve {data: Array, ...} - extraer el array
                let results = data.data || data || [];

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

                // Filtrar solo actividades permitidas por el plan
                if (filtros.soloMiPlan && suscripcionActiva?.plan_info) {
                    const plan = suscripcionActiva.plan_info;
                    if (plan.tipo_acceso === "limitado" && plan.actividades_permitidas) {
                        results = results.filter(actividad =>
                            plan.actividades_permitidas.includes(actividad.categoria)
                        );
                    }
                }

                // Mapear campos (Activities API devuelve 'lugares', no 'cupo_disponible')
                const mappedResults = results.map(actividad => ({
                    // Mantener ambos IDs por compatibilidad con el resto del frontend
                    id: actividad.id,
                    id_actividad: actividad.id,
                    titulo: actividad.titulo,
                    descripcion: actividad.descripcion,
                    categoria: actividad.categoria,
                    instructor: actividad.instructor,
                    dia: actividad.dia,
                    hora_inicio: actividad.horario_inicio || actividad.hora_inicio,
                    hora_fin: actividad.horario_final || actividad.hora_fin,
                    cupo: actividad.cupo || 20, // Cupo total
                    lugares: actividad.lugares ?? actividad.cupo_disponible, // Lugares disponibles
                    foto_url: actividad.foto_url,
                    sucursal_id: actividad.sucursal_id,
                    sucursal_nombre: actividad.sucursal_nombre
                }));

                console.log("✅ ACTIVIDADES MAPEADAS:", mappedResults);
                console.log("✅ IDs de actividades:", mappedResults.map(a => ({id: a.id, id_actividad: a.id_actividad})));
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
            // Usar el endpoint /inscripciones que usa el token JWT para identificar al usuario
            const response = await fetch(ACTIVITIES_API.inscripciones, {
                headers: {'Authorization': `Bearer ${localStorage.getItem('access_token')}`}
            });
            if (response.ok) {
                const data = await response.json();
                console.log("✅ INSCRIPCIONES RAW:", data);
                const inscripcionesActivas = Array.isArray(data)
                    ? data.filter(insc => insc.is_activa)
                    : [];
                console.log("✅ INSCRIPCIONES ACTIVAS:", inscripcionesActivas);
                console.log("✅ IDs de actividades inscritas:", inscripcionesActivas.map(i => i.actividad_id));
                setInscripciones(inscripcionesActivas);
            } else {
                console.error("❌ ERROR al cargar inscripciones. Status:", response.status);
                const errorData = await response.json().catch(() => ({}));
                console.error("❌ Error data:", errorData);

                if (isAuthError(response)) {
                    handleSessionExpired(toast, navigate);
                } else {
                    // Si hay error pero no es de autenticación, dejamos inscripciones vacío
                    setInscripciones([]);
                }
            }
        } catch (error) {
            console.error("❌ EXCEPCIÓN al cargar inscripciones:", error);
            setInscripciones([]);
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
            return true;
        }

        const plan = suscripcionActiva.plan_info;

        if (plan.tipo_acceso === "completo") {
            return true;
        }

        if (plan.tipo_acceso === "limitado" && plan.actividades_permitidas) {
            return plan.actividades_permitidas.includes(actividad.categoria);
        }

        return true;
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
                    actividad_id: parseInt(actividadId),
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
            // Delay para dar tiempo a RabbitMQ + Solr + flush caché
            await new Promise(resolve => setTimeout(resolve, 1500));
            // Actualizar la lista de actividades desde Search API
            await searchActividades();
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
                // Delay para dar tiempo a RabbitMQ + Solr + flush caché
                await new Promise(resolve => setTimeout(resolve, 1500));
                // Actualizar la lista de actividades desde Search API
                await searchActividades();
            } else if (isAuthError(response)) {
                handleSessionExpired(toast, navigate);
                return;
            } else {
                toast.error("Ups! algo salió mal, vuelve a intentarlo más tarde");
            }
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
        if (!actividad.id) {
            console.error("Error: La actividad no tiene ID", actividad);
            toast.error('Error: No se puede eliminar la actividad porque no tiene ID');
            return;
        }

        if (window.confirm('¿Estás seguro de que deseas eliminar esta actividad?')) {
            try {
                console.log("Intentando eliminar actividad con ID:", actividad.id);
                const response = await fetch(ACTIVITIES_API.actividadById(actividad.id), {
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
            Number(insc.actividad_id) === Number(id_actividad) && insc.is_activa
        );
    };

    const toggleExpand = (actividadId) => {
        console.log('Toggle expand:', { actividadId, current: expandedActividadId, type: typeof actividadId });
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
                    <>
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
                        {suscripcionActiva?.plan_info && (
                            <div className="toggle-wrapper">
                                <label className="toggle-label">
                                    <input
                                        type="checkbox"
                                        name="soloMiPlan"
                                        checked={filtros.soloMiPlan}
                                        onChange={(e) => setFiltros(prev => ({
                                            ...prev,
                                            soloMiPlan: e.target.checked
                                        }))}
                                        className="toggle-input"
                                    />
                                    <span className="toggle-slider"></span>
                                    <span className="toggle-text">Solo mi plan</span>
                                </label>
                            </div>
                        )}
                    </>
                )}
            </div>

            <div className={`actividades-grid ${expandedActividadId ? 'has-expanded' : ''}`}>
                {loading ? (
                    <Spinner size="large" message="Cargando actividades..." />
                ) : actividadesFiltradas.length === 0 ? (
                    <div className="mensaje-no-actividades">
                        No se encontraron actividades.
                    </div>
                ) : (
                    actividadesFiltradas.map((actividad) => {
                        const actividadId = actividad.id_actividad || actividad.id;
                        const inscrito = estaInscripto(actividadId);
                        const isExpanded = expandedActividadId !== null && String(expandedActividadId) === String(actividadId);
                        const shouldHide = expandedActividadId !== null && !isExpanded;
                        return (
                        <div
                            className={`actividad-card ${isExpanded ? 'expanded' : ''}`}
                            key={actividadId}
                            style={shouldHide ? { display: 'none' } : {}}
                        >
                            <h3>{actividad.titulo}</h3>
                            <div className="actividad-info-basic">
                                <p>Instructor: {actividad.instructor || "No especificado"}</p>
                                <p>
                                    Horario: {actividad.hora_inicio} a {actividad.hora_fin}
                                </p>
                            </div>

                            {isExpanded && (
                                <div className="actividad-info-expanded">
                                    <div className="actividad-imagen">
                                        <img
                                            src={actividad.foto_url || "https://media.istockphoto.com/id/1339701353/es/foto/atleta-negra-haciendo-ejercicios-de-estiramiento-mientras-calienta-con-un-grupo-de-mujeres-en.jpg?s=1024x1024&w=is&k=20&c=WCBDDweSYURamdiEqmw7IZ_uWOsgr0HkZmwVJuOLp8s="}
                                            alt={actividad.titulo}
                                            onError={(e) => {
                                                e.target.onerror = null;
                                                e.target.src = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='300' height='200'%3E%3Crect width='300' height='200' fill='%23e0e0e0'/%3E%3Ctext x='50%25' y='50%25' dominant-baseline='middle' text-anchor='middle' font-family='Arial, sans-serif' font-size='18' fill='%23757575'%3ESin imagen%3C/text%3E%3C/svg%3E";
                                            }}
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
                                                        className={`inscripcion-button ${inscrito ? 'cancelar' : ''}`}
                                                        onClick={() =>
                                                            inscrito ?
                                                                handleUnenrolling(actividadId) :
                                                                handleEnroling(actividadId)
                                                        }
                                                    >
                                                        {inscrito ? "Cancelar Inscripción" : "Inscribirse"}
                                                    </button>
                                                )}
                                            </>
                                        )}
                                    </>
                                )}
                                <button
                                    className="ver-mas-button"
                                    onClick={() => toggleExpand(actividadId)}
                                >
                                    {isExpanded ? "Ver menos 🔼" : "Ver más 🔽"}
                                </button>
                            </div>
                        </div>
                        );
                    })
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
