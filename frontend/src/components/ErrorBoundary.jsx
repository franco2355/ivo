import React from 'react';
import '../styles/ErrorBoundary.css';

/**
 * Error Boundary para capturar errores de React y mostrar UI de fallback
 */
class ErrorBoundary extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            hasError: false,
            error: null,
            errorInfo: null
        };
    }

    static getDerivedStateFromError(error) {
        // Actualizar el estado para mostrar la UI de fallback
        return { hasError: true };
    }

    componentDidCatch(error, errorInfo) {
        // Log del error a un servicio de reporting
        console.error('Error capturado por ErrorBoundary:', error, errorInfo);
        this.setState({
            error: error,
            errorInfo: errorInfo
        });
    }

    handleReset = () => {
        this.setState({
            hasError: false,
            error: null,
            errorInfo: null
        });
        // Recargar la página
        window.location.href = '/';
    };

    render() {
        if (this.state.hasError) {
            // UI de fallback personalizada
            return (
                <div className="error-boundary-container">
                    <div className="error-boundary-content">
                        <h1>😕 Algo salió mal</h1>
                        <p>Lo sentimos, ocurrió un error inesperado.</p>
                        <details className="error-details">
                            <summary>Detalles del error</summary>
                            <p className="error-message">{this.state.error && this.state.error.toString()}</p>
                            <pre className="error-stack">
                                {this.state.errorInfo && this.state.errorInfo.componentStack}
                            </pre>
                        </details>
                        <button onClick={this.handleReset} className="error-reset-button">
                            Volver al inicio
                        </button>
                    </div>
                </div>
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;
