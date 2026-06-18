import { create } from 'zustand';

export interface Product {
  id: string;
  title: string;
  description?: string;
  category?: string;
  status: string;
  updatedAt: string;
}

interface PIMState {
  // UI State only
  searchQuery: string;
  setSearchQuery: (query: string) => void;
}

export const usePIMStore = create<PIMState>((set) => ({
  searchQuery: "",
  setSearchQuery: (query: string) => set({ searchQuery: query }),
}));
