import sodium from 'libsodium-wrappers';
import { hkdf } from './hkdf';
import { encryptMessage, decryptMessage, encryptBytes, decryptBytes } from './aes_gcm';
import { RatchetState } from './types';

const te = new TextEncoder();

/** Return a fresh Uint8Array to avoid aliasing. */
function freshU8(view: Uint8Array | ArrayBuffer): Uint8Array {
  if (view instanceof Uint8Array) {
    const out = new Uint8Array(view.length);
    out.set(view);
    return out;
  }
  return new Uint8Array(view.slice(0));
}

/** Derive root key and chain key via HKDF. */
async function kdf_rk(rk: Uint8Array, dh_out: Uint8Array): Promise<{ newRk: Uint8Array; chainKey: Uint8Array }> {
  const input = new Uint8Array([...rk, ...dh_out]);
  const output = await hkdf(input, new Uint8Array(32), te.encode('ratchet-rk'), 64);
  return { newRk: output.slice(0, 32), chainKey: output.slice(32) };
}

/** Derive next chain key and message key via HKDF. */
async function kdf_ck(ck: Uint8Array): Promise<{ newCk: Uint8Array; mk: Uint8Array }> {
  const output = await hkdf(ck, new Uint8Array([0x01]), te.encode('ratchet-ck'), 64);
  return { newCk: output.slice(32), mk: output.slice(0, 32) };
}

/** Encrypt text with Double Ratchet. */
export async function ratchetEncrypt(
  state: RatchetState,
  plaintext: string
): Promise<{ header: { dhPub: string; msgNum: number }; ciphertext: string; nonce: string }> {
  await sodium.ready;

  if (state.dhRatchetNeeded && state.remoteDhPub) {
    state.dhPair = sodium.crypto_box_keypair();
    const dhOut = sodium.crypto_scalarmult(state.dhPair.privateKey, state.remoteDhPub);
    const { newRk, chainKey } = await kdf_rk(state.rootKey, freshU8(dhOut));
    state.rootKey = newRk;
    state.sendChainKey = chainKey;
    state.dhRatchetNeeded = false;
  }

  const { newCk, mk } = await kdf_ck(state.sendChainKey);
  state.sendChainKey = newCk;
  state.sendCount++;

  const { nonce, ciphertext } = await encryptMessage(mk, plaintext);
  const header = {
    dhPub: sodium.to_base64(state.dhPair.publicKey, sodium.base64_variants.URLSAFE_NO_PADDING),
    msgNum: state.sendCount
  };

  return { header, ciphertext, nonce };
}

/** Encrypt binary data with Double Ratchet. */
export async function ratchetEncryptBytes(
  state: RatchetState,
  bytes: Uint8Array
): Promise<{ header: { dhPub: string; msgNum: number }; ciphertext: string; nonce: string }> {
  await sodium.ready;

  if (state.dhRatchetNeeded && state.remoteDhPub) {
    state.dhPair = sodium.crypto_box_keypair();
    const dhOut = sodium.crypto_scalarmult(state.dhPair.privateKey, state.remoteDhPub);
    const { newRk, chainKey } = await kdf_rk(state.rootKey, freshU8(dhOut));
    state.rootKey = newRk;
    state.sendChainKey = chainKey;
    state.dhRatchetNeeded = false;
  }

  const { newCk, mk } = await kdf_ck(state.sendChainKey);
  state.sendChainKey = newCk;
  state.sendCount++;

  const { nonce, ciphertext } = await encryptBytes(mk, bytes);
  const header = {
    dhPub: sodium.to_base64(state.dhPair.publicKey, sodium.base64_variants.URLSAFE_NO_PADDING),
    msgNum: state.sendCount
  };

  return { header, ciphertext, nonce };
}

/** Decrypt text with Double Ratchet. */
export async function ratchetDecrypt(
  state: RatchetState,
  header: { dhPub: string; msgNum: number },
  ciphertext: string,
  nonce: string
): Promise<string> {
  await sodium.ready;

  const remoteDhPub = sodium.from_base64(header.dhPub, sodium.base64_variants.URLSAFE_NO_PADDING);
  if (!state.remoteDhPub || !sodium.memcmp(state.remoteDhPub, remoteDhPub)) {
    state.remoteDhPub = freshU8(remoteDhPub);
    const dhOut = sodium.crypto_scalarmult(state.dhPair.privateKey, state.remoteDhPub);
    const { newRk, chainKey } = await kdf_rk(state.rootKey, freshU8(dhOut));
    state.rootKey = newRk;
    state.recvChainKey = chainKey;
    state.recvCount = 0;
    state.dhRatchetNeeded = true;
  }

  const keyId = `${header.dhPub}:${header.msgNum}`;
  if (state.skippedKeys.has(keyId)) {
    const mk = state.skippedKeys.get(keyId)!;
    state.skippedKeys.delete(keyId);
    return decryptMessage(mk, nonce, ciphertext);
  }

  if (header.msgNum > state.recvCount) {
    for (let i = state.recvCount; i < header.msgNum; i++) {
      const { newCk, mk: skippedMk } = await kdf_ck(state.recvChainKey);
      state.recvChainKey = newCk;
      state.skippedKeys.set(`${header.dhPub}:${i}`, skippedMk);
    }
    state.recvCount = header.msgNum;
  }

  const { newCk, mk } = await kdf_ck(state.recvChainKey);
  state.recvChainKey = newCk;
  state.recvCount++;

  return decryptMessage(mk, nonce, ciphertext);
}

/** Decrypt binary data with Double Ratchet. */
export async function ratchetDecryptBytes(
  state: RatchetState,
  header: { dhPub: string; msgNum: number },
  ciphertext: string,
  nonce: string
): Promise<Uint8Array> {
  await sodium.ready;

  const remoteDhPub = sodium.from_base64(header.dhPub, sodium.base64_variants.URLSAFE_NO_PADDING);
  if (!state.remoteDhPub || !sodium.memcmp(state.remoteDhPub, remoteDhPub)) {
    state.remoteDhPub = freshU8(remoteDhPub);
    const dhOut = sodium.crypto_scalarmult(state.dhPair.privateKey, state.remoteDhPub);
    const { newRk, chainKey } = await kdf_rk(state.rootKey, freshU8(dhOut));
    state.rootKey = newRk;
    state.recvChainKey = chainKey;
    state.recvCount = 0;
    state.dhRatchetNeeded = true;
  }

  const keyId = `${header.dhPub}:${header.msgNum}`;
  if (state.skippedKeys.has(keyId)) {
    const mk = state.skippedKeys.get(keyId)!;
    state.skippedKeys.delete(keyId);
    return decryptBytes(mk, nonce, ciphertext);
  }

  if (header.msgNum > state.recvCount) {
    for (let i = state.recvCount; i < header.msgNum; i++) {
      const { newCk, mk: skippedMk } = await kdf_ck(state.recvChainKey);
      state.recvChainKey = newCk;
      state.skippedKeys.set(`${header.dhPub}:${i}`, skippedMk);
    }
    state.recvCount = header.msgNum;
  }

  const { newCk, mk } = await kdf_ck(state.recvChainKey);
  state.recvChainKey = newCk;
  state.recvCount++;

  return decryptBytes(mk, nonce, ciphertext);
}

/** Initialize RatchetState after X3DH. */
export async function initRatchetState(sharedSecret: Uint8Array): Promise<RatchetState> {
  await sodium.ready;
  const dhPair = sodium.crypto_box_keypair();
  return {
    rootKey: freshU8(sharedSecret),
    sendChainKey: freshU8(sharedSecret),
    recvChainKey: freshU8(sharedSecret),
    sendCount: 0,
    recvCount: 0,
    dhPair,
    remoteDhPub: null,
    dhRatchetNeeded: true,
    skippedKeys: new Map()
  };
}