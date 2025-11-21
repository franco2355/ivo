import { useState, useEffect } from 'react';
import { mockSucursales } from '../data/mockData';
import '../styles/Sucursales.css';

const Sucursales = () => {
    const [sucursales, setSucursales] = useState([]);
    const [loading, setLoading] = useState(true);
    const [sucursalSeleccionada, setSucursalSeleccionada] = useState(null);

    useEffect(() => {
        // Simular carga de API
        setTimeout(() => {
            setSucursales(mockSucursales);
            setLoading(false);
        }, 500);
    }, []);

    const handleVerMapa = (direccion) => {
        const query = encodeURIComponent(direccion);
        window.open(`https://www.google.com/maps/search/?api=1&query=${query}`, '_blank');
    };

    const handleLlamar = (telefono) => {
        window.location.href = `tel:${telefono}`;
    };

    const handleEmail = (email) => {
        window.location.href = `mailto:${email}`;
    };

    if (loading) {
        return (
            <div className="sucursales-container">
                <div className="loading-message">Cargando sucursales...</div>
            </div>
        );
    }

    return (
        <div className="sucursales-container">
            <div className="sucursales-header">
                <h1>Nuestras Sucursales</h1>
                <p>Encontrá la sucursal más cercana a vos</p>
            </div>

            <div className="sucursales-grid">
                {sucursales.map((sucursal) => (
                    <div
                        key={sucursal.id}
                        className={`sucursal-card ${sucursal.destacada ? 'destacada' : ''}`}
                    >
                        {sucursal.destacada && (
                            <div className="destacada-badge">⭐ Destacada</div>
                        )}

                        <div className="sucursal-imagen">
                            <img
                                src={sucursal.imagen}
                                alt={sucursal.nombre}
                                onError={(e) => {
                                    e.target.src = "https://via.placeholder.com/800x400?text=Gimnasio";
                                }}
                            />
                        </div>

                        <div className="sucursal-content">
                            <h2>{sucursal.nombre}</h2>

                            <div className="sucursal-info">
                                <div className="info-item">
                                    <span className="info-icon">📍</span>
                                    <span className="info-texto">{sucursal.direccion}</span>
                                </div>

                                <div className="info-item">
                                    <span className="info-icon">📞</span>
                                    <span className="info-texto">{sucursal.telefono}</span>
                                </div>

                                <div className="info-item">
                                    <span className="info-icon">📧</span>
                                    <span className="info-texto">{sucursal.email}</span>
                                </div>

                                <div className="info-item horarios">
                                    <span className="info-icon">🕐</span>
                                    <span className="info-texto">{sucursal.horarios}</span>
                                </div>
                            </div>

                            <div className="sucursal-servicios">
                                <h3>Servicios disponibles:</h3>
                                <div className="servicios-lista">
                                    {sucursal.servicios.map((servicio, index) => (
                                        <span key={index} className="servicio-tag">
                                            ✓ {servicio}
                                        </span>
                                    ))}
                                </div>
                            </div>

                            <div className="sucursal-acciones">
                                <button
                                    className="btn-accion"
                                    onClick={() => handleVerMapa(sucursal.direccion)}
                                >
                                    📍 Mapa
                                </button>
                                <button
                                    className="btn-accion"
                                    onClick={() => handleLlamar(sucursal.telefono)}
                                >
                                    📞 Llamar
                                </button>
                                <button
                                    className="btn-accion"
                                    onClick={() => handleEmail(sucursal.email)}
                                >
                                    📧 Email
                                </button>
                            </div>

                            <button
                                className="btn-ver-detalle"
                                onClick={() => setSucursalSeleccionada(sucursal)}
                            >
                                Ver más información
                            </button>
                        </div>
                    </div>
                ))}
            </div>

            <div className="sucursales-info-adicional">
                <div className="info-box">
                    <h3>🎯 ¿Primera vez?</h3>
                    <p>Visitá cualquiera de nuestras sucursales y obtené una clase de prueba gratuita</p>
                </div>
                <div className="info-box">
                    <h3>🚗 Estacionamiento</h3>
                    <p>Todas nuestras sucursales cuentan con estacionamiento disponible</p>
                </div>
                <div className="info-box">
                    <h3>♿ Accesibilidad</h3>
                    <p>Instalaciones completamente accesibles para personas con movilidad reducida</p>
                </div>
            </div>

            {/* Modal de detalle de sucursal */}
            {sucursalSeleccionada && (
                <div className="modal-overlay" onClick={() => setSucursalSeleccionada(null)}>
                    <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                        <button
                            className="modal-close"
                            onClick={() => setSucursalSeleccionada(null)}
                        >
                            ✕
                        </button>

                        <div className="modal-header">
                            <h2>{sucursalSeleccionada.nombre}</h2>
                        </div>

                        <div className="modal-body">
                            <img
                                src={sucursalSeleccionada.imagen}
                                alt={sucursalSeleccionada.nombre}
                                className="modal-imagen"
                            />

                            <div className="modal-info">
                                <h3>Información de Contacto</h3>
                                <p><strong>Dirección:</strong> {sucursalSeleccionada.direccion}</p>
                                <p><strong>Teléfono:</strong> {sucursalSeleccionada.telefono}</p>
                                <p><strong>Email:</strong> {sucursalSeleccionada.email}</p>
                                <p><strong>Horarios:</strong> {sucursalSeleccionada.horarios}</p>
                            </div>

                            <div className="modal-servicios">
                                <h3>Servicios e Instalaciones</h3>
                                <ul>
                                    {sucursalSeleccionada.servicios.map((servicio, index) => (
                                        <li key={index}>✓ {servicio}</li>
                                    ))}
                                </ul>
                            </div>

                            <div className="modal-acciones">
                                <button
                                    className="btn-modal-accion"
                                    onClick={() => handleVerMapa(sucursalSeleccionada.direccion)}
                                >
                                    📍 Cómo Llegar
                                </button>
                                <button
                                    className="btn-modal-accion"
                                    onClick={() => handleLlamar(sucursalSeleccionada.telefono)}
                                >
                                    📞 Contactar
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Sucursales;
