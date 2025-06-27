// lib/crypto/x3dh.ts
import sodium from 'libsodium-wrappers';

export interface OneTimePreKey {
  keyId: string;
  publicKey: string;
}

export interface RegisterBundleRequest {
  userId: string;
  identityKey: string;
  signedPreKey: string;
  spkSignature: string;
  oneTimePreKeys: OneTimePreKey[];
}
// types/x3dh.ts
export interface BundleResponse {
  identityKey: string;
  signedPreKey: string;
  spkSignature: string;
  oneTimePreKey?: string;
}

export interface SessionInitResult {
  sharedSecret: Uint8Array;
  ephemeralPubKey: Uint8Array;
}


export async function generateX3DHKeyBundle(userId: string): Promise<RegisterBundleRequest> {
  await sodium.ready;

  // Identity Key
  const ik = sodium.crypto_sign_keypair();

  // Signed PreKey
  const spk = sodium.crypto_sign_keypair();
  const spkSignature = sodium.crypto_sign_detached(spk.publicKey, ik.privateKey);

  // One-Time PreKeys
  const opks: OneTimePreKey[] = [];
  for (let i = 0; i < 20; i++) {
    const opk = sodium.crypto_box_keypair();
    opks.push({
      keyId: crypto.randomUUID(),
      publicKey: sodium.to_base64(opk.publicKey)
    });
  }

  return {
    userId,
    identityKey: sodium.to_base64(ik.publicKey),
    signedPreKey: sodium.to_base64(spk.publicKey),
    spkSignature: sodium.to_base64(spkSignature),
    oneTimePreKeys: opks,
  };
}

export async function deriveSharedSecretInitiator(
  ourIK: Uint8Array,
  recipientBundle: BundleResponse
): Promise<Uint8Array> {
  await sodium.ready;

  const recipientIK = sodium.from_base64(recipientBundle.identityKey);
  const recipientSPK = sodium.from_base64(recipientBundle.signedPreKey);
  const recipientOPK = recipientBundle.oneTimePreKey
    ? sodium.from_base64(recipientBundle.oneTimePreKey)
    : null;

  const ek = sodium.crypto_box_keypair(); // Ephemeral Key

  // 1. DH1: DH(IKa, SPKb)
  const dh1 = sodium.crypto_scalarmult(ourIK, recipientSPK);

  // 2. DH2: DH(EKa, IKb)
  const dh2 = sodium.crypto_scalarmult(ek.privateKey, recipientIK);

  // 3. DH3: DH(EKa, SPKb)
  const dh3 = sodium.crypto_scalarmult(ek.privateKey, recipientSPK);

  // 4. DH4: DH(EKa, OPKb) — optional if OPK exists
  const dh4 = recipientOPK
    ? sodium.crypto_scalarmult(ek.privateKey, recipientOPK)
    : new Uint8Array();

  // Concatenate shared secrets
// Concatenate all DH results into one buffer
const concatenated = new Uint8Array(
  dh1.length + dh2.length + dh3.length + dh4.length
);
concatenated.set(dh1, 0);
concatenated.set(dh2, dh1.length);
concatenated.set(dh3, dh1.length + dh2.length);
concatenated.set(dh4, dh1.length + dh2.length + dh3.length);

// Hash into final shared secret
const sharedSecret = sodium.crypto_generichash(32, concatenated);


  return sharedSecret;
}
