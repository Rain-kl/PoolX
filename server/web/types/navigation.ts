export type NavigationIconKey =
  | 'home'
  | 'import'
  | 'node'
  | 'workspace'
  | 'runtime'
  | 'log'
  | 'file'
  | 'user'
  | 'setting';

export interface NavigationItem {
  href: string;
  label: string;
  icon: NavigationIconKey;
  children?: NavigationItem[];
}
