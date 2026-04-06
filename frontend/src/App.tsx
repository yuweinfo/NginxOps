import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { AuthProvider } from './contexts/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'
import MainLayout from './views/MainLayout'
import Login from './views/Login'
import Welcome from './views/Welcome'
import Dashboard from './views/Dashboard'
import Sites from './views/Sites'
import LoadBalancer from './views/LoadBalancer'
import Certificates from './views/Certificates'
import Logs from './views/Logs'
import Control from './views/Control'
import Profile from './views/Profile'
import Audit from './views/Audit'
import { Toaster } from './components/ui/toaster'
import { setupApi } from './api/setup'
import { Loader2 } from 'lucide-react'

// 配置状态检查组件
function SetupGuard({ children }: { children: React.ReactNode }) {
  const [checking, setChecking] = useState(true)
  const [configured, setConfigured] = useState<boolean | null>(null)
  const navigate = useNavigate()

  useEffect(() => {
    const checkSetup = async () => {
      try {
        const response = await setupApi.getStatus()
        setConfigured(response.data.configured)
        if (!response.data.configured) {
          navigate('/welcome', { replace: true })
        }
      } catch (error) {
        console.error('Failed to check setup status:', error)
        // 如果检查失败，假设未配置
        setConfigured(false)
        navigate('/welcome', { replace: true })
      } finally {
        setChecking(false)
      }
    }
    checkSetup()
  }, [navigate])

  if (checking) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <div className="text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto text-neutral-400" />
          <p className="mt-4 text-sm text-neutral-500">检查系统状态...</p>
        </div>
      </div>
    )
  }

  return configured ? <>{children}</> : null
}

export default function App() {
  const [dark, setDark] = useState(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('theme')
      if (saved) return saved === 'dark'
      return window.matchMedia('(prefers-color-scheme: dark)').matches
    }
    return true
  })

  useEffect(() => {
    document.documentElement.classList.toggle('dark', dark)
    localStorage.setItem('theme', dark ? 'dark' : 'light')
  }, [dark])

  const toggleTheme = () => setDark(!dark)

  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Welcome 页面 - 无需检查配置 */}
          <Route path="/welcome" element={<Welcome />} />
          
          {/* Login 页面 */}
          <Route path="/login" element={<Login />} />
          
          {/* 需要配置检查的路由 */}
          <Route
            path="/"
            element={
              <SetupGuard>
                <ProtectedRoute>
                  <MainLayout dark={dark} toggleTheme={toggleTheme} />
                </ProtectedRoute>
              </SetupGuard>
            }
          >
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="sites" element={<Sites />} />
            <Route path="loadbalancer" element={<LoadBalancer />} />
            <Route path="certificates" element={<Certificates />} />
            <Route path="logs" element={<Logs />} />
            <Route path="control" element={<Control />} />
            <Route path="profile" element={<Profile />} />
            <Route path="audit" element={<Audit />} />
          </Route>
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
        <Toaster />
      </AuthProvider>
    </BrowserRouter>
  )
}
