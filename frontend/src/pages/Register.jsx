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
    const [fieldErrors, setFieldErrors] = useState({});
    const [touched, setTouched] = useState({});
    const navigate = useNavigate();
    const toast = useToastContext();

    const validateField = (name, value) => {
        let error = "";

        switch (name) {
            case "nombre":
            case "apellido":
                if (!value.trim()) {
                    error = `El ${name} es obligatorio`;
                } else if (value.length > 30) {
                    error = `El ${name} debe tener máximo 30 caracteres`;
                }
                break;

            case "username":
                if (!value.trim()) {
                    error = "El nombre de usuario es obligatorio";
                } else if (value.length < 3 || value.length > 30) {
                    error = "Debe tener entre 3 y 30 caracteres";
                } else if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
                    error = "Solo letras, números, guiones y guiones bajos";
                }
                break;

            case "email":
                if (!value.trim()) {
                    error = "El email es obligatorio";
                } else if (!/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(value)) {
                    error = "Formato de email inválido";
                }
                break;

            case "password":
                const passwordErrors = validatePassword(value);
                if (passwordErrors.length > 0) {
                    error = `Debe tener ${passwordErrors.join(", ")}`;
                }
                break;

            case "confirmPassword":
                if (value && value !== formData.password) {
                    error = "Las contraseñas no coinciden";
                }
                break;

            default:
                break;
        }

        return error;
    };

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prevState => ({
            ...prevState,
            [name]: value
        }));

        // Validar campo si ya fue tocado
        if (touched[name]) {
            const error = validateField(name, value);
            setFieldErrors(prev => ({
                ...prev,
                [name]: error
            }));
        }

        // Validar confirmPassword cuando cambia password
        if (name === "password" && touched.confirmPassword) {
            const confirmError = formData.confirmPassword && formData.confirmPassword !== value ?
                "Las contraseñas no coinciden" : "";
            setFieldErrors(prev => ({
                ...prev,
                confirmPassword: confirmError
            }));
        }
    };

    const handleBlur = (e) => {
        const { name, value } = e.target;
        setTouched(prev => ({
            ...prev,
            [name]: true
        }));

        const error = validateField(name, value);
        setFieldErrors(prev => ({
            ...prev,
            [name]: error
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
        if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
            errors.push("un carácter especial (!@#$%^&*(),.?\":{}|<>)");
        }

        return errors;
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setIsLoading(true);
        setError("");

        // Validar todos los campos
        const errors = {};
        Object.keys(formData).forEach(key => {
            const error = validateField(key, formData[key]);
            if (error) {
                errors[key] = error;
            }
        });

        // Si hay errores, marcar todos los campos como tocados y mostrar errores
        if (Object.keys(errors).length > 0) {
            setFieldErrors(errors);
            setTouched({
                nombre: true,
                apellido: true,
                username: true,
                email: true,
                password: true,
                confirmPassword: true
            });
            setError("Por favor, corrige los errores en el formulario");
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
                // El backend ahora devuelve mensajes en español
                let errorMessage = errorData.error || "Error al registrar usuario";

                // Si el error es específico de un campo, marcarlo
                if (errorData.details) {
                    if (errorData.details.includes("username_already_exists")) {
                        setFieldErrors(prev => ({
                            ...prev,
                            username: "Este nombre de usuario ya está en uso"
                        }));
                    } else if (errorData.details.includes("email_already_exists")) {
                        setFieldErrors(prev => ({
                            ...prev,
                            email: "Este email ya está registrado"
                        }));
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

                {/* Fila 1: Nombre y Apellido */}
                <div className="input-row">
                    <div className="input-group">
                        <input
                            type="text"
                            name="nombre"
                            placeholder="Nombre"
                            value={formData.nombre}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            disabled={isLoading}
                            required
                            className={fieldErrors.nombre && touched.nombre ? 'input-error' : ''}
                        />
                        {fieldErrors.nombre && touched.nombre && (
                            <small className="field-error">{fieldErrors.nombre}</small>
                        )}
                    </div>

                    <div className="input-group">
                        <input
                            type="text"
                            name="apellido"
                            placeholder="Apellido"
                            value={formData.apellido}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            disabled={isLoading}
                            required
                            className={fieldErrors.apellido && touched.apellido ? 'input-error' : ''}
                        />
                        {fieldErrors.apellido && touched.apellido && (
                            <small className="field-error">{fieldErrors.apellido}</small>
                        )}
                    </div>
                </div>

                {/* Fila 2: Usuario y Email */}
                <div className="input-row">
                    <div className="input-group">
                        <input
                            type="text"
                            name="username"
                            placeholder="Usuario"
                            value={formData.username}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            disabled={isLoading}
                            required
                            className={fieldErrors.username && touched.username ? 'input-error' : ''}
                        />
                        {fieldErrors.username && touched.username && (
                            <small className="field-error">{fieldErrors.username}</small>
                        )}
                    </div>

                    <div className="input-group">
                        <input
                            type="email"
                            name="email"
                            placeholder="Email"
                            value={formData.email}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            disabled={isLoading}
                            required
                            className={fieldErrors.email && touched.email ? 'input-error' : ''}
                        />
                        {fieldErrors.email && touched.email && (
                            <small className="field-error">{fieldErrors.email}</small>
                        )}
                    </div>
                </div>

                {/* Fila 3: Contraseña y Confirmar Contraseña */}
                <div className="input-row">
                    <div className="input-group">
                        <input
                            type="password"
                            name="password"
                            placeholder="Contraseña"
                            value={formData.password}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            disabled={isLoading}
                            required
                            minLength={8}
                            className={fieldErrors.password && touched.password ? 'input-error' : ''}
                        />
                        {fieldErrors.password && touched.password && (
                            <small className="field-error">{fieldErrors.password}</small>
                        )}
                    </div>

                    <div className="input-group">
                        <input
                            type="password"
                            name="confirmPassword"
                            placeholder="Confirmar Contraseña"
                            value={formData.confirmPassword}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            disabled={isLoading}
                            required
                            className={fieldErrors.confirmPassword && touched.confirmPassword ? 'input-error' : ''}
                        />
                        {fieldErrors.confirmPassword && touched.confirmPassword && (
                            <small className="field-error">{fieldErrors.confirmPassword}</small>
                        )}
                    </div>
                </div>

                {/* Hint de contraseña (debajo de las contraseñas) */}
                <small className={fieldErrors.password && touched.password ? "password-hint error" : "password-hint"}>
                    {fieldErrors.password && touched.password ? fieldErrors.password : "Debe tener: 8+ caracteres, mayúscula, minúscula, número y carácter especial"}
                </small>

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