// src/lib/crypto/groupSecret.ts
import sodium from "libsodium-wrappers";
import { hkdf } from "./hkdf"; // hkdf(ikm, salt, info, length) => Promise<Uint8Array>

/* =========================
   Types (adjust or reuse yours)
   ========================= */
export type ChatMember = {
  id: string;
  email?: string;
  name?: string;
  /** X25519 public key bytes (if you preload them; optional here) */
  publicKey?: Uint8Array;
};

export type Chat = {
  id: string;
  name?: string;
  members: ChatMember[];
};

export type GroupSecretResult = {
  /** 32 bytes derived key (TESTING ONLY if sourced from public keys) */
  sharedSecret: Uint8Array;
  /** Base64url-encoded ephemeral X25519 public key (demo) */
  ephemeralPubKey?: string;
  /** Placeholder for OPK tracking */
  opkIdUsed?: string;
};

/* =========================
   Internals / Helpers
   ========================= */

const te = new TextEncoder();

/** Return a fresh Uint8Array (no SharedArrayBuffer, no aliasing). */
function freshU8(view: Uint8Array | ArrayBuffer): Uint8Array {
  if (view instanceof Uint8Array) {
    const out = new Uint8Array(view.length);
    out.set(view);
    return out;
  }
  return new Uint8Array(view.slice(0));
}

/** Try multiple common bundle field names for an X25519 identity/public key (base64 string). */
function pickBundlePubKeyField(bundle: any): string | undefined {
  return (
    bundle?.ikX25519 ??
    bundle?.identityKeyX25519 ??
    bundle?.identity_x25519 ??
    bundle?.x25519?.identity ??
    bundle?.identity_key ??
    bundle?.public_key ??
    bundle?.identityKey
  );
}

/** Base64/base64url (and last-resort hex) decode with sanitation and variant tries. */
async function decodePubKeyFlexible(b64OrHex: string): Promise<Uint8Array> {
  await sodium.ready;

  const clean = String(b64OrHex).trim().replace(/\s+/g, "");
  const variants = [
    sodium.base64_variants.URLSAFE_NO_PADDING,
    sodium.base64_variants.URLSAFE,
    sodium.base64_variants.ORIGINAL_NO_PADDING,
    sodium.base64_variants.ORIGINAL,
  ];

  for (const v of variants) {
    try {
      return sodium.from_base64(clean, v);
    } catch {
      /* try next */
    }
  }

  // Last resort: hex
  const hex = clean.toLowerCase();
  if (/^[0-9a-f]+$/i.test(hex) && hex.length % 2 === 0) {
    const out = new Uint8Array(hex.length / 2);
    for (let i = 0; i < out.length; i++) {
      out[i] = parseInt(hex.slice(i * 2, i * 2 + 2), 16);
    }
    return out;
  }

  throw new Error("Invalid public key encoding (not base64/base64url/hex)");
}

/**
 * Pull a 32-byte X25519 public key from the bundle.
 * If you know the bundle returns Ed25519 here, enable conversion below.
 */
async function extractX25519FromBundle(bundle: any, tryEd25519Convert = false): Promise<Uint8Array> {
  await sodium.ready;

  const field = pickBundlePubKeyField(bundle);
  if (!field || typeof field !== "string") {
    const keys = bundle ? Object.keys(bundle) : [];
    throw new Error(`No base64 X25519 public key field found in bundle. Fields: [${keys.join(", ")}]`);
  }

  let pub = await decodePubKeyFlexible(field);

  // Expect 32 bytes. If not, it’s not a raw X25519 pub key.
  if (pub.length !== 32) {
    throw new Error(`Unexpected public key length: ${pub.length} (expected 32 for X25519)`);
  }

  if (tryEd25519Convert) {
    // Only enable if your backend supplies Ed25519 instead of X25519 here.
    try {
      pub = sodium.crypto_sign_ed25519_pk_to_curve25519(pub);
    } catch {
      /* if not actually Ed25519, keep as-is */
    }
  }

  return freshU8(pub);
}

/**
 * ⚠️ SECURITY NOTE:
 * This derives a key from concatenated *public* keys + HKDF, which is deterministic and NOT secret.
 * Use ONLY for testing. In production, derive a *secret* via ECDH (X25519):
 *   secret_i = X25519(selfPrivate, peerPublic)
 *   IKM = concat(secret_1, secret_2, ...)
 *   sharedSecret = HKDF(IKM, salt, info, 32)
 */
export async function deriveSharedSecret(
  groupId: string,
  publicKeys: Uint8Array[]
): Promise<Uint8Array> {
  if (!Array.isArray(publicKeys) || publicKeys.length === 0) {
    throw new Error("No public keys provided for secret derivation.");
  }
  if (!publicKeys.every((k) => k instanceof Uint8Array)) {
    throw new Error("All public keys must be Uint8Array instances.");
  }
  if (!publicKeys.every((k) => k.length === publicKeys[0].length)) {
    throw new Error("All public keys must have the same length.");
  }

  // IKM = concatenation of all public keys (TESTING ONLY)
  const keyLen = publicKeys[0].length;
  const ikm = new Uint8Array(publicKeys.length * keyLen);
  publicKeys.forEach((k, i) => ikm.set(k, i * keyLen));

  // Salt = H(groupId) for fixed-length and good distribution
  const saltBuf = await crypto.subtle.digest("SHA-256", te.encode(groupId));
  const salt = new Uint8Array(saltBuf);

  // Context for domain separation
  const info = te.encode("shared-secret-info");

  return hkdf(ikm, salt, info, 32);
}

/** Demo helper: produce a base64url X25519 ephemeral *public* key. */
function makeEphemeralPubKeyB64Url(): string {
  const { publicKey } = sodium.crypto_box_keypair();
  return sodium.to_base64(publicKey, sodium.base64_variants.URLSAFE_NO_PADDING);
}

/* =========================
   Public API
   ========================= */

/**
 * Generates a (testing) "group shared secret" using the current user's bundle.
 * In production: aggregate ECDH secrets for all members, then HKDF.
 */
export async function generateGroupSharedSecretForChat(
  chat: Chat,
  token: string,
  currentUserId: string
): Promise<GroupSecretResult | null> {
  try {
    await sodium.ready;

    // Fetch current user's bundle (adjust endpoint as needed)
    const resp = await fetch(
      `http://localhost:8080/api/v1/x3dh/bundle/${encodeURIComponent(currentUserId)}`,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    if (!resp.ok) {
      throw new Error(`Failed to fetch bundle for current user: ${resp.status}`);
    }

    const bundle = await resp.json();

    // Extract a valid 32-byte X25519 pub key
    const currentUserPub = await extractX25519FromBundle(bundle /*, tryEd25519Convert = true */);

    // TESTING: only current user's key. (Others won't be able to decrypt.)
    const publicKeys: Uint8Array[] = [currentUserPub];

    // Derive (testing) HKDF-based "shared secret"
    const sharedSecret = await deriveSharedSecret(chat.id, publicKeys);

    // Demo ephemeral public key and fake OPK id
    const ephemeralPubKey = makeEphemeralPubKeyB64Url();
    const opkIdUsed = `temp_opk_id_${Date.now()}`;

    return { sharedSecret, ephemeralPubKey, opkIdUsed };
  } catch (err) {
    console.error("Failed to generate group shared secret:", err);
    return null;
  }
}

