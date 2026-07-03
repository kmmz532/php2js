// PHP Hash/Crypto functions
function _md5cmn(q, a, b, x, s, t) {
  a = (((a + q) + x) + t) | 0;
  return (((a << s) | (a >>> (32 - s))) + b) | 0;
}

function _md5ff(a, b, c, d, x, s, t) { return _md5cmn((b & c) | (~b & d), a, b, x, s, t); }
function _md5gg(a, b, c, d, x, s, t) { return _md5cmn((b & d) | (c & ~d), a, b, x, s, t); }
function _md5hh(a, b, c, d, x, s, t) { return _md5cmn(b ^ c ^ d, a, b, x, s, t); }
function _md5ii(a, b, c, d, x, s, t) { return _md5cmn(c ^ (b | ~d), a, b, x, s, t); }

function _md5blk(s) {
  const blocks = [];
  for (let i = 0; i < 64; i += 4) {
    blocks[i >> 2] = s.charCodeAt(i)
      + (s.charCodeAt(i + 1) << 8)
      + (s.charCodeAt(i + 2) << 16)
      + (s.charCodeAt(i + 3) << 24);
  }
  return blocks;
}

function _md51(s) {
  let n = s.length;
  let state = [1732584193, -271733879, -1732584194, 271733878];
  for (let i = 64; i <= n; i += 64) {
    state = _md5cycle(state, _md5blk(s.substring(i - 64, i)));
  }
  s = s.substring(n - 64);
  const tail = Array(16).fill(0);
  for (let i = 0; i < s.length; i++) {
    tail[i >> 2] |= s.charCodeAt(i) << ((i % 4) << 3);
  }
  tail[s.length >> 2] |= 0x80 << ((s.length % 4) << 3);
  if (s.length > 55) {
    state = _md5cycle(state, tail);
    for (let i = 0; i < 16; i++) tail[i] = 0;
  }
  tail[14] = n * 8;
  return _md5cycle(state, tail);
}

function _md5cycle(state, block) {
  let [a, b, c, d] = state;

  a = _md5ff(a, b, c, d, block[0], 7, -680876936);
  d = _md5ff(d, a, b, c, block[1], 12, -389564586);
  c = _md5ff(c, d, a, b, block[2], 17, 606105819);
  b = _md5ff(b, c, d, a, block[3], 22, -1044525330);
  a = _md5ff(a, b, c, d, block[4], 7, -176418897);
  d = _md5ff(d, a, b, c, block[5], 12, 1200080426);
  c = _md5ff(c, d, a, b, block[6], 17, -1473231341);
  b = _md5ff(b, c, d, a, block[7], 22, -45705983);
  a = _md5ff(a, b, c, d, block[8], 7, 1770035416);
  d = _md5ff(d, a, b, c, block[9], 12, -1958414417);
  c = _md5ff(c, d, a, b, block[10], 17, -42063);
  b = _md5ff(b, c, d, a, block[11], 22, -1990404162);
  a = _md5ff(a, b, c, d, block[12], 7, 1804603682);
  d = _md5ff(d, a, b, c, block[13], 12, -40341101);
  c = _md5ff(c, d, a, b, block[14], 17, -1502002290);
  b = _md5ff(b, c, d, a, block[15], 22, 1236535329);

  a = _md5gg(a, b, c, d, block[1], 5, -165796510);
  d = _md5gg(d, a, b, c, block[6], 9, -1069501632);
  c = _md5gg(c, d, a, b, block[11], 14, 643717713);
  b = _md5gg(b, c, d, a, block[0], 20, -373897302);
  a = _md5gg(a, b, c, d, block[5], 5, -701558691);
  d = _md5gg(d, a, b, c, block[10], 9, 38016083);
  c = _md5gg(c, d, a, b, block[15], 14, -660478335);
  b = _md5gg(b, c, d, a, block[4], 20, -405537848);
  a = _md5gg(a, b, c, d, block[9], 5, 568446438);
  d = _md5gg(d, a, b, c, block[14], 9, -1019803690);
  c = _md5gg(c, d, a, b, block[3], 14, -187363961);
  b = _md5gg(b, c, d, a, block[8], 20, 1163531501);
  a = _md5gg(a, b, c, d, block[13], 5, -1444681467);
  d = _md5gg(d, a, b, c, block[2], 9, -51403784);
  c = _md5gg(c, d, a, b, block[7], 14, 1735328473);
  b = _md5gg(b, c, d, a, block[12], 20, -1926607734);

  a = _md5hh(a, b, c, d, block[5], 4, -378558);
  d = _md5hh(d, a, b, c, block[8], 11, -2022574463);
  c = _md5hh(c, d, a, b, block[11], 16, 1839030562);
  b = _md5hh(b, c, d, a, block[14], 23, -35309556);
  a = _md5hh(a, b, c, d, block[1], 4, -1530992060);
  d = _md5hh(d, a, b, c, block[4], 11, 1272893353);
  c = _md5hh(c, d, a, b, block[7], 16, -155497632);
  b = _md5hh(b, c, d, a, block[10], 23, -1094730640);
  a = _md5hh(a, b, c, d, block[13], 4, 681279174);
  d = _md5hh(d, a, b, c, block[0], 11, -358537222);
  c = _md5hh(c, d, a, b, block[3], 16, -722521979);
  b = _md5hh(b, c, d, a, block[6], 23, 76029189);
  a = _md5hh(a, b, c, d, block[9], 4, -640364487);
  d = _md5hh(d, a, b, c, block[12], 11, -421815835);
  c = _md5hh(c, d, a, b, block[15], 16, 530742520);
  b = _md5hh(b, c, d, a, block[2], 23, -995338651);

  a = _md5ii(a, b, c, d, block[0], 6, -198630844);
  d = _md5ii(d, a, b, c, block[7], 10, 1126891415);
  c = _md5ii(c, d, a, b, block[14], 15, -1416354905);
  b = _md5ii(b, c, d, a, block[5], 21, -57434055);
  a = _md5ii(a, b, c, d, block[12], 6, 1700485571);
  d = _md5ii(d, a, b, c, block[3], 10, -1894986606);
  c = _md5ii(c, d, a, b, block[10], 15, -1051523);
  b = _md5ii(b, c, d, a, block[1], 21, -2054922799);
  a = _md5ii(a, b, c, d, block[8], 6, 1873313359);
  d = _md5ii(d, a, b, c, block[15], 10, -30611744);
  c = _md5ii(c, d, a, b, block[6], 15, -1560198380);
  b = _md5ii(b, c, d, a, block[13], 21, 1309151649);
  a = _md5ii(a, b, c, d, block[4], 6, -145523070);
  d = _md5ii(d, a, b, c, block[11], 10, -1120210379);
  c = _md5ii(c, d, a, b, block[2], 15, 718787259);
  b = _md5ii(b, c, d, a, block[9], 21, -343485551);

  state[0] = (state[0] + a) | 0;
  state[1] = (state[1] + b) | 0;
  state[2] = (state[2] + c) | 0;
  state[3] = (state[3] + d) | 0;
  return state;
}

function _md5hex(x) {
  const hex = [];
  for (let i = 0; i < x.length; i++) {
    for (let j = 0; j < 4; j++) {
      hex.push(((x[i] >> (j * 8)) & 0xff).toString(16).padStart(2, '0'));
    }
  }
  return hex.join('');
}

export async function md5(s) {
  const data = unescape(encodeURIComponent(String(s)));
  return _md5hex(_md51(data));
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
  if (webAlgo === 'MD5') return await md5(s);
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
