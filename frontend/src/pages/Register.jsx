import { useState } from "react";
import '../styles/Register.css';
import { useNavigate } from "react-router-dom";
import { storeUserSession } from "./Login";
import { useToastContext } from '../context/ToastContext';
import { USERS_API } from '../config/api';

const Register = () => {
    const [formData, setFormData] = useState({
        nombre: "",
        apellido: "",
        username: "",
        email: "",
        password: "",
        confirmPassword: ""
    });
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");
    const navigate = useNavigate();
    const toast = useToastContext();

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prevState => ({
            ...prevState,
            [name]: value
        }));
    };

    const validatePassword = (password) => {
        const errors = [];

        if (password.length < 8) {
            errors.push("al menos 8 caracteres");
        }
        if (!/[A-Z]/.test(password)) {
            errors.push("una letra mayúscula");
        }
        if (!/[a-z]/.test(password)) {
            errors.push("una letra minúscula");
        }
        if (!/[0-9]/.test(password)) {
            errors.push("un número");
        }

        return errors;
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setIsLoading(true);
        setError("");

        // Validar contraseña
        const passwordErrors = validatePassword(formData.password);
        if (passwordErrors.length > 0) {
            setError(`La contraseña debe tener ${passwordErrors.join(", ")}`);
            setIsLoading(false);
            return;
        }

        if (formData.password !== formData.confirmPassword) {
            setError("Las contraseñas no coinciden");
            setIsLoading(false);
            return;
        }

        try {
            const response = await fetch(USERS_API.register, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    nombre: formData.nombre.trim(),
                    apellido: formData.apellido.trim(),
                    username: formData.username.trim(), 
                    email: formData.email.trim(),       
                    password: formData.password.trim() 
                })
            });

            if (response.ok) {
                const data = await response.json();

                // Guardar nombre completo del usuario
                localStorage.setItem("nombre", `${formData.nombre} ${formData.apellido}`);

                storeUserSession(data.token);
                toast.success("Usuario registrado exitosamente");
                navigate("/");
            } else {
                const errorData = await response.json();
                // Mejorar el mensaje de error del backend
                let errorMessage = errorData.error || "Error al registrar usuario";

                if (errorData.details) {
                    // Extraer información útil de los detalles
                    if (errorData.details.includes("username")) {
                        errorMessage = "El nombre de usuario ya está en uso";
                    } else if (errorData.details.includes("email")) {
                        errorMessage = "El email ya está registrado";
                    } else if (errorData.details.includes("Password")) {
                        errorMessage = "La contraseña no cumple con los requisitos";
                    }
                }

                setError(errorMessage);
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
        <div className="register-container">
            <button onClick={handleBack} className="back-button">
                ← Inicio
            </button>
            <form className="register-form" onSubmit={handleSubmit}>
                <h2>Registro de Usuario</h2>

                {error && <div className="error-message">{error}</div>}

                <div className="input-group">
                    <input
                        type="text"
                        name="nombre"
                        placeholder="Nombre"
                        value={formData.nombre}
                        onChange={handleChange}
                        disabled={isLoading}
                        required
                    />
                </div>

                <div className="input-group">
                    <input
                        type="text"
                        name="apellido"
                        placeholder="Apellido"
                        value={formData.apellido}
                        onChange={handleChange}
                        disabled={isLoading}
                        required
                    />
                </div>

                <div className="input-group">
                    <input
                        type="text"
                        name="username"
                        placeholder="Usuario"
                        value={formData.username}
                        onChange={handleChange}
                        disabled={isLoading}
                        required
                    />
                </div>
                <div className="input-group">
                    <input
                        type="email"
                        name="email"
                        placeholder="Email"
                        value={formData.email}
                        onChange={handleChange}
                        disabled={isLoading}
                        required
                    />
                </div>

                <div className="input-group">
                    <input
                        type="password"
                        name="password"
                        placeholder="Contraseña"
                        value={formData.password}
                        onChange={handleChange}
                        disabled={isLoading}
                        required
                        minLength={8}
                    />
                    <small className="password-hint">
                        Debe tener: 8+ caracteres, mayúscula, minúscula y número
                    </small>
                </div>

                <div className="input-group">
                    <input
                        type="password"
                        name="confirmPassword"
                        placeholder="Confirmar Contraseña"
                        value={formData.confirmPassword}
                        onChange={handleChange}
                        disabled={isLoading}
                        required
                    />
                </div>

                <button type="submit" disabled={isLoading}>
                    {isLoading ? "Registrando..." : "Registrarse"}
                </button>

                <div className="login-link">
                    ¿Ya tienes una cuenta? <a href="/login">Iniciar Sesión</a>
                </div>
            </form>
        </div>
    );
};

export default Register; 