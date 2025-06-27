import * as _sodium from 'libsodium-wrappers';

let sodium: typeof _sodium;

export async function getSodium() {
  if (!sodium) {
    await _sodium.ready;
    sodium = _sodium;
  }
  return sodium;
}
