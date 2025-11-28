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

        </div>
    );
};

export default Home;