// store/userKeys.ts
import { create } from 'zustand';
import { get, del } from 'idb-keyval';
import { set as setIndexedDB } from 'idb-keyval';

interface PrivateKeyStore {
  ik?: Uint8Array;
  spk?: Uint8Array;
  opks: Uint8Array[];
  setKeys: (keys: { ik: Uint8Array; spk: Uint8Array; opks: Uint8Array[] }, persist?: boolean) => Promise<void>;
  clearKeys: () => Promise<void>;
  loadPersistedKeys: () => Promise<void>;
}

export const useUserKeys = create<PrivateKeyStore>((set) => ({
  ik: undefined,
  spk: undefined,
  opks: [],

  setKeys: async ({ ik, spk, opks }, persist = false) => {
    set({ ik, spk, opks });

    if (persist) {
      await setIndexedDB('userKeys', {
        ik: Array.from(ik),
        spk: Array.from(spk),
        opks: opks.map(opk => Array.from(opk)),
      });
    }
  },

  clearKeys: async () => {
    set({ ik: undefined, spk: undefined, opks: [] });
    await del('userKeys');
  },

  loadPersistedKeys: async () => {
    const stored = await get<{
      ik: number[];
      spk: number[];
      opks: number[][];
    }>('userKeys');

    if (stored) {
      set({
        ik: new Uint8Array(stored.ik),
        spk: new Uint8Array(stored.spk),
        opks: stored.opks.map((arr) => new Uint8Array(arr)),
      });
    }
  }
}));
