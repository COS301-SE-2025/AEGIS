export function encodeBase64(data: Uint8Array): string {
  return btoa(String.fromCharCode(...data));
}

export function decodeBase64(base64: string): Uint8Array {
  const binaryString = atob(base64);
  const len = binaryString.length;
  const bytes = new Uint8Array(len);
  for (let i = 0; i < len; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes;
}

// utils/b64.ts
import sodium from "libsodium-wrappers";
export const b64e = (u8: Uint8Array) =>
  sodium.to_base64(u8, sodium.base64_variants.ORIGINAL);
export const b64d = (s: string) =>
  sodium.from_base64(s, sodium.base64_variants.ORIGINAL); // => Uint8Array
