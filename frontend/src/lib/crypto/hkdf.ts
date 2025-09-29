// src/lib/crypto/hkdf.ts

function toArrayBuffer(u8: Uint8Array): ArrayBuffer {
  const ab = new ArrayBuffer(u8.byteLength);
  new Uint8Array(ab).set(u8);
  return ab;
}

function importHmacKey(raw: ArrayBuffer | Uint8Array): Promise<CryptoKey> {
  const material = raw instanceof Uint8Array ? toArrayBuffer(raw) : raw;
  return crypto.subtle.importKey(
    "raw",
    material,
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign"]
  );
}

async function hkdfExtract(ikm: Uint8Array, salt?: Uint8Array): Promise<ArrayBuffer> {
  const saltBytes = salt && salt.length ? salt : new Uint8Array(32);
  const saltKey = await importHmacKey(saltBytes);
  const prk = await crypto.subtle.sign("HMAC", saltKey, toArrayBuffer(ikm));
  return prk;
}

async function hkdfExpand(prk: ArrayBuffer, info: Uint8Array, length: number): Promise<Uint8Array> {
  const hashLen = 32;
  const n = Math.ceil(length / hashLen);
  if (n > 255) throw new Error("hkdfExpand: length too large");

  const prkKey = await importHmacKey(prk);

  const okm = new Uint8Array(length);
  let prev = new Uint8Array(0);
  let offset = 0;

  for (let i = 1; i <= n; i++) {
    const input = new Uint8Array(prev.length + info.length + 1);
    input.set(prev, 0);
    input.set(info, prev.length);
    input.set([i], prev.length + info.length);

    const blockBuf = await crypto.subtle.sign("HMAC", prkKey, toArrayBuffer(input));
    const block = new Uint8Array(blockBuf);

    const take = Math.min(hashLen, length - offset);
    okm.set(block.subarray(0, take), offset);
    offset += take;
    prev = block;
  }

  return okm;
}

export async function hkdf(
  ikm: Uint8Array,
  salt: Uint8Array | undefined,
  info: Uint8Array,
  length: number
): Promise<Uint8Array> {
  const prk = await hkdfExtract(ikm, salt);
  return hkdfExpand(prk, info, length);
}
