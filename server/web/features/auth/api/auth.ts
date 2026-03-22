import { apiRequest } from '@/lib/api/client';
import type {
  AuthUser,
  LoginPayload,
  PasswordResetRequestPayload,
  RegisterPayload,
} from '@/types/auth';

export function getCurrentUser() {
  return apiRequest<AuthUser>('/user/self');
}

export function login(payload: LoginPayload) {
  return apiRequest<AuthUser>('/user/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function logout() {
  return apiRequest<void>('/user/logout');
}

export function register(payload: RegisterPayload, turnstileToken?: string) {
  const query = turnstileToken ? `?turnstile=${encodeURIComponent(turnstileToken)}` : '';

  return apiRequest<void>(`/user/register${query}`, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function sendEmailVerification(email: string, turnstileToken?: string) {
  const searchParams = new URLSearchParams({ email });
  if (turnstileToken) {
    searchParams.set('turnstile', turnstileToken);
  }

  return apiRequest<void>(`/verification?${searchParams.toString()}`);
}

export function sendPasswordResetEmail(email: string, turnstileToken?: string) {
  const searchParams = new URLSearchParams({ email });
  if (turnstileToken) {
    searchParams.set('turnstile', turnstileToken);
  }

  return apiRequest<void>(`/reset_password?${searchParams.toString()}`);
}

export function resetPassword(payload: PasswordResetRequestPayload) {
  return apiRequest<string>('/user/reset', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function exchangeGitHubCode(code: string) {
  return apiRequest<AuthUser>(`/oauth/github?code=${encodeURIComponent(code)}`);
}
