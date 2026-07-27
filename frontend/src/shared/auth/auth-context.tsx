import { useCallback, useEffect, useState, type ReactNode } from 'react';

import {
  ApiError,
  apiRequest,
  decodeAdminDTO,
  decodeLoggedOut,
  decodeLoginResponseDTO,
  refreshAccessToken,
  setAccessToken,
  subscribeSessionInvalidated,
  type AdminDTO,
} from '@/shared/api/client';
import { AuthContext, type AuthStatus } from '@/shared/auth/auth-state';

export function AuthProvider({ children }: { children: ReactNode }) {
  const [admin, setAdmin] = useState<AdminDTO | null>(null);
  const [status, setStatus] = useState<AuthStatus>('restoring');

  const restoreSession = useCallback(async (): Promise<void> => {
    setStatus('restoring');
    const refreshResult = await refreshAccessToken();
    if (refreshResult === 'invalid') {
      setAdmin(null);
      setStatus('anonymous');
      return;
    }
    if (refreshResult === 'unavailable') {
      setStatus('unavailable');
      return;
    }

    try {
      const value = await apiRequest(
        '/api/v1/user/self',
        { retryAuth: false },
        decodeAdminDTO,
      );
      setAdmin(value);
      setStatus('authenticated');
    } catch (error) {
      setAccessToken(null);
      setAdmin(null);
      setStatus(
        error instanceof ApiError && error.status === 401
          ? 'anonymous'
          : 'unavailable',
      );
    }
  }, []);

  useEffect(() => {
    const unsubscribe = subscribeSessionInvalidated(() => {
      setAdmin(null);
      setStatus('anonymous');
    });

    const restoreTimer = window.setTimeout(() => {
      void restoreSession();
    }, 0);

    return () => {
      window.clearTimeout(restoreTimer);
      unsubscribe();
    };
  }, [restoreSession]);

  async function login(username: string, password: string): Promise<void> {
    const response = await apiRequest(
      '/api/v1/user/login',
      {
        method: 'POST',
        body: { username, password },
        authenticated: false,
        retryAuth: false,
      },
      decodeLoginResponseDTO,
    );
    setAccessToken(response.tokens.access_token);
    setAdmin({
      id: response.id,
      username: response.username,
      is_admin: response.is_admin,
    });
    setStatus('authenticated');
  }

  async function logout(): Promise<void> {
    try {
      await apiRequest(
        '/api/v1/user/logout',
        {
          method: 'GET',
          authenticated: false,
          retryAuth: false,
        },
        decodeLoggedOut,
      );
    } finally {
      setAccessToken(null);
      setAdmin(null);
      setStatus('anonymous');
    }
  }

  async function changePassword(
    currentPassword: string,
    newPassword: string,
  ): Promise<void> {
    await apiRequest(
      '/api/v1/user/change-password',
      {
        method: 'POST',
        body: { old_password: currentPassword, new_password: newPassword },
      },
      (value) => value,
    );
  }

  return (
    <AuthContext.Provider
      value={{
        admin,
        status,
        retryRestore: restoreSession,
        login,
        logout,
        changePassword,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
