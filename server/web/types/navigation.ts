export type NavigationIconKey =
  | 'home'
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
