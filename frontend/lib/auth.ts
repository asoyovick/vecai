import { apiGet, apiPost } from "./api";

/**
 * Auth session helpers for the vecai backend
 * (backend/internal/auth: register, login, /users/me).
 *
 * The backend issues a JWT valid for 24h. It is kept in localStorage and
 * attached as a Bearer token to authenticated requests. See "Security
 * considerations" at the bottom for why and what to change for production.
 */

export type Role =
  | "owner"
  | "architect"
  | "quantity_surveyor"
  | "engineer"
  | "contractor"
  | "supplier";

export type User = {
  id: string;
  email: string;
  fullname: string;
  role: Role;
  phone?: string;
  is_verified: boolean;
  created_at: string;
};

export type AuthResponse = { token: string; user: User };

const TOKEN_KEY = "vecai_token";
const USER_KEY = "vecai_user";

const ROLE_LABELS: Record<Role, string> = {
  owner: "Project Owner",
  architect: "Architect",
  quantity_surveyor: "Quantity Surveyor",
  engineer: "Engineer",
  contractor: "Contractor",
  supplier: "Supplier",
};

export const roleOptions: { value: Role; label: string }[] = Object.entries(
  ROLE_LABELS,
).map(([value, label]) => ({
  value: value as Role,
  label,
}));

export function roleLabel(role: Role): string {
  return ROLE_LABELS[role] ?? role;
}

// ---- Token / session storage -------------------------------------------

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(TOKEN_KEY);
}

export function getStoredUser(): User | null {
  if (typeof window === "undefined") return null;
  const raw = window.localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as User;
  } catch {
    return null;
  }
}

function storeSession(token: string, user: User) {
  window.localStorage.setItem(TOKEN_KEY, token);
  window.localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function clearSession() {
  window.localStorage.removeItem(TOKEN_KEY);
  window.localStorage.removeItem(USER_KEY);
}

// ---- API calls ----------------------------------------------------------

export async function register(input: {
  fullname: string;
  email: string;
  password: string;
  role: Role;
  phone?: string;
}): Promise<AuthResponse> {
  const res = await apiPost<AuthResponse>("/api/v1/auth/register", input);
  storeSession(res.token, res.user);
  return res;
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const res = await apiPost<AuthResponse>("/api/v1/auth/login", {
    email,
    password,
  });
  storeSession(res.token, res.user);
  return res;
}

/** Fetch the profile of the currently authenticated user. */
export async function fetchProfile(token: string): Promise<User> {
  return apiGet<User>("/api/v1/users/me", token);
}

/**
 * Validate the stored session against the backend. Returns the fresh user
 * when the token is still valid; clears storage and returns null otherwise.
 */
export async function restoreSession(): Promise<User | null> {
  const token = getToken();
  if (!token) return null;
  try {
    const user = await fetchProfile(token);
    window.localStorage.setItem(USER_KEY, JSON.stringify(user));
    return user;
  } catch {
    clearSession();
    return null;
  }
}

// ---- Security considerations --------------------------------------------
//
// localStorage keeps the token available to any XSS payload. For production,
// prefer an HttpOnly cookie set by the backend (requires a backend change or
// a BFF layer) and short-lived refresh tokens. For this MVP the 24h JWT and
// same-site API make the tradeoff acceptable.
