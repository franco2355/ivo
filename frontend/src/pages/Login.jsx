import { useState } from "react";
import '../styles/Login.css';
import { useNavigate } from "react-router-dom";

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
    const idUsuario = payload.id_usuario

    localStorage.setItem("access_token", accessToken);
    localStorage.setItem("idUsuario", parseInt(idUsuario));
    localStorage.setItem("isAdmin", admin.toString());
    localStorage.setItem("isLoggedIn", "true");
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
            const response = await fetch('http://localhost:8080/login', {
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
                
                storeUserSession(data.token)
                
                navigate("/");
            } else {
                const errorData = await response.json();
                setError(errorData.error || "Error de autenticación");
                alert("Error al loguearse");
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
                        placeholder="Usuario"
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