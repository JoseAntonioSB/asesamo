/**
 * ASESAMO — Global Theme Context & Theme Manager
 * Location: asesamo-web/contexts/theme-context.tsx
 */

'use client';

import React, { createContext, useContext, useEffect, useState, useCallback, useMemo } from 'react';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { setStoredUser, getStoredUser } from '@/lib/auth';
import { toast } from 'sonner';

export type Theme = 'light' | 'dark';

interface ThemeContextType {
  theme: Theme;
  toggleTheme: () => Promise<void>;
  setTheme: (theme: Theme) => Promise<void>;
  isAdminUser: boolean;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

const ADMIN_ROLES = new Set(['plant_admin', 'super_admin', 'ehs_manager', 'contractor_admin']);

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();

  const isAdminUser = useMemo(() => {
    return !!user?.role && ADMIN_ROLES.has(user.role as string);
  }, [user?.role]);

  // Determine initial theme:
  // 1. If user has explicit preference ('light' | 'dark'), use it.
  // 2. Otherwise: if administrative user -> default to 'light'.
  // 3. Otherwise (guard/worker/unauthenticated) -> default to 'dark'.
  const resolveDefaultTheme = useCallback((): Theme => {
    if (typeof window === 'undefined') return 'dark';

    const localOverride = localStorage.getItem('asesamo_theme') as Theme | null;
    if (localOverride === 'light' || localOverride === 'dark') {
      return localOverride;
    }

    if (user?.theme_preference === 'light' || user?.theme_preference === 'dark') {
      return user.theme_preference;
    }

    if (isAdminUser) {
      return 'light';
    }

    return 'dark';
  }, [user, isAdminUser]);

  const [theme, setThemeState] = useState<Theme>('dark');
  const [isMounted, setIsMounted] = useState(false);

  // Apply class to document element
  const applyThemeClass = (t: Theme) => {
    if (typeof document === 'undefined') return;
    const root = document.documentElement;
    if (t === 'dark') {
      root.classList.add('dark');
      root.classList.remove('light');
    } else {
      root.classList.remove('dark');
      root.classList.add('light');
    }
  };

  // Sync on mount and when user identity changes
  useEffect(() => {
    const initial = resolveDefaultTheme();
    setThemeState(initial);
    applyThemeClass(initial);
    setIsMounted(true);
  }, [resolveDefaultTheme]);

  const setTheme = useCallback(
    async (newTheme: Theme, showToast = true) => {
      setThemeState(newTheme);
      applyThemeClass(newTheme);

      if (typeof window !== 'undefined') {
        localStorage.setItem('asesamo_theme', newTheme);
      }

      // Update cached user profile
      const stored = getStoredUser();
      if (stored) {
        stored.theme_preference = newTheme;
        setStoredUser(stored);
      }

      // If user is authenticated, persist preference in PostgreSQL
      if (user) {
        try {
          await apiClient.updateTheme(newTheme);
        } catch {
          // Non-blocking background sync
        }
      }

      if (showToast) {
        toast.info(newTheme === 'light' ? 'Modo claro activado' : 'Modo oscuro activado', {
          duration: 1800,
        });
      }
    },
    [user]
  );

  const toggleTheme = useCallback(async () => {
    const next: Theme = theme === 'dark' ? 'light' : 'dark';
    await setTheme(next, true);
  }, [theme, setTheme]);

  return (
    <ThemeContext.Provider
      value={{
        theme: isMounted ? theme : 'dark',
        toggleTheme,
        setTheme,
        isAdminUser,
      }}
    >
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme(): ThemeContextType {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
