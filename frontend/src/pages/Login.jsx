import { useState } from "react";
import '../styles/Login.css';
import { useNavigate } from "react-router-dom";
import { USERS_API } from '../config/api';

const getTokenPayload = (token) => {
    if (!token) return null; 
    const parts = token.split('.');
    const decodedPaylod = atob(parts[1]);

    return JSON.parse(decodedPaylod);
}

const storeUserSession = (accessToken) => {
    const payload = getTokenPayload(accessToken)
    if (!payload) return;
    const admin = payload.is_admin;
    const idUsuario = payload.id_usuario;
    const username = payload.username;

    localStorage.setItem("access_token", accessToken);
    localStorage.setItem("idUsuario", parseInt(idUsuario));
    localStorage.setItem("isAdmin", admin.toString());
    localStorage.setItem("isLoggedIn", "true");
    localStorage.setItem("nombre", username);
};

const Login = () => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");
    const navigate = useNavigate();

    const handlerLogin = async (e) => {
        e.preventDefault();
        setIsLoading(true);
        setError("");

        try {
            const response = await fetch(USERS_API.login, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    username_or_email: username.trim(),
                    password: password
                })
            });

            if (response.ok) {
                const data = await response.json();

                if (!data.token) {
                    setError("No se recibió ningún token del servidor");
                    return;
                }

                // Guardar datos adicionales del usuario
                if (data.user && data.user.nombre) {
                    localStorage.setItem("nombre", `${data.user.nombre} ${data.user.apellido || ''}`);
                }

                storeUserSession(data.token);

                navigate("/");
            } else {
                const errorData = await response.json();
                if (response.status === 401) {
                    setError("Usuario o contraseña incorrectos");
                } else {
                    setError(errorData.error || "Error de autenticación");
                }
            }

        } catch (error) {
            setError("Error de conexión");
            console.error("Error de conexión:", error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleBack = () => {
        navigate('/');
    };

    return (
        <div className="login-container">
            <button onClick={handleBack} className="back-button">
                ← Inicio
            </button>
            <form className="login-form" onSubmit={handlerLogin}>
                <h2>Iniciar Sesión</h2>

                {error && <div className="error-message">{error}</div>}

                <div className="input-group">
                    <input
                        type="text"
                        placeholder="Email o Usuario"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        disabled={isLoading}
                        required
                    />
                </div>

                <div className="input-group">
                    <input
                        type="password"
                        placeholder="Contraseña"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        disabled={isLoading}
                        required
                        minLength={8}
                    />
                </div>

                <button type="submit" disabled={isLoading}>
                    {isLoading ? "Ingresando..." : "Ingresar"}
                </button>

                <div className="register-link">
                    ¿No tienes una cuenta? <a href="/register">Regístrate ahora</a>
                </div>
            </form>
        </div>
    );
};

export { storeUserSession };
export default Login;