// PHP Hash/Crypto functions
export async function md5(s) {
  const data = new TextEncoder().encode(String(s));
  const hash = await crypto.subtle.digest('MD5', data);
  return Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2, '0')).join('');
}

export async function sha1(s) {
  const data = new TextEncoder().encode(String(s));
  const hash = await crypto.subtle.digest('SHA-1', data);
  return Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2, '0')).join('');
}

export async function hash(algo, s) {
  const algoMap = { 'md5': 'MD5', 'sha1': 'SHA-1', 'sha256': 'SHA-256', 'sha384': 'SHA-384', 'sha512': 'SHA-512' };
  const webAlgo = algoMap[algo.toLowerCase()];
  if (!webAlgo) { console.warn(`Unsupported hash algo: ${algo}`); return false; }
  const data = new TextEncoder().encode(String(s));
  const hashBuf = await crypto.subtle.digest(webAlgo, data);
  return Array.from(new Uint8Array(hashBuf)).map(b => b.toString(16).padStart(2, '0')).join('');
}

export function crc32(s) {
  let crc = 0xFFFFFFFF;
  for (let i = 0; i < s.length; i++) {
    crc ^= s.charCodeAt(i);
    for (let j = 0; j < 8; j++) crc = (crc >>> 1) ^ (crc & 1 ? 0xEDB88320 : 0);
  }
  return (crc ^ 0xFFFFFFFF) | 0;
}

export function base64_encode(s) {
  if (typeof btoa === 'function') return btoa(unescape(encodeURIComponent(String(s))));
  return Buffer.from(String(s)).toString('base64');
}

export function base64_decode(s) {
  if (typeof atob === 'function') return decodeURIComponent(escape(atob(String(s))));
  return Buffer.from(String(s), 'base64').toString();
}
