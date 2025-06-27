import * as sodium from 'libsodium-wrappers';



export function encryptMessage(message: string, sharedSecret: Uint8Array): { ciphertext: Uint8Array, nonce: Uint8Array } {
  const nonce = sodium.randombytes_buf(sodium.crypto_secretbox_NONCEBYTES);
  const ciphertext = sodium.crypto_secretbox_easy(
    sodium.from_string(message),
    nonce,
    sharedSecret
  );
  return { ciphertext, nonce };
}

export function decryptMessage(ciphertext: Uint8Array, nonce: Uint8Array, sharedSecret: Uint8Array): string | null {
  const decrypted = sodium.crypto_secretbox_open_easy(
    ciphertext,
    nonce,
    sharedSecret
  );
  return decrypted ? sodium.to_string(decrypted) : null;
}
