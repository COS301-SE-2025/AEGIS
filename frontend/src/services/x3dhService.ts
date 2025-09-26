// services/x3dhService.ts
import axios from "axios";

type BundleResponseWire = {
  identity_key: string;
  signed_prekey: string;
  spk_signature: string;
  one_time_prekey?: string;
  opk_id?: string; // backend includes this; handy if you want to debug
};

type OPKCountWire = {
  user_id: string;
  available_opks: number;
};

export type BundleResponse = {
  identityKey: string;
  signedPreKey: string;
  spkSignature: string;
  oneTimePreKey?: string;
  opkId?: string;
};

const API = (import.meta as any).env?.VITE_API_BASE?.replace(/\/$/, "") || "http://localhost:8080";
const X3DH_BASE = `${API}/api/v1/x3dh`;

function authHeaders() {
  const token = sessionStorage.getItem("authToken") || "";
  return token ? { Authorization: `Bearer ${token}` } : {};
}

// ------- register-bundle -------
type RegisterBundleWire = {
  user_id: string;
  identity_key: string;
  signed_prekey: string;
  spk_signature: string;
  one_time_prekeys: { key_id: string; public_key: string }[];
};

export async function uploadKeyBundle(bundle: {
  userId: string;
  identityKey: string;
  signedPreKey: string;
  spkSignature: string;
  oneTimePreKeys: { keyId: string; publicKey: string }[];
}): Promise<void> {
  const wire: RegisterBundleWire = {
    user_id: bundle.userId,
    identity_key: bundle.identityKey,
    signed_prekey: bundle.signedPreKey,
    spk_signature: bundle.spkSignature,
    one_time_prekeys: bundle.oneTimePreKeys.map(k => ({
      key_id: k.keyId,
      public_key: k.publicKey,
    })),
  };
  await axios.post(`${X3DH_BASE}/register-bundle`, wire, { headers: authHeaders() });
}

// ------- bundle GET -------
export async function fetchBundle(userId: string): Promise<BundleResponse> {
  const { data } = await axios.get<BundleResponseWire>(`${X3DH_BASE}/bundle/${userId}`, {
    headers: authHeaders(),
  });
  return {
    identityKey: data.identity_key,
    signedPreKey: data.signed_prekey,
    spkSignature: data.spk_signature,
    oneTimePreKey: data.one_time_prekey,
    opkId: data.opk_id,
  };
}

// ------- OPK count -------
export async function getAvailableOPKCount(userId: string): Promise<number> {
  const { data } = await axios.get<OPKCountWire>(`${X3DH_BASE}/opk-count/${userId}`, {
    headers: authHeaders(),
  });
  return data.available_opks;
}

// ------- refill-opks -------
export async function refillOPKs(
  userId: string,
  newOPKs: { keyId: string; publicKey: string }[]
): Promise<void> {
  const wire = {
    user_id: userId,
    opks: newOPKs.map(k => ({ key_id: k.keyId, public_key: k.publicKey })),
  };
  await axios.post(`${X3DH_BASE}/refill-opks`, wire, { headers: authHeaders() });
}
