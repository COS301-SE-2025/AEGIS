import { create } from 'zustand';

interface SharedKeyStore {
  keys: Record<string, Uint8Array>; // key: userId, value: derived shared key
  setSharedKey: (_: string, _unused: Uint8Array) => void;
  getSharedKey: (_: string) => Uint8Array | undefined;
  clearAllKeys: () => void;
}

export const useDerivedKeys = create<SharedKeyStore>((set, get) => ({
  keys: {},
  setSharedKey: (userId, key) => {
    set((state) => ({
      keys: {
        ...state.keys,
        [userId]: key,
      },
    }));
  },
  getSharedKey: (userId) => {
    return get().keys[userId];
  },
  clearAllKeys: () => set({ keys: {} }),
}));
