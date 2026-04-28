import { create } from "zustand";

interface AuthSessionState {
  /** Memory-only auth has no persisted hydration step; kept for route gating compatibility. */
  _hasHydrated: boolean;
  accessToken: string | null;
  setHasHydrated: (value: boolean) => void;
  setAccessToken: (token: string | null) => void;
  clear: () => void;
}

export const useAuthSessionStore = create<AuthSessionState>()((set) => ({
  _hasHydrated: true,
  accessToken: null,
  setHasHydrated: (value: boolean) => set({ _hasHydrated: value }),
  setAccessToken: (token: string | null) => set({ accessToken: token, _hasHydrated: true }),
  clear: () => set({ accessToken: null, _hasHydrated: true }),
}));
