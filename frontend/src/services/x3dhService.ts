import axios from 'axios';
import { RegisterBundleRequest } from '../lib/crypto/x3dh';

const API_BASE = '/api/v1/x3dh';

export interface BundleResponse {
  identityKey: string;
  signedPreKey: string;
  spkSignature: string;
  oneTimePreKey?: string;
}

export async function uploadKeyBundle(bundle: RegisterBundleRequest): Promise<void> {
  await axios.post(`${API_BASE}/register-bundle`, bundle);
}

export async function fetchBundle(userId: string): Promise<BundleResponse> {
  const res = await axios.get<BundleResponse>(`${API_BASE}/bundle/${userId}`);
  return res.data;
}

export async function getAvailableOPKCount(userId: string): Promise<number> {
  const res = await axios.get<{ count: number }>(`${API_BASE}/opk-count/${userId}`);
  return res.data.count;
}

export async function refillOPKs(
  userId: string,
  newOPKs: { keyId: string; publicKey: string }[]
): Promise<void> {
  await axios.post(`${API_BASE}/refill-opks`, {
    userId,
    oneTimePreKeys: newOPKs,
  });
}
