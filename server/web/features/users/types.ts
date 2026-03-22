export interface UserItem {
  id: number;
  username: string;
  display_name: string;
  role: number;
  status: number;
  email?: string;
  github_id?: string;
  wechat_id?: string;
}

export interface UserMutationPayload {
  id?: number;
  username: string;
  display_name: string;
  password: string;
}

export type ManageUserAction = 'promote' | 'demote' | 'delete' | 'disable' | 'enable';

export interface ManageUserResult {
  role: number;
  status: number;
}
