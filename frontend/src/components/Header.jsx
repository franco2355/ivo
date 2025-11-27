import { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import "../styles/Header.css";

const Header = () => {
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const dropdownRef = useRef(null);
    const isLoggedIn = localStorage.getItem("isLoggedIn") === "true";
    const isAdmin = localStorage.getItem("isAdmin") === "true";
    const userName = localStorage.getItem("nombre") || "Usuario";
    const navigate = useNavigate();

    const logout = () => {
        localStorage.removeItem("isLoggedIn");
        localStorage.removeItem("isAdmin");
        localStorage.removeItem("access_token");
        localStorage.removeItem("idUsuario");
        localStorage.removeItem("nombre");
        navigate("/");
    };

    // Cerrar dropdown al hacer click fuera
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setDropdownOpen(false);
            }
        };

        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);

    const handleNavigation = (path) => {
        navigate(path);
        setDropdownOpen(false);
    };

    return (
        <header className="modern-header">
            <div className="header-container">
                <nav className="header-content">
                    <h1 className="header-title" onClick={() => navigate("/")}>
                        <span className="logo-icon">💪</span>
                        GymPro
                    </h1>

                    <div className="header-links">
                        <a href="/actividades" className="nav-link">Actividades</a>
                        <a href="/planes" className="nav-link">Planes</a>
                        <a href="/sucursales" className="nav-link">Sucursales</a>

                        {isAdmin && (
                            <a href="/admin" className="nav-link admin-link">Panel Admin</a>
                        )}

                        {isLoggedIn ? (
                            <div className="user-menu" ref={dropdownRef}>
                                <button
                                    className="user-menu-button"
                                    onClick={() => setDropdownOpen(!dropdownOpen)}
                                >
                                    <div className="user-avatar">
                                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                                            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z"/>
                                        </svg>
                                    </div>
                                    <span className="user-name">{userName}</span>
                                    <svg
                                        className={`dropdown-arrow ${dropdownOpen ? 'open' : ''}`}
                                        xmlns="http://www.w3.org/2000/svg"
                                        viewBox="0 0 24 24"
                                        fill="currentColor"
                                    >
                                        <path d="M7 10l5 5 5-5z"/>
                                    </svg>
                                </button>

                                {dropdownOpen && (
                                    <div className="user-dropdown">
                                        {!isAdmin && (
                                            <>
                                                <button
                                                    className="dropdown-item"
                                                    onClick={() => handleNavigation('/dashboard')}
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                                                        <path d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z"/>
                                                    </svg>
                                                    Dashboard
                                                </button>
                                                <button
                                                    className="dropdown-item"
                                                    onClick={() => handleNavigation('/mi-suscripcion')}
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                                                        <path d="M20 4H4c-1.11 0-1.99.89-1.99 2L2 18c0 1.11.89 2 2 2h16c1.11 0 2-.89 2-2V6c0-1.11-.89-2-2-2zm0 14H4v-6h16v6zm0-10H4V6h16v2z"/>
                                                    </svg>
                                                    Mi Suscripción
                                                </button>
                                                <button
                                                    className="dropdown-item"
                                                    onClick={() => handleNavigation('/pagos')}
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                                                        <path d="M11.8 10.9c-2.27-.59-3-1.2-3-2.15 0-1.09 1.01-1.85 2.7-1.85 1.78 0 2.44.85 2.5 2.1h2.21c-.07-1.72-1.12-3.3-3.21-3.81V3h-3v2.16c-1.94.42-3.5 1.68-3.5 3.61 0 2.31 1.91 3.46 4.7 4.13 2.5.6 3 1.48 3 2.41 0 .69-.49 1.79-2.7 1.79-2.06 0-2.87-.92-2.98-2.1h-2.2c.12 2.19 1.76 3.42 3.68 3.83V21h3v-2.15c1.95-.37 3.5-1.5 3.5-3.55 0-2.84-2.43-3.81-4.7-4.4z"/>
                                                    </svg>
                                                    Pagos
                                                </button>
                                                <div className="dropdown-divider"></div>
                                            </>
                                        )}
                                        <button
                                            className="dropdown-item logout"
                                            onClick={logout}
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                                                <path d="M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.58L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z"/>
                                            </svg>
                                            Cerrar Sesión
                                        </button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <div className="auth-buttons">
                                <a href="/login" className="login-button">
                                    Iniciar Sesión
                                </a>
                                <a href="/register" className="register-button">
                                    Registrarse
                                </a>
                            </div>
                        )}
                    </div>
                </nav>
            </div>
        </header>
    );
};

export default Header;