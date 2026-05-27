import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from '@/contexts/AuthContext'
import { Home } from '@/pages/sessions/Home'
import { Lobby } from '@/pages/sessions/Lobby'
import { Game } from '@/pages/sessions/Game'
import { Login } from '@/pages/Login'

function AuthCallback() {
  // Supabase JS SDK picks up the OAuth tokens from the URL fragment automatically
  // via onAuthStateChange. Just redirect home.
  return <Navigate to="/" replace />
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/auth/callback" element={<AuthCallback />} />
          <Route path="/sessions" element={<Home />} />
          <Route path="/sessions/lobby/:id" element={<Lobby />} />
          <Route path="/sessions/game/:id" element={<Game />} />
          <Route path="*" element={<Navigate to="/sessions" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
