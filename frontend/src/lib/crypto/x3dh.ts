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
// One-Time PreKeys
const opks: OneTimePreKey[] = [];
const opkPrivates: Record<string, Uint8Array> = {}; // local storage map
for (let i = 0; i < 20; i++) {
  const opk = sodium.crypto_box_keypair();
  const keyId = crypto.randomUUID();
  opks.push({
    keyId,
    publicKey: sodium.to_base64(opk.publicKey)
  });
  opkPrivates[keyId] = opk.privateKey; // persist locally (IndexedDB or Zustand)
}

// When returning, store private keys somewhere safe locally:
localStorage.setItem(`x3dh_${userId}_ikPriv`, sodium.to_base64(ik.privateKey));
localStorage.setItem(`x3dh_${userId}_spkPriv`, sodium.to_base64(spk.privateKey));
localStorage.setItem(`x3dh_${userId}_opkPrivs`, JSON.stringify(
  Object.fromEntries(Object.entries(opkPrivates).map(([k, v]) => [k, sodium.to_base64(v)]))
));

  return {
    userId,
    identityKey: sodium.to_base64(ik.publicKey),
    signedPreKey: sodium.to_base64(spk.publicKey),
    spkSignature: sodium.to_base64(spkSignature),
    oneTimePreKeys: opks,
  };
}

export async function deriveSharedSecretInitiator(
  ourIKPrivEd25519: Uint8Array,    // our identity private key (ed25519 secret key)
  recipientBundle: BundleResponse
): Promise<SessionInitResult> {
  await sodium.ready;

  // decode incoming
  const recipientIKEd = sodium.from_base64(recipientBundle.identityKey);
  const recipientSPK = sodium.from_base64(recipientBundle.signedPreKey);
  const recipientOPK = recipientBundle.oneTimePreKey ? sodium.from_base64(recipientBundle.oneTimePreKey) : null;

  // generate ephemeral curve25519 keypair for this session (use crypto_box keypair => curve25519)
  const ek = sodium.crypto_box_keypair(); // ek.publicKey, ek.privateKey (curve25519)

  // Convert our ed25519 identity private -> curve25519 for DH (responder will convert their ed25519 private too)
  const ourIKCurvePriv = sodium.crypto_sign_ed25519_sk_to_curve25519(ourIKPrivEd25519);

  // Convert recipient IK ed25519 pub -> curve25519 pub
  const recipientIKCurvePub = sodium.crypto_sign_ed25519_pk_to_curve25519(recipientIKEd);

  // DHs (X25519 scalar multiplications)
  // DH1: IKa(priv_curve) x SPKb(pub)
  const dh1 = sodium.crypto_scalarmult(ourIKCurvePriv, recipientSPK);
  // DH2: EKa(priv) x IKb(pub_curve)
  const dh2 = sodium.crypto_scalarmult(ek.privateKey, recipientIKCurvePub);
  // DH3: EKa(priv) x SPKb(pub)
  const dh3 = sodium.crypto_scalarmult(ek.privateKey, recipientSPK);
  // Optional DH4: EKa(priv) x OPKb(pub)
  const dh4 = recipientOPK ? sodium.crypto_scalarmult(ek.privateKey, recipientOPK) : new Uint8Array();

  // concatenate
  const concatenated = new Uint8Array(dh1.length + dh2.length + dh3.length + dh4.length);
  let off = 0;
  concatenated.set(dh1, off); off += dh1.length;
  concatenated.set(dh2, off); off += dh2.length;
  concatenated.set(dh3, off); off += dh3.length;
  if (dh4.length) concatenated.set(dh4, off);

  // KDF: hash into 32-byte symmetric key (you can use HKDF too; using generic hash here)
  const sharedSecret = sodium.crypto_generichash(32, concatenated);

  return {
    sharedSecret,
    ephemeralPubKey: ek.publicKey
  };
}


export async function deriveSharedSecretResponder(
  ourIKPrivEd25519: Uint8Array,
  ourSPKPrivEd25519: Uint8Array,
  initiatorEphemeralPub: Uint8Array,
  initiatorIKPubEd25519: Uint8Array,
  ourOPKPrivCurve?: Uint8Array   // optional last
): Promise<Uint8Array> {

  await sodium.ready;

  // Convert Ed25519 keys to Curve25519
  const ourIKCurvePriv = sodium.crypto_sign_ed25519_sk_to_curve25519(ourIKPrivEd25519);
  const ourSPKCurvePriv = sodium.crypto_sign_ed25519_sk_to_curve25519(ourSPKPrivEd25519);
  const initiatorIKCurvePub = sodium.crypto_sign_ed25519_pk_to_curve25519(initiatorIKPubEd25519);

  // DH computations
  const dh1 = sodium.crypto_scalarmult(ourSPKCurvePriv, initiatorIKCurvePub);   // SPKr x IKa
  const dh2 = sodium.crypto_scalarmult(ourIKCurvePriv, initiatorEphemeralPub);  // IKr x EKa
  const dh3 = sodium.crypto_scalarmult(ourSPKCurvePriv, initiatorEphemeralPub); // SPKr x EKa
  const dh4 = ourOPKPrivCurve
    ? sodium.crypto_scalarmult(ourOPKPrivCurve, initiatorEphemeralPub)          // OPKr x EKa
    : new Uint8Array();

  // Concatenate DH outputs
  const concatenated = new Uint8Array(dh1.length + dh2.length + dh3.length + dh4.length);
  let off = 0;
  concatenated.set(dh1, off); off += dh1.length;
  concatenated.set(dh2, off); off += dh2.length;
  concatenated.set(dh3, off); off += dh3.length;
  if (dh4.length) concatenated.set(dh4, off);

  // Hash into 32-byte shared secret
  return sodium.crypto_generichash(32, concatenated);
}
