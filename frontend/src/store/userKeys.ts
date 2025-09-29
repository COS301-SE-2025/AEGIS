// store/userKeys.ts
import { create } from "zustand";
import { get as idbGet, set as idbSet, del as idbDel } from "idb-keyval";

/** On-disk shape (v2) for robust persistence */
type PersistV2 = {
  version: 2;
  ikPub?: number[];   // ed25519 public
  ikPriv?: number[];  // ed25519 private
  spkPub?: number[];  // x25519 public
  spkPriv?: number[]; // x25519 private
  opks: { id: string; pub: number[]; priv: number[]; used?: boolean }[];
};

/** Back-compat shape (your old v1) */
type PersistV1 = {
  ik?: number[];     // ambiguous (you stored one key array)
  spk?: number[];    // ambiguous
  opks: number[][];  // just key blobs
};

type OPK = { id: string; pub: Uint8Array; priv: Uint8Array; used?: boolean };

type UserKeysState = {
  // optional namespace for IDB keying
  tenantId?: string;
  userId?: string;

  // Identity key (ed25519)
  ikPub?: Uint8Array;
  ikPriv?: Uint8Array;

  // Signed PreKey (x25519)
  spkPub?: Uint8Array;
  spkPriv?: Uint8Array;

  // One-time PreKeys
  opks: OPK[];

  /** Configure namespacing for persistence key (recommended) */
  setNamespace: (tenantId: string, userId: string) => void;

  /** Set any subset of keys; set persist=true to write to IndexedDB */
  setKeys: (
    k: Partial<Pick<UserKeysState, "ikPub" | "ikPriv" | "spkPub" | "spkPriv" | "opks">>,
    persist?: boolean
  ) => Promise<void>;

  /** Replace/merge OPKs (e.g., after refill) */
  upsertOPKs: (opks: OPK[], persist?: boolean) => Promise<void>;

  /** Mark a specific OPK as used (by id) */
  markOPKUsed: (id: string, persist?: boolean) => Promise<void>;

  /** Clear all keys from memory and disk (best-effort zeroization) */
  clearKeys: () => Promise<void>;

  /** Load keys from disk into memory (handles v1→v2 migration) */
  loadPersistedKeys: () => Promise<void>;

  /** Export ONLY publics (defensive copies) */
  getPublics: () => {
    ikPub?: Uint8Array;
    spkPub?: Uint8Array;
    opkPubs: { id: string; pub: Uint8Array; used?: boolean }[];
  };
};

// ---------- Helpers ----------

const BASE_IDB_KEY = "userKeys:v2";
const makeIDBKey = (tenantId?: string, userId?: string) =>
  tenantId && userId ? `${BASE_IDB_KEY}:${tenantId}:${userId}` : BASE_IDB_KEY;

function toPersistV2(state: UserKeysState): PersistV2 {
  return {
    version: 2,
    ikPub: state.ikPub ? Array.from(state.ikPub) : undefined,
    ikPriv: state.ikPriv ? Array.from(state.ikPriv) : undefined,
    spkPub: state.spkPub ? Array.from(state.spkPub) : undefined,
    spkPriv: state.spkPriv ? Array.from(state.spkPriv) : undefined,
    opks: state.opks.map(o => ({
      id: o.id,
      pub: Array.from(o.pub),
      priv: Array.from(o.priv),
      used: o.used,
    })),
  };
}

function fromPersistV2(p: PersistV2): Partial<UserKeysState> {
  return {
    ikPub: p.ikPub ? new Uint8Array(p.ikPub) : undefined,
    ikPriv: p.ikPriv ? new Uint8Array(p.ikPriv) : undefined,
    spkPub: p.spkPub ? new Uint8Array(p.spkPub) : undefined,
    spkPriv: p.spkPriv ? new Uint8Array(p.spkPriv) : undefined,
  // ensure stable structure even if opks missing
    opks: (p.opks || []).map(o => ({
      id: o.id,
      pub: new Uint8Array(o.pub || []),
      priv: new Uint8Array(o.priv || []),
      used: o.used,
    })),
  };
}

/** tiny non-cryptographic hash for stable IDs in migration */
function shortHash(nums: number[]): string {
  let h = 0 >>> 0;
  for (let i = 0; i < nums.length; i++) {
    h = (h * 31 + (nums[i] & 0xff)) >>> 0;
  }
  return h.toString(16);
}

/** v1 → v2 best-effort migration:
 *  - Treat v1.ik and v1.spk as *private* keys (most likely how you used them).
 *  - OPKs had no IDs; synthesize stable IDs from a short hash of the blob.
 *  - Publics are unknown at this point; they can be filled on next generation/fetch.
 */
function migrateV1ToV2(v1: PersistV1): PersistV2 {
  const opks: PersistV2["opks"] = (v1.opks || []).map((blob) => ({
    id: `opk-${shortHash(blob)}`,
    pub: [], // unknown in v1
    priv: blob,
    used: false,
  }));
  return {
    version: 2,
    ikPriv: v1.ik,
    spkPriv: v1.spk,
    opks,
  };
}

// ---------- Store ----------

export const useUserKeys = create<UserKeysState>((set, get) => ({
  tenantId: undefined,
  userId: undefined,

  ikPub: undefined,
  ikPriv: undefined,
  spkPub: undefined,
  spkPriv: undefined,
  opks: [],

  setNamespace: (tenantId, userId) => {
    set({ tenantId, userId });
  },

  setKeys: async (k, persist = false) => {
    set(k);
    if (persist) {
      try {
        await idbSet(makeIDBKey(get().tenantId, get().userId), toPersistV2(get()));
      } catch (e) {
        // swallow storage errors to avoid crashing app; optionally log
        // console.warn("Persist setKeys failed:", e);
      }
    }
  },

  upsertOPKs: async (opks, persist = false) => {
    const current = get().opks.slice();
    const map = new Map<string, OPK>(current.map(o => [o.id, o]));
    for (const o of opks) map.set(o.id, o);
    const merged = Array.from(map.values());
    set({ opks: merged });
    if (persist) {
      try {
        await idbSet(makeIDBKey(get().tenantId, get().userId), toPersistV2(get()));
      } catch (e) {
        // console.warn("Persist upsertOPKs failed:", e);
      }
    }
  },

  markOPKUsed: async (id, persist = false) => {
    set({
      opks: get().opks.map(o => (o.id === id ? { ...o, used: true } : o)),
    });
    if (persist) {
      try {
        await idbSet(makeIDBKey(get().tenantId, get().userId), toPersistV2(get()));
      } catch (e) {
        // console.warn("Persist markOPKUsed failed:", e);
      }
    }
  },

  clearKeys: async () => {
    // best-effort zeroization before dropping references
    const s = get();
    try {
      s.ikPriv?.fill(0);
      s.spkPriv?.fill(0);
      s.opks.forEach(o => { o.priv.fill(0); });
    } catch {
      // ignore zeroization errors
    }

    set({
      ikPub: undefined,
      ikPriv: undefined,
      spkPub: undefined,
      spkPriv: undefined,
      opks: [],
    });

    // delete both namespaced and legacy keys
    try {
      await idbDel(makeIDBKey(get().tenantId, get().userId));
    } catch {
      // ignore
    }
    try {
      await idbDel("userKeys"); // legacy v1 key
    } catch {
      // ignore
    }
  },

  loadPersistedKeys: async () => {
    // Try namespaced v2 first
    try {
      const v2 = await idbGet<PersistV2>(makeIDBKey(get().tenantId, get().userId));
      if (v2 && v2.version === 2) {
        set(fromPersistV2(v2));
        return;
      }
    } catch {
      // ignore read failure
    }

    // Try non-namespaced v2 (fallback if namespace was added later)
    try {
      const v2global = await idbGet<PersistV2>(BASE_IDB_KEY);
      if (v2global && v2global.version === 2) {
        set(fromPersistV2(v2global));
        // migrate it into namespaced key if namespace is present
        if (get().tenantId && get().userId) {
          try {
            await idbSet(makeIDBKey(get().tenantId, get().userId), v2global);
            await idbDel(BASE_IDB_KEY);
          } catch {
            // ignore
          }
        }
        return;
      }
    } catch {
      // ignore
    }

    // Try legacy v1
    try {
      const v1 = await idbGet<PersistV1>("userKeys");
      if (v1) {
        const migrated = migrateV1ToV2(v1);
        set(fromPersistV2(migrated));
        // write back as v2 for future
        try {
          await idbSet(makeIDBKey(get().tenantId, get().userId), migrated);
          // optionally delete legacy
          await idbDel("userKeys");
        } catch {
          // ignore
        }
      }
    } catch {
      // ignore read failure
    }
  },

  getPublics: () => {
    const { ikPub, spkPub, opks } = get();
    return {
      ikPub: ikPub ? new Uint8Array(ikPub) : undefined,
      spkPub: spkPub ? new Uint8Array(spkPub) : undefined,
      opkPubs: opks.map(({ id, pub, used }) => ({ id, pub: new Uint8Array(pub), used })),
    };
  },
}));



