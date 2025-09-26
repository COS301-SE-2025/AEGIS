// src/lib/crypto/aesgcm.ts

// -----------------------------
// Text encoders/decoders
// -----------------------------
const te = new TextEncoder();
const td = new TextDecoder();

// -----------------------------
// Small utilities
// -----------------------------

/** Return a fresh ArrayBuffer clone of a Uint8Array (avoids SAB/offset pitfalls). */
export function toArrayBuffer(u8: Uint8Array): ArrayBuffer {
  // If already full-span view over buffer, we can return a slice to ensure a fresh AB.
  if (u8.byteOffset === 0 && u8.byteLength === u8.buffer.byteLength) {
    // Slice returns a new ArrayBuffer
    return ArrayBuffer.prototype.slice.call(u8.buffer, 0) as ArrayBuffer;
  }
  const ab = new ArrayBuffer(u8.byteLength);
  new Uint8Array(ab).set(u8);
  return ab;
}

/** Optional Uint8Array -> BufferSource (with fresh ArrayBuffer) */
export function toBufferSource(u8?: Uint8Array): BufferSource | undefined {
  return u8 ? toArrayBuffer(u8) : undefined;
}

// -----------------------------
// Base64 / Base64url (chunk-safe)
// -----------------------------

/** Uint8Array -> standard base64 (chunk-safe) */
export function u8ToStdB64(u8: Uint8Array): string {
  // Build the binary string in manageable chunks to avoid "Maximum call stack" or huge args.
  let s = "";
  const CHUNK = 0x8000; // 32k
  for (let i = 0; i < u8.length; i += CHUNK) {
    const chunk = u8.subarray(i, i + CHUNK);
    // Use fromCharCode on a small array (converted to number[]) to keep arg count bounded.
    s += String.fromCharCode.apply(null, Array.from(chunk) as number[]);
  }
  return btoa(s);
}

/** standard base64 -> Uint8Array */
export function stdB64ToU8(b64: string): Uint8Array {
  const bin = atob(b64);
  const out = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
  return out;
}

/** Uint8Array -> base64url (no padding) */
export function u8ToB64url(u8: Uint8Array): string {
  const b64 = u8ToStdB64(u8);
  return b64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

/** base64url (no/any padding) -> Uint8Array */
export function b64urlToU8(s: string): Uint8Array {
  // Add required padding to multiple of 4
  const padLen = (4 - (s.length % 4)) % 4;
  const b64 = s.replace(/-/g, "+").replace(/_/g, "/") + "=".repeat(padLen);
  return stdB64ToU8(b64);
}

/** base64url -> standard base64 (string transform only; no full decode) */
export function b64urlToStd(b64u: string): string {
  let b64 = b64u.replace(/-/g, "+").replace(/_/g, "/");
  const pad = b64.length % 4;
  if (pad) b64 += "=".repeat(4 - pad);
  return b64;
}

/** standard base64 -> base64url (string transform only) */
export function stdToB64url(b64: string): string {
  return b64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

/** LEGACY NAMES (kept for backward compatibility). Internally use chunk-safe versions. */
export function b64url(bytes: Uint8Array): string {
  return u8ToB64url(bytes);
}
export function unb64url(s: string): Uint8Array {
  return b64urlToU8(s);
}

// -----------------------------
// AES-GCM key import
// -----------------------------

export async function importAesGcmKey(raw: Uint8Array): Promise<CryptoKey> {
  if (!(raw instanceof Uint8Array) || raw.byteLength !== 32) {
    throw new Error("AES-GCM key must be 32 bytes");
  }
  // SubtleCrypto accepts BufferSource; we pass a fresh AB to avoid SAB issues.
  return crypto.subtle.importKey("raw", toArrayBuffer(raw), "AES-GCM", false, ["encrypt", "decrypt"]);
}

// -----------------------------
// AES-GCM (strings)
// -----------------------------

/**
 * Encrypt a UTF-8 string with AES-GCM 256.
 * Returns { nonce: base64url, ciphertext: base64url }.
 */
export async function encryptMessage(
  keyBytes: Uint8Array,
  message: string,
  aad?: Uint8Array
): Promise<{ nonce: string; ciphertext: string }> {
  const key = await importAesGcmKey(keyBytes);
  const iv = crypto.getRandomValues(new Uint8Array(12));

  // Only include additionalData if provided (Chromium is strict about undefined)
  const params: AesGcmParams = { name: "AES-GCM", iv }; // Uint8Array is a BufferSource
  if (aad && aad.byteLength) {
    (params as any).additionalData = toArrayBuffer(aad);
  }

  const pt = te.encode(message);
  const encBuf = await crypto.subtle.encrypt(params, key, toArrayBuffer(pt));
  const encU8 = new Uint8Array(encBuf);

  return { nonce: u8ToB64url(iv), ciphertext: u8ToB64url(encU8) };
}

/**
 * Decrypt a UTF-8 string with AES-GCM 256.
 * Expects { nonce: base64url, ciphertext: base64url }.
 */
export async function decryptMessage(
  keyBytes: Uint8Array,
  nonceB64u: string,
  ciphertextB64u: string,
  aad?: Uint8Array
): Promise<string> {
  const key = await importAesGcmKey(keyBytes);
  const ivU8 = b64urlToU8(nonceB64u);
  const ctU8 = b64urlToU8(ciphertextB64u);

  const params: AesGcmParams = { name: "AES-GCM", iv: new Uint8Array(toArrayBuffer(ivU8)) };
  if (aad && aad.byteLength) {
    (params as any).additionalData = toArrayBuffer(aad);
  }

  const decBuf = await crypto.subtle.decrypt(params, key, toArrayBuffer(ctU8));
  return td.decode(decBuf);
}

// -----------------------------
// AES-GCM (bytes)
// -----------------------------

/**
 * Encrypt raw bytes with AES-GCM 256.
 * Returns { nonce: base64url, ciphertext: base64url }.
 */
export async function encryptBytes(
  keyBytes: Uint8Array,
  bytes: Uint8Array,
  aad?: Uint8Array
): Promise<{ nonce: string; ciphertext: string }> {
  const key = await importAesGcmKey(keyBytes);
  const iv = crypto.getRandomValues(new Uint8Array(12));

  const params: AesGcmParams = { name: "AES-GCM", iv: toArrayBuffer(iv) };
  if (aad && aad.byteLength) (params as any).additionalData = toArrayBuffer(aad);

  const encBuf = await crypto.subtle.encrypt(params, key, toArrayBuffer(bytes));
  const encU8 = new Uint8Array(encBuf);
  return { nonce: u8ToB64url(iv), ciphertext: u8ToB64url(encU8) };
}

/**
 * Decrypt raw bytes with AES-GCM 256.
 * Expects { nonce: base64url, ciphertext: base64url }.
 */
export async function decryptBytes(
  keyBytes: Uint8Array,
  nonceB64u: string,
  ciphertextB64u: string,
  aad?: Uint8Array
): Promise<Uint8Array> {
  const key = await importAesGcmKey(keyBytes);
  const iv = b64urlToU8(nonceB64u);
  const ct = b64urlToU8(ciphertextB64u);

  const params: AesGcmParams = { name: "AES-GCM", iv: toArrayBuffer(iv) };
  if (aad && aad.byteLength) (params as any).additionalData = toArrayBuffer(aad);

  const decBuf = await crypto.subtle.decrypt(params, key, toArrayBuffer(ct));
  return new Uint8Array(decBuf);
}

// Add a light type for what you need:
type AttachmentMsg = {
  attachments?: Array<{
    url?: string;
    file_type?: string;
    file_size?: number | string;
    is_encrypted?: boolean;
  }>;
  envelope?: {
    nonce?: string; // base64url
    ct?: string;    // base64url
  };
};

export async function buildDecryptedAttachment(
  msg: AttachmentMsg,
  shared: Uint8Array,
  trackBlobUrl?: (url: string) => void,
): Promise<{ url: string; mime: string; size: number; isImage: boolean; revoke: () => void }> {
  const mime = msg.attachments?.[0]?.file_type || "application/octet-stream";
  const size = Number(msg.attachments?.[0]?.file_size || 0);

  // A) External ciphertext (recommended for large files)
  if (msg.attachments?.[0]?.is_encrypted && msg.attachments[0].url) {
    const nonceB64u = msg.envelope?.nonce;
    if (!nonceB64u) throw new Error("Missing nonce for encrypted attachment");

    const resp = await fetch(msg.attachments[0].url, { mode: "cors" });
    const ctBuf = new Uint8Array(await resp.arrayBuffer());

    const plain = await decryptBytes(shared, nonceB64u, u8ToB64url(ctBuf));
    const blobUrl = URL.createObjectURL(
  new Blob([toArrayBuffer(plain)], { type: mime })
);
    trackBlobUrl?.(blobUrl);

    return {
      url: blobUrl,
      mime,
      size: plain.byteLength,
      isImage: mime.startsWith("image/"),
      revoke: () => URL.revokeObjectURL(blobUrl),
    };
  }

  // B) Inline ciphertext inside envelope.ct
  if (msg.envelope?.ct && msg.envelope?.nonce) {
    const plain = await decryptBytes(shared, msg.envelope.nonce, msg.envelope.ct);
   const blobUrl = URL.createObjectURL(
  new Blob([toArrayBuffer(plain)], { type: mime })
);
    trackBlobUrl?.(blobUrl);

    return {
      url: blobUrl,
      mime,
      size: plain.byteLength,
      isImage: mime.startsWith("image/"),
      revoke: () => URL.revokeObjectURL(blobUrl),
    };
  }

  // C) Plain (legacy)
  if (msg.attachments?.[0]?.url) {
    const blobUrl = msg.attachments[0].url;
    trackBlobUrl?.(blobUrl);
    return {
      url: blobUrl,
      mime,
      size,
      isImage: mime.startsWith("image/"),
      revoke: () => { /* remote URL: nothing to revoke */ },
    };
  }

  throw new Error("No attachment payload found");
}
