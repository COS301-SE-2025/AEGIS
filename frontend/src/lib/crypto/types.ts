export interface RatchetState {
  rootKey: Uint8Array; // Root key from X3DH or previous DH ratchet
  sendChainKey: Uint8Array; // Current sending chain key
  recvChainKey: Uint8Array; // Current receiving chain key
  sendCount: number; // Number of messages sent in this chain
  recvCount: number; // Number of messages received in this chain
  dhPair: { publicKey: Uint8Array; privateKey: Uint8Array }; // Current X25519 DH keypair
  remoteDhPub: Uint8Array | null; // Remote party's DH public key
  dhRatchetNeeded: boolean; // Whether a DH ratchet is needed
  skippedKeys: Map<string, Uint8Array>; // Skipped message keys for out-of-order messages
}


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

export interface ChatMember {
  id: string; // UUID
  email?: string; // Optional for display
  name?: string;
  publicKey?: Uint8Array;
}

export interface Chat {
  id: string; // Group ID (UUID)
  name?: string;
  members: ChatMember[]; // Array of { id: UUID, email: string }
}