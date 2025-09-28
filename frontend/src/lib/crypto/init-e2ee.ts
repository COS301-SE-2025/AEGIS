// lib/crypto/init-e2ee.ts
import sodium from "libsodium-wrappers";
import { useUserKeys } from "../../store/userKeys";
import { uploadKeyBundle, getAvailableOPKCount, refillOPKs } from "../../services/x3dhService";
import type { OneTimePreKey, RegisterBundleRequest } from "../../lib/crypto/x3dh";

/** tiny guards so we don't accept empty or truncated keys */
const isU8 = (v: unknown, len?: number) =>
  v instanceof Uint8Array && (len ? v.length === len : v.length > 0);

const ED25519_PUB = 32;
const ED25519_SK  = 64; // ed25519 secret key (libsodium format)
const X25519_PUB  = 32;
const X25519_SK   = 32;

type InitOpts = {
  tenantId?: string;
  userId: string | number;
  userEmail?: string;
  minOPKs?: number;     // maintain at least this many on server
  targetOPKs?: number;  // when refilling, upload up to this many
};

/** Generate + upload OPKs if server pool is low. Also persists local private halves. */
export async function ensureOPKs(userId: string, minOPKs = 10, targetOPKs = 40) {
  await sodium.ready;

  let serverCount = 0;
  try {
    serverCount = await getAvailableOPKCount(userId);
  } catch {
    // server unavailable — skip quietly
    return;
  }
  if (serverCount >= minOPKs) return;

  // Generate enough to reach targetOPKs on server
  const toGen = Math.max(0, targetOPKs - serverCount);
  if (!toGen) return;

  const localOPKs: { id: string; pub: Uint8Array; priv: Uint8Array; used?: boolean }[] = [];
  const uploadOPKs: OneTimePreKey[] = [];

  for (let i = 0; i < toGen; i++) {
    const { publicKey, privateKey } = sodium.crypto_box_keypair();
    const id = crypto.randomUUID();
    localOPKs.push({ id, pub: publicKey, priv: privateKey, used: false });
    uploadOPKs.push({ keyId: id, publicKey: sodium.to_base64(publicKey) });
  }

  // Persist privates locally; upload publics to server
  await useUserKeys.getState().upsertOPKs(localOPKs, true);
  await refillOPKs(userId, uploadOPKs);
}

/**
 * Ensures all local keys exist and are valid,
 * writes them to IndexedDB, registers server bundle if missing,
 * and tops up the server OPK pool.
 */
export async function initializeE2EE(opts: InitOpts) {
  const { tenantId, userId, minOPKs = 10, targetOPKs = 40 } = opts;
  if (!userId) throw new Error("initializeE2EE: missing userId");

  await sodium.ready;

  // 1) Namespace + load persisted
  if (tenantId) useUserKeys.getState().setNamespace(tenantId, String(userId));
  await useUserKeys.getState().loadPersistedKeys();

  let { ikPub, ikPriv, spkPub, spkPriv, opks } = useUserKeys.getState();

  // 2) Generate if missing
  const needIK  = !isU8(ikPriv, ED25519_SK) || !isU8(ikPub, ED25519_PUB);
  const needSPK = !isU8(spkPriv, X25519_SK) || !isU8(spkPub, X25519_PUB);

  if (needIK) {
    const kp = sodium.crypto_sign_keypair(); // ed25519
    ikPub = kp.publicKey;
    ikPriv = kp.privateKey;
    await useUserKeys.getState().setKeys({ ikPub, ikPriv }, true);
  }
  if (needSPK) {
    const kp = sodium.crypto_box_keypair(); // x25519
    spkPub = kp.publicKey;
    spkPriv = kp.privateKey;
    await useUserKeys.getState().setKeys({ spkPub, spkPriv }, true);
  }

  if (!isU8(ikPriv, ED25519_SK) || !isU8(ikPub, ED25519_PUB)) {
    throw new Error("initializeE2EE: identity keypair not available after generation");
  }
  if (!isU8(spkPriv, X25519_SK) || !isU8(spkPub, X25519_PUB)) {
    throw new Error("initializeE2EE: signed-prekey keypair not available after generation");
  }

  // 3) Local OPK pool (keep a small buffer locally)
  if (!Array.isArray(opks)) opks = [];
  const localMissing = Math.max(0, 10 - opks.length);
  if (localMissing > 0) {
    const newLocal = Array.from({ length: localMissing }).map(() => {
      const kp = sodium.crypto_box_keypair();
      return {
        id: crypto.randomUUID(),
        pub: kp.publicKey,
        priv: kp.privateKey,
        used: false,
      };
    });
    await useUserKeys.getState().upsertOPKs(newLocal, true);
    opks = useUserKeys.getState().opks;
  }

  // 4) Register bundle server-side if missing
  const spkSignature = sodium.crypto_sign_detached(spkPub!, ikPriv!); // sig over SPK
  let mustRegisterBundle = false;
  try {
    const r = await fetch(`http://localhost:8080/api/v1/x3dh/bundle/${userId}`, {
      headers: { Authorization: `Bearer ${sessionStorage.getItem("authToken") || ""}` },
    });
    mustRegisterBundle = r.status === 404;
  } catch {
    mustRegisterBundle = true; // network hiccup — server is idempotent anyway
  }

  if (mustRegisterBundle) {
    // upload an initial slice of OPKs so server pool is non-empty
    const initialUpload = Math.min(opks.length, 20);
    const oneTimePreKeys: OneTimePreKey[] = opks.slice(0, initialUpload).map(o => ({
      keyId: o.id,
      publicKey: sodium.to_base64(o.pub),
    }));

    const bundle: RegisterBundleRequest = {
      userId: String(userId),
      identityKey: sodium.to_base64(ikPub!),
      signedPreKey: sodium.to_base64(spkPub!),
      spkSignature: sodium.to_base64(spkSignature),
      oneTimePreKeys,
    };

    await uploadKeyBundle(bundle);
  }

  // 5) ✅ Always top-up server OPK pool after init
  try {
    await ensureOPKs(String(userId), minOPKs, targetOPKs);
  } catch {
    // non-fatal; app can still run locally
  }

  // 6) Return publics for convenience
  return {
    ikPub: useUserKeys.getState().getPublics().ikPub!,
    spkPub: useUserKeys.getState().getPublics().spkPub!,
  };
}
