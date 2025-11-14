import { useNavigate } from "react-router-dom";
import "../styles/Header.css";


const Header = ( ) => {
    const isLoggedIn = localStorage.getItem("isLoggedIn") === "true";
    const isAdmin = localStorage.getItem("isAdmin") === "true";
    const navigate = useNavigate();
    const logout = () => {
        localStorage.removeItem("isLoggedIn");
        localStorage.removeItem("isAdmin");
        localStorage.removeItem("access_token");
        localStorage.removeItem("idUsuario");
        navigate("/");
    }

    return (
        <header>
            <div className="header-container">
                <nav className="header-content">
                    <h1 className="header-title" onClick={() => navigate("/")}>GymPro</h1>
                    <div className="header-links">
                        {isLoggedIn && !isAdmin && (
                            <a href="/dashboard">Dashboard</a>
                        )}
                        <a href="/actividades">Actividades</a>
                        <a href="/planes">Planes</a>
                        <a href="/sucursales">Sucursales</a>
                        {isLoggedIn && !isAdmin && (
                            <>
                                <a href="/mi-suscripcion">Mi Suscripción</a>
                                <a href="/pagos">Pagos</a>
                            </>
                        )}
                        {isAdmin && (
                            <a href="/admin">Panel Admin</a>
                        )}
                        {isLoggedIn ? (
                            <button onClick={logout}>Cerrar sesión</button>
                        ) : (
                            <a href="/login">Iniciar Sesión</a>
                        )}
                    </div>
                </nav>
            </div>
        </header>
    );
}

export default Header;