import { create } from 'zustand';

interface AppShellState {
  isSidebarCollapsed: boolean;
  isMobileSidebarOpen: boolean;
  toggleSidebar: () => void;
  setSidebarCollapsed: (value: boolean) => void;
  setMobileSidebarOpen: (value: boolean) => void;
}

export const useAppShellStore = create<AppShellState>((set) => ({
  isSidebarCollapsed: false,
  isMobileSidebarOpen: false,
  toggleSidebar: () =>
    set((state) => ({
      isSidebarCollapsed: !state.isSidebarCollapsed,
    })),
  setSidebarCollapsed: (value) => set({ isSidebarCollapsed: value }),
  setMobileSidebarOpen: (value) => set({ isMobileSidebarOpen: value }),
}));
