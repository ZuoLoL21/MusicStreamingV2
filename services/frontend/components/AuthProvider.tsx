'use client';

import { createContext, useContext, useEffect, useState } from 'react';
import Cookies from 'js-cookie';
import { useAuthStore } from '@/lib/store';

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  isLoading: true,
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isLoading, setIsLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const { setAuth, token } = useAuthStore();

  useEffect(() => {
    // Check for token in cookies on mount
    const cookieToken = Cookies.get('token');
    const cookieRefreshToken = Cookies.get('refresh_token');
    const cookieUserUuid = Cookies.get('user_uuid');

    // Validate cookies exist and are not the string "undefined"
    const isValid =
      cookieToken && cookieToken !== 'undefined' &&
      cookieRefreshToken && cookieRefreshToken !== 'undefined' &&
      cookieUserUuid && cookieUserUuid !== 'undefined';

    if (isValid) {
      // Sync cookies with Zustand store
      setAuth(cookieToken!, cookieRefreshToken!, cookieUserUuid!);
      setIsAuthenticated(true);
    } else {
      // Clear invalid cookies
      Cookies.remove('token');
      Cookies.remove('refresh_token');
      Cookies.remove('user_uuid');
      setIsAuthenticated(false);
    }

    setIsLoading(false);
  }, [setAuth]);

  // Also watch the Zustand store token for changes
  useEffect(() => {
    setIsAuthenticated(!!token);
  }, [token]);

  return (
    <AuthContext.Provider value={{ isAuthenticated, isLoading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
