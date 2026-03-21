import { apiRequest } from '@/lib/api/client';

import type {
  ManageUserAction,
  ManageUserResult,
  UserItem,
  UserMutationPayload,
} from '@/features/users/types';

export function getUsers(page: number) {
  return apiRequest<UserItem[]>(`/user/?p=${page}`);
}

export function searchUsers(keyword: string) {
  return apiRequest<UserItem[]>(`/user/search?keyword=${encodeURIComponent(keyword)}`);
}

export function getUser(id: number) {
  return apiRequest<UserItem>(`/user/${id}`);
}

export function createUser(payload: UserMutationPayload) {
  return apiRequest<void>('/user/', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateUser(payload: UserMutationPayload & { id: number }) {
  return apiRequest<void>('/user/update', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function manageUser(username: string, action: ManageUserAction) {
  return apiRequest<ManageUserResult>('/user/manage', {
    method: 'POST',
    body: JSON.stringify({ username, action }),
  });
}
