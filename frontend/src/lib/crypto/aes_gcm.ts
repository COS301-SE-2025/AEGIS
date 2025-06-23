// lib/crypto/aes/gcm.ts
import sodium from 'libsodium-wrappers';

/**
 * Encrypts a message using XChaCha20-Poly1305 (instead of AES-GCM)
 */
export async function encryptMessage(plaintext: string, key: Uint8Array) {
  await sodium.ready;

  const nonce = sodium.randombytes_buf(sodium.crypto_aead_xchacha20poly1305_ietf_NPUBBYTES);
  const ciphertext = sodium.crypto_aead_xchacha20poly1305_ietf_encrypt(
    sodium.from_string(plaintext),
    null,
    null,
    nonce,
    key
  );

  return {
    ciphertext,
    nonce,
  };
}

/**
 * Decrypts a ciphertext using XChaCha20-Poly1305
 */
export async function decryptMessage(ciphertext: Uint8Array, nonce: Uint8Array, key: Uint8Array) {
  await sodium.ready;

  const plaintext = sodium.crypto_aead_xchacha20poly1305_ietf_decrypt(
    null,
    ciphertext,
    null,
    nonce,
    key
  );

  return sodium.to_string(plaintext);
}
