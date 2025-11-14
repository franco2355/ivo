import React, { useState, useEffect } from "react";
import EditarActividadModal from '../components/EditarActividadModal';
import "../styles/Actividades.css";
import { useNavigate } from "react-router-dom";

const Actividades = () => {
    const [actividades, setActividades] = useState([]);
    const [actividadesFiltradas, setActividadesFiltradas] = useState([]);
    const [inscripciones, setInscripciones] = useState([]);
    const [actividadEditar, setActividadEditar] = useState(null);
    const [expandedActividadId, setExpandedActividadId] = useState(null);
    const [filtros, setFiltros] = useState({
        busqueda: "",
        categoria: "",
        dia: "",
        soloInscripto: false
    });
    const isLoggedIn = localStorage.getItem("isLoggedIn") === "true";
    const isAdmin = localStorage.getItem("isAdmin") === "true";
    const navigate = useNavigate();

    useEffect(() => {
        fetchInscripciones();
    }, []);

    useEffect(() => {
        searchActividades();
    }, [filtros]);

    const searchActividades = async () => {
        try {
            // Construir query params para search-api
            const params = new URLSearchParams();

            if (filtros.busqueda) {
                params.append('q', filtros.busqueda);
            }

            params.append('type', 'activity');
            params.append('page', '1');
            params.append('page_size', '100');

            // Construir URL con filtros
            let url = `http://localhost:8084/search?${params.toString()}`;

            console.log("Searching activities:", url);
            const response = await fetch(url);

            if (response.ok) {
                const data = await response.json();
                console.log("Search results:", data);

                // Aplicar filtros adicionales del lado del cliente (categoría, día, solo inscripto)
                let results = data.results || [];

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
                    const idsInscripto = inscripciones.filter(insc => insc.is_activa).map(insc => insc.id_actividad);
                    results = results.filter(actividad =>
                        idsInscripto.includes(parseInt(actividad.id))
                    );
                }

                // Mapear campos de search-api a formato esperado por el frontend
                const mappedResults = results.map(doc => ({
                    id_actividad: parseInt(doc.id),
                    titulo: doc.titulo,
                    descripcion: doc.descripcion,
                    categoria: doc.categoria,
                    instructor: doc.instructor,
                    dia: doc.dia,
                    hora_inicio: doc.horario_inicio,
                    hora_fin: doc.horario_final,
                    cupo: doc.cupo_disponible,
                    lugares: doc.cupo_disponible,
                    sucursal_nombre: doc.sucursal_nombre,
                    id_sucursal: doc.sucursal_id
                }));

                setActividadesFiltradas(mappedResults);
                setActividades(mappedResults);
            }
        } catch (error) {
            console.error("Error al buscar actividades:", error);
        }
    };
    
    const fetchInscripciones = async () => {
        try {
            const response = await fetch("http://localhost:8080/inscripciones", {
                headers: {'Authorization': `Bearer ${localStorage.getItem('access_token')}`
            },
            });
            if (response.ok) {
                const resp = await response.json();
                const data = resp.filter(insc => insc.is_activa)
                
                console.log("Inscripciones cargadas:", data);
                setInscripciones(data);
            }
        } catch (error) {
            console.error("Error al cargar inscripciones:", error);
        }
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
            const response = await fetch("http://localhost:8080/inscripciones", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${localStorage.getItem("access_token")}`,
                },
                body: JSON.stringify({
                    id: actividadId,
                }),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || "Error al inscribirse en la actividad");
            }

            // Actualizar la lista de inscripciones
            fetchInscripciones();
            // Actualizar la lista de actividades para reflejar el cambio en los cupos
            searchActividades();
            alert("¡Inscripción exitosa!");
        } catch (error) {
            console.error("Error al inscribirse:", error);
            alert(error.message);
        }
    };
    
    const handleUnenrolling = async (id_actividad) => {
        try {
            const response = await fetch("http://localhost:8080/inscripciones", {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                },
                body: JSON.stringify({
                    id: parseInt(id_actividad)
                })
            });

            if (response.status == 204) {
                alert(`Desinscripto exitosamente`);
                fetchInscripciones();
            } else {
                alert(`Ups! algo salio mal, vuelve a intentarlo mas tarde`);
            }

            searchActividades();
        } catch (error) {
            alert(`Ups! algo salio mal, vuelve a intentarlo mas tarde`);
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
            alert('Error: No se puede eliminar la actividad porque no tiene ID');
            return;
        }

        if (window.confirm('¿Estás seguro de que deseas eliminar esta actividad?')) {
            try {
                console.log("Intentando eliminar actividad con ID:", actividad.id_actividad);
                const response = await fetch(`http://localhost:8080/actividades/${actividad.id_actividad}`, {
                    method: 'DELETE',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                        'Content-Type': 'application/json'
                    }
                });

                if (response.ok) {
                    searchActividades();
                    alert('Actividad eliminada con éxito');
                } else {
                    const errorData = await response.json().catch(() => ({}));
                    alert(errorData.message || 'Error al eliminar la actividad');
                }
            } catch (error) {
                console.error("Error al eliminar:", error);
                alert('Error al eliminar la actividad');
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
                <input
                    type="text"
                    name="categoria"
                    placeholder="Categoría..."
                    value={filtros.categoria}
                    onChange={handleFiltroChange}
                    className="filtro-input"
                />
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
                {actividadesFiltradas.length === 0 ? (
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
                                            <button
                                                className="inscripcion-button"
                                                onClick={() => 
                                                    estaInscripto(actividad.id_actividad) ? 
                                                        handleUnenrolling(actividad.id_actividad) :
                                                        handleEnroling(actividad.id_actividad)
                                                }
                                            >
                                                {estaInscripto(actividad.id_actividad) ? "Desinscribir ❌" : "Inscribir ✔️"}
                                            </button>
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
