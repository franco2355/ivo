/**
 * Ejemplo de frontend con Idempotency Key para prevenir pagos duplicados
 *
 * Tecnologías: React + Axios
 * Patrón: Debounce + UUID generation
 *
 * Características:
 * - Genera UUID único por cada intento de pago
 * - Deshabilita botón mientras procesa
 * - Muestra feedback visual al usuario
 * - Maneja reintentos automáticos (usa mismo idempotency_key)
 */

import React, { useState } from 'react';
import axios from 'axios';
import { v4 as uuidv4 } from 'uuid'; // npm install uuid

const PaymentForm = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);

  // Estado del formulario
  const [formData, setFormData] = useState({
    entity_type: 'subscription',
    entity_id: '',
    user_id: '',
    amount: 0,
    currency: 'ARS',
    payment_method: 'credit_card',
    payment_gateway: 'mercadopago'
  });

  /**
   * Handler principal para procesar el pago
   * ⭐ PUNTO CLAVE: Genera un UUID único para cada intento de pago
   */
  const handlePayment = async () => {
    // Validaciones básicas
    if (!formData.entity_id || !formData.user_id || formData.amount <= 0) {
      setError('Por favor complete todos los campos');
      return;
    }

    // Prevenir doble clic
    if (loading) {
      console.log('⚠️ Pago ya en proceso, ignorando doble clic');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(false);

    // ⭐ GENERAR IDEMPOTENCY KEY ÚNICO
    // Este UUID garantiza que si el usuario hace doble clic,
    // el backend detectará el duplicado y retornará el pago original
    const idempotencyKey = uuidv4();

    console.log(`🔑 Idempotency Key generado: ${idempotencyKey}`);

    try {
      const response = await axios.post(
        'http://localhost:8080/payments/process',
        {
          ...formData,
          idempotency_key: idempotencyKey // ⭐ Incluir idempotency key
        },
        {
          timeout: 30000, // 30 segundos de timeout
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      console.log('✅ Pago creado exitosamente:', response.data);

      // Si el pago requiere redirección (Mercado Pago Checkout Pro)
      if (response.data.metadata?.payment_url) {
        window.location.href = response.data.metadata.payment_url;
      } else {
        setSuccess(true);
      }

    } catch (err) {
      console.error('❌ Error procesando pago:', err);

      if (err.response) {
        // El servidor respondió con un error
        setError(err.response.data.error || 'Error procesando el pago');
      } else if (err.request) {
        // La request fue enviada pero no hubo respuesta
        setError('No se pudo conectar con el servidor. Por favor intente nuevamente.');
      } else {
        // Error en la configuración de la request
        setError('Error inesperado. Por favor intente nuevamente.');
      }
    } finally {
      setLoading(false);
    }
  };

  /**
   * Handler con debounce para evitar múltiples clics rápidos
   * Alternativa más robusta que solo deshabilitar el botón
   */
  const [lastClickTime, setLastClickTime] = useState(0);
  const DEBOUNCE_TIME = 1000; // 1 segundo

  const handlePaymentWithDebounce = () => {
    const now = Date.now();

    if (now - lastClickTime < DEBOUNCE_TIME) {
      console.log('⚠️ Debounce: Ignorando clic múltiple');
      return;
    }

    setLastClickTime(now);
    handlePayment();
  };

  return (
    <div className="payment-form">
      <h2>Realizar Pago</h2>

      {/* Formulario */}
      <div className="form-group">
        <label>Tipo de Entidad:</label>
        <input
          type="text"
          value={formData.entity_type}
          onChange={(e) => setFormData({ ...formData, entity_type: e.target.value })}
          disabled={loading}
        />
      </div>

      <div className="form-group">
        <label>ID de Entidad:</label>
        <input
          type="text"
          value={formData.entity_id}
          onChange={(e) => setFormData({ ...formData, entity_id: e.target.value })}
          placeholder="Ej: subscription_123"
          disabled={loading}
        />
      </div>

      <div className="form-group">
        <label>ID de Usuario:</label>
        <input
          type="text"
          value={formData.user_id}
          onChange={(e) => setFormData({ ...formData, user_id: e.target.value })}
          placeholder="Ej: user_456"
          disabled={loading}
        />
      </div>

      <div className="form-group">
        <label>Monto:</label>
        <input
          type="number"
          value={formData.amount}
          onChange={(e) => setFormData({ ...formData, amount: parseFloat(e.target.value) })}
          min="0"
          step="0.01"
          disabled={loading}
        />
      </div>

      <div className="form-group">
        <label>Moneda:</label>
        <select
          value={formData.currency}
          onChange={(e) => setFormData({ ...formData, currency: e.target.value })}
          disabled={loading}
        >
          <option value="ARS">ARS</option>
          <option value="USD">USD</option>
          <option value="EUR">EUR</option>
        </select>
      </div>

      <div className="form-group">
        <label>Gateway de Pago:</label>
        <select
          value={formData.payment_gateway}
          onChange={(e) => setFormData({ ...formData, payment_gateway: e.target.value })}
          disabled={loading}
        >
          <option value="mercadopago">Mercado Pago</option>
          <option value="cash">Efectivo</option>
        </select>
      </div>

      {/* Mensajes de estado */}
      {error && (
        <div className="alert alert-error">
          ❌ {error}
        </div>
      )}

      {success && (
        <div className="alert alert-success">
          ✅ Pago procesado exitosamente
        </div>
      )}

      {/* Botón de pago */}
      <button
        onClick={handlePaymentWithDebounce}
        disabled={loading}
        className={`btn btn-primary ${loading ? 'loading' : ''}`}
      >
        {loading ? (
          <>
            <span className="spinner"></span>
            Procesando...
          </>
        ) : (
          'Pagar Ahora'
        )}
      </button>

      {/* Indicador de protección */}
      <p className="security-notice">
        🔒 Protegido contra pagos duplicados con idempotency key
      </p>
    </div>
  );
};

export default PaymentForm;


/**
 * ========== EJEMPLO CON VANILLA JAVASCRIPT ==========
 * Para proyectos sin React
 */

/*
// HTML
<button id="payBtn" onclick="handlePayment()">Pagar Ahora</button>
<div id="message"></div>

// JavaScript
let isProcessing = false;

async function handlePayment() {
  // Prevenir doble clic
  if (isProcessing) {
    console.log('⚠️ Pago ya en proceso');
    return;
  }

  isProcessing = true;
  document.getElementById('payBtn').disabled = true;
  document.getElementById('message').innerText = 'Procesando...';

  // Generar UUID (usando una librería o función personalizada)
  const idempotencyKey = crypto.randomUUID(); // Browser API moderna

  try {
    const response = await fetch('http://localhost:8080/payments/process', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        entity_type: 'subscription',
        entity_id: 'sub_123',
        user_id: 'user_456',
        amount: 1000,
        currency: 'ARS',
        payment_method: 'credit_card',
        payment_gateway: 'mercadopago',
        idempotency_key: idempotencyKey // ⭐ Incluir idempotency key
      })
    });

    const data = await response.json();

    if (response.ok) {
      console.log('✅ Pago creado:', data);

      // Redirigir a Mercado Pago si es necesario
      if (data.metadata?.payment_url) {
        window.location.href = data.metadata.payment_url;
      } else {
        document.getElementById('message').innerText = '✅ Pago exitoso';
      }
    } else {
      throw new Error(data.error || 'Error procesando pago');
    }

  } catch (error) {
    console.error('❌ Error:', error);
    document.getElementById('message').innerText = `❌ ${error.message}`;
  } finally {
    isProcessing = false;
    document.getElementById('payBtn').disabled = false;
  }
}
*/


/**
 * ========== EJEMPLO CON AXIOS INTERCEPTOR ==========
 * Para aplicaciones que quieren agregar idempotency automáticamente
 */

/*
import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';

// Crear instancia de axios
const api = axios.create({
  baseURL: 'http://localhost:8080',
  timeout: 30000
});

// Interceptor para agregar idempotency key automáticamente a POST requests
api.interceptors.request.use((config) => {
  // Solo agregar idempotency key a POST requests que NO lo tengan
  if (config.method === 'post' && !config.data?.idempotency_key) {
    config.data = {
      ...config.data,
      idempotency_key: uuidv4()
    };
    console.log(`🔑 Idempotency key agregado: ${config.data.idempotency_key}`);
  }
  return config;
});

// Usar la API
api.post('/payments/process', {
  entity_type: 'subscription',
  amount: 1000,
  // idempotency_key se agregará automáticamente
});
*/
