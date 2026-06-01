import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto flex h-14 items-center justify-between px-4">
          <span className="font-bold text-lg">Loaded Questions</span>
          <nav className="flex gap-1">
            <NavLink
              to="/daily"
              className={({ isActive }) =>
                `px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                }`
              }
            >
              Daily
            </NavLink>
            <NavLink
              to="/sessions"
              className={({ isActive }) =>
                `px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                }`
              }
            >
              Quick Session
            </NavLink>
          </nav>
        </div>
      </header>
      <main className="container mx-auto px-4 py-8">{children}</main>
    </div>
  )
}
