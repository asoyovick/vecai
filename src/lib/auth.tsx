import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react"
import type { ReactNode } from "react"
import { api } from "./api"
import type { User } from "./api"

interface AuthState {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (name: string, email: string, password: string, jobTitle: string) => Promise<void>
  logout: () => Promise<void>
  refresh: () => Promise<void>
  setUser: (user: User) => void
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUserState] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  const refresh = useCallback(async () => {
    try {
      const data = await api.get<{ user: User }>("/api/auth/me")
      setUserState(data.user)
    } catch {
      setUserState(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  const login = useCallback(async (email: string, password: string) => {
    const data = await api.post<{ user: User }>("/api/auth/login", {
      email,
      password,
    })
    setUserState(data.user)
  }, [])

  const register = useCallback(
    async (name: string, email: string, password: string, jobTitle: string) => {
      const data = await api.post<{ user: User }>("/api/auth/register", {
        name,
        email,
        password,
        jobTitle,
      })
      setUserState(data.user)
    },
    [],
  )

  const logout = useCallback(async () => {
    try {
      await api.post("/api/auth/logout")
    } finally {
      setUserState(null)
    }
  }, [])

  const setUser = useCallback((u: User) => setUserState(u), [])

  const value = useMemo(
    () => ({ user, loading, login, register, logout, refresh, setUser }),
    [user, loading, login, register, logout, refresh, setUser],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used inside AuthProvider")
  return ctx
}
