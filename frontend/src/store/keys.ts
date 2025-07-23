// store/keys.ts
import { create } from 'zustand';

type SharedSecretStore = {
  sharedSecrets: Record<string, Uint8Array>; // Maps userId -> shared secret
  setSharedSecret: (_: string, _unused: Uint8Array) => void;
  getSharedSecret: (_: string) => Uint8Array | undefined;
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
