import { StrictMode, Suspense, lazy } from 'react'
import { createRoot } from 'react-dom/client'
import '../styles/index.css'
import { BrowserRouter, Routes, Route } from "react-router-dom"
import { initializeMockData } from '../data/mockData.js'
import { ToastProvider } from '../context/ToastContext.jsx'
import Spinner from '../components/Spinner.jsx'
import ErrorBoundary from '../components/ErrorBoundary.jsx'

// Lazy loading de componentes para mejorar performance
const Login = lazy(() => import('../pages/Login.jsx'));
const Register = lazy(() => import('../pages/Register.jsx'));
const Actividades = lazy(() => import('../pages/Actividades.jsx'));
const AdminPanel = lazy(() => import('../pages/AdminPanel.jsx'));
const Dashboard = lazy(() => import('../pages/Dashboard.jsx'));
const Planes = lazy(() => import('../pages/Planes.jsx'));
const MiSuscripcion = lazy(() => import('../pages/MiSuscripcion.jsx'));
const Checkout = lazy(() => import('../pages/Checkout.jsx'));
const PagoResultado = lazy(() => import('../pages/PagoResultado.jsx'));
const Pagos = lazy(() => import('../pages/Pagos.jsx'));
const Sucursales = lazy(() => import('../pages/Sucursales.jsx'));
const Layout = lazy(() => import('../components/Layout.jsx'));
const Home = lazy(() => import('../pages/Home.jsx'));

// Inicializar datos mock en localStorage
initializeMockData();

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <ErrorBoundary>
      <ToastProvider>
        <BrowserRouter>
          <Suspense fallback={<Spinner size="large" message="Cargando..." />}>
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />
              <Route path="/" element={<Layout />}>
                <Route index element={<Home />} />
                <Route path="dashboard" element={<Dashboard />} />
                <Route path="actividades" element={<Actividades />} />
                <Route path="planes" element={<Planes />} />
                <Route path="mi-suscripcion" element={<MiSuscripcion />} />
                <Route path="checkout/:planId" element={<Checkout />} />
                <Route path="pago/resultado" element={<PagoResultado />} />
                <Route path="pagos" element={<Pagos />} />
                <Route path="sucursales" element={<Sucursales />} />
                <Route path="admin" element={<AdminPanel />} />
              </Route>
            </Routes>
          </Suspense>
        </BrowserRouter>
      </ToastProvider>
    </ErrorBoundary>
  </StrictMode>,
)