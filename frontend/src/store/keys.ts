// store/keys.ts
import { create } from 'zustand';

type SharedSecretStore = {
  sharedSecrets: Record<string, Uint8Array>; // Maps userId -> shared secret
  setSharedSecret: (userId: string, secret: Uint8Array) => void;
  getSharedSecret: (userId: string) => Uint8Array | undefined;
  clearAllSecrets: () => void;
};

export const useSharedSecrets = create<SharedSecretStore>((set, get) => ({
  sharedSecrets: {},

  setSharedSecret: (userId, secret) => {
    set((state) => ({
      sharedSecrets: {
        ...state.sharedSecrets,
        [userId]: secret,
      },
    }));
  },

  getSharedSecret: (userId) => {
    return get().sharedSecrets[userId];
  },

  clearAllSecrets: () => {
    set({ sharedSecrets: {} });
  },
}));
