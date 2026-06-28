import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Subdomains from './pages/Subdomains'
import Health from './pages/Health'
import DNS from './pages/DNS'
import Settings from './pages/Settings'

function Protected({ children }) {
  return <Layout>{children}</Layout>
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<Protected><Dashboard /></Protected>} />
        <Route path="/subdomains" element={<Protected><Subdomains /></Protected>} />
        <Route path="/health" element={<Protected><Health /></Protected>} />
        <Route path="/dns" element={<Protected><DNS /></Protected>} />
        <Route path="/settings" element={<Protected><Settings /></Protected>} />
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </BrowserRouter>
  )
}
