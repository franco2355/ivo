import Toast from './Toast';
import '../styles/ToastContainer.css';

const ToastContainer = ({ toasts, onRemove }) => {
    return (
        <div className="toast-container">
            {toasts.map((toast, index) => (
                <div
                    key={toast.id}
                    className="toast-wrapper"
                    style={{ top: `${20 + index * 90}px` }}
                >
                    <Toast
                        message={toast.message}
                        type={toast.type}
                        duration={toast.duration}
                        onClose={() => onRemove(toast.id)}
                    />
                </div>
            ))}
        </div>
    );
};

export default ToastContainer;
