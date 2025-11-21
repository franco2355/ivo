import '../styles/Home.css'
import gymPortada from '../assets/login/gimnasio1.jpeg'
import { useNavigate } from 'react-router-dom'

const Home = () => {
    const navigate = useNavigate();

    return (
        <div className="home-container">
            {/* Hero Section */}
            <div className="hero-section">
                <img
                    className="img-gym"
                    src={gymPortada}
                    alt="Gimnasio portada"
                />
                <div className="hero-overlay"></div>
                <div className="hero-content">
                    <h1 className="hero-title">Transforma Tu Cuerpo</h1>
                    <p className="hero-subtitle">El mejor gimnasio de la ciudad te espera</p>
                    <div className="hero-buttons">
                        <button
                            className="btn-primary-hero"
                            onClick={() => navigate('/actividades')}
                        >
                            Ver Actividades
                        </button>
                        <button
                            className="btn-secondary-hero"
                            onClick={() => navigate('/planes')}
                        >
                            Nuestros Planes
                        </button>
                    </div>
                </div>
            </div>

            {/* Features Section */}
            <div className="features-section">
                <h2 className="section-title">¿Por Qué Elegirnos?</h2>
                <div className="features-grid">
                    <div className="feature-card">
                        <div className="feature-icon">🏋️</div>
                        <h3>Equipamiento Moderno</h3>
                        <p>Maquinaria de última generación para tu entrenamiento</p>
                    </div>
                    <div className="feature-card">
                        <div className="feature-icon">👥</div>
                        <h3>Entrenadores Certificados</h3>
                        <p>Profesionales que te guiarán en cada paso</p>
                    </div>
                    <div className="feature-card">
                        <div className="feature-icon">⏰</div>
                        <h3>Horarios Flexibles</h3>
                        <p>Abierto desde las 6 AM hasta las 11 PM</p>
                    </div>
                    <div className="feature-card">
                        <div className="feature-icon">📍</div>
                        <h3>Múltiples Sucursales</h3>
                        <p>Encontrá la más cercana a vos</p>
                    </div>
                </div>
            </div>

            {/* CTA Section */}
            <div className="cta-section">
                <div className="cta-content">
                    <h2>¿Listo para empezar?</h2>
                    <p>Unite hoy y obtené tu primera clase gratis</p>
                    <button
                        className="btn-cta-large"
                        onClick={() => navigate('/register')}
                    >
                        Registrate Ahora
                    </button>
                </div>
            </div>
        </div>
    );
};

export default Home;