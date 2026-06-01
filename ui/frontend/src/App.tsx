import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { AuthProvider } from '@/contexts/AuthContext'
import { AuthGuard } from '@/components/AuthGuard'
import { Home } from '@/pages/sessions/Home'
import { Lobby } from '@/pages/sessions/Lobby'
import { Game } from '@/pages/sessions/Game'
import { Login } from '@/pages/Login'
import { DailyHome } from '@/pages/daily/Home'
import { DailyGroup } from '@/pages/daily/Group'
import { supabase } from '@/lib/supabase'

// Sits on /auth/callback while the SDK finishes the OAuth flow, then navigates.
// Uses both getSession() (for implicit flow — tokens already stored before React mounts)
// and onAuthStateChange (for PKCE — async code exchange fires after mount).
function AuthCallback() {
  const navigate = useNavigate()

  useEffect(() => {
    let done = false
    const go = (path: string) => {
      if (!done) { done = true; navigate(path, { replace: true }) }
    }

    const { data: { subscription } } = supabase.auth.onAuthStateChange((_event, session) => {
      if (session) go('/daily')
    })

    supabase.auth.getSession().then(({ data: { session } }) => {
      if (session) go('/daily')
    })

    const timeout = setTimeout(() => go('/login'), 10_000)

    return () => { subscription.unsubscribe(); clearTimeout(timeout) }
  }, [navigate])

  return null
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/auth/callback" element={<AuthCallback />} />

          <Route
            path="/daily"
            element={
              <AuthGuard>
                <DailyHome />
              </AuthGuard>
            }
          />
          <Route
            path="/daily/groups/:groupId"
            element={
              <AuthGuard>
                <DailyGroup />
              </AuthGuard>
            }
          />

          <Route path="/sessions" element={<Home />} />
          <Route path="/sessions/lobby/:id" element={<Lobby />} />
          <Route path="/sessions/game/:id" element={<Game />} />

          <Route path="*" element={<Navigate to="/daily" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
