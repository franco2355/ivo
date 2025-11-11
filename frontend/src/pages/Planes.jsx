import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { mockPlanes } from '../data/mockData';
import '../styles/Planes.css';

const Planes = () => {
    const [planes, setPlanes] = useState([]);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();
    const isLoggedIn = localStorage.getItem("isLoggedIn") === "true";

    useEffect(() => {
        // Simular carga de API
        setTimeout(() => {
            setPlanes(mockPlanes);
            setLoading(false);
        }, 500);
    }, []);

    const handleSelectPlan = (planId) => {
        if (!isLoggedIn) {
            alert("Debes iniciar sesión para suscribirte a un plan");
            navigate('/login');
            return;
        }
        // Navegar al checkout con el plan seleccionado
        navigate(`/checkout/${planId}`);
    };

    if (loading) {
        return (
            <div className="planes-container">
                <div className="loading-message">Cargando planes...</div>
            </div>
        );
    }

    return (
        <div className="planes-container">
            <div className="planes-header">
                <h1>Nuestros Planes</h1>
                <p>Elegí el plan que mejor se adapte a tus necesidades</p>
            </div>

            <div className="planes-grid">
                {planes.map((plan) => (
                    <div
                        key={plan.id}
                        className={`plan-card ${plan.popular ? 'popular' : ''}`}
                    >
                        {plan.popular && <div className="popular-badge">Más Popular</div>}
                        {plan.ahorro && <div className="ahorro-badge">{plan.ahorro}</div>}

                        <div className="plan-header" style={{ borderTopColor: plan.color }}>
                            <h2>{plan.nombre}</h2>
                            <p className="plan-descripcion">{plan.descripcion}</p>
                        </div>

                        <div className="plan-precio">
                            <span className="precio-signo">$</span>
                            <span className="precio-monto">{plan.precio_mensual.toFixed(2)}</span>
                            <span className="precio-periodo">/ mes</span>
                        </div>

                        {plan.duracion_dias > 30 && (
                            <div className="plan-duracion">
                                Compromiso de {plan.duracion_dias} días
                            </div>
                        )}

                        <div className="plan-beneficios">
                            <h3>Beneficios incluidos:</h3>
                            <ul>
                                {plan.beneficios.map((beneficio, index) => (
                                    <li key={index}>
                                        <span className="check-icon">✓</span>
                                        {beneficio}
                                    </li>
                                ))}
                            </ul>
                        </div>

                        <button
                            className="btn-seleccionar-plan"
                            style={{ backgroundColor: plan.color }}
                            onClick={() => handleSelectPlan(plan.id)}
                        >
                            Seleccionar Plan
                        </button>
                    </div>
                ))}
            </div>

            <div className="planes-footer">
                <div className="info-adicional">
                    <h3>¿Necesitás ayuda para elegir?</h3>
                    <p>Contactanos al <strong>0351-123-4567</strong> o visitá cualquiera de nuestras sucursales</p>
                    <button
                        className="btn-ver-sucursales"
                        onClick={() => navigate('/sucursales')}
                    >
                        Ver Sucursales
                    </button>
                </div>
            </div>
        </div>
    );
};

export default Planes;
