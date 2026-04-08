import { create } from "zustand";
import { persist } from "zustand/middleware";
import { api } from "./api";

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: { id: string; plan: string; email?: string } | null;
  isAuthenticated: boolean;
  login: (accessToken: string, refreshToken: string) => void;
  logout: () => void;
  setUser: (user: AuthState["user"]) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      user: null,
      isAuthenticated: false,

      login: (accessToken, refreshToken) => {
        api.setToken(accessToken);
        api.setRefreshToken(refreshToken);
        set({ accessToken, refreshToken, isAuthenticated: true });

        // Wire up token refresh callback so api client can sync state
        api.onTokenRefresh = (newAccess, newRefresh) => {
          set({ accessToken: newAccess, refreshToken: newRefresh });
        };

        // Wire up logout handler for expired sessions
        api.setLogoutHandler(() => {
          useAuthStore.getState().logout();
        });

        // Fetch user info after login
        api.getMe().then((me) => {
          set({ user: { id: me.user_id, plan: me.plan } });
        }).catch(() => {
          // Non-fatal: user info will be fetched on next page load
        });
      },

      logout: () => {
        api.clearToken();
        api.onTokenRefresh = null;
        set({ accessToken: null, refreshToken: null, user: null, isAuthenticated: false });
      },

      setUser: (user) => set({ user }),
    }),
    {
      name: "tene-auth",
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state?.accessToken) {
          api.setToken(state.accessToken);
        }
        if (state?.refreshToken) {
          api.setRefreshToken(state.refreshToken);
        }
        if (state?.isAuthenticated) {
          // Re-wire callbacks after rehydration
          api.onTokenRefresh = (newAccess, newRefresh) => {
            useAuthStore.setState({ accessToken: newAccess, refreshToken: newRefresh });
          };
          api.setLogoutHandler(() => {
            useAuthStore.getState().logout();
          });
        }
      },
    }
  )
);
