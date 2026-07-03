import { _resolveCallable } from './index.js';

// PHP Regex functions
export function preg_match(pattern, subject, matches) {
  const { regex, flags } = _parsePattern(pattern);
  let re;
  try { re = new RegExp(regex, flags.replace('g', '')); }
  catch (e) { /* console.warn('preg_match: ' + e.message); */ return 0; }
  const m = String(subject).match(re);
  if (!m) return 0;
  if (matches && typeof matches === 'object') {
    // Populate matches array
    for (let i = 0; i < m.length; i++) matches[i] = m[i] ?? '';
  }
  return 1;
}

export function preg_match_all(pattern, subject, matches) {
  const { regex, flags } = _parsePattern(pattern);
  let re;
  try { re = new RegExp(regex, flags.includes('g') ? flags : flags + 'g'); }
  catch (e) { console.warn('preg_match_all: ' + e.message); return 0; }
  const allMatches = [...String(subject).matchAll(re)];
  if (matches && typeof matches === 'object') {
    // Group by capture group index
    const maxGroups = allMatches.reduce((max, m) => Math.max(max, m.length), 0);
    for (let i = 0; i < maxGroups; i++) {
      matches[i] = allMatches.map(m => m[i] ?? '');
    }
  }
  return allMatches.length;
}

export function preg_replace(pattern, replacement, subject) {
  if (Array.isArray(pattern)) {
    let result = String(subject);
    for (let i = 0; i < pattern.length; i++) {
      const rep = Array.isArray(replacement) ? (replacement[i] ?? '') : replacement;
      result = preg_replace(pattern[i], rep, result);
    }
    return result;
  }
  const { regex, flags } = _parsePattern(pattern);
  let re;
  try { re = new RegExp(regex, flags.includes('g') ? flags : flags + 'g'); }
  catch (e) { console.warn('preg_replace: ' + e.message); return String(subject); }
  // Convert PHP backreferences ($1) to JS ($1 is same)
  const jsReplacement = String(replacement).replace(/\\(\d+)/g, '$$$1');
  return String(subject).replace(re, jsReplacement);
}

export function preg_replace_callback(pattern, callback, subject) {
  const { regex, flags } = _parsePattern(pattern);
  let re;
  try { re = new RegExp(regex, flags.includes('g') ? flags : flags + 'g'); }
  catch (e) { console.warn('preg_replace_callback: ' + e.message); return String(subject); }
  let cb = _resolveCallable(callback);
  return String(subject).replace(re, (...args) => {
    const matches = args.slice(0, -2); // Remove offset and full string
    return cb(matches);
  });
}

export function preg_split(pattern, subject, limit = -1) {
  const { regex, flags } = _parsePattern(pattern);
  let re;
  try { re = new RegExp(regex, flags); }
  catch (e) { console.warn('preg_split: ' + e.message); return [String(subject)]; }
  const parts = String(subject).split(re);
  if (limit > 0 && parts.length > limit) {
    const result = parts.slice(0, limit - 1);
    result.push(parts.slice(limit - 1).join(''));
    return result;
  }
  return parts;
}

// Parse PHP regex pattern like /pattern/flags
function _parsePattern(pattern) {
  const s = String(pattern);
  if (!s || s.length < 2) return { regex: s, flags: '' };

  const delim = s[0];
  
  // 最後のデリミタ位置を逆方向で探す（バックスラッシュエスケープ対応）
  let lastDelimIndex = -1;
  for (let i = s.length - 1; i >= 1; i--) {
    if (s[i] === delim) {
      let escapeCount = 0;
      for (let j = i - 1; j >= 0 && s[j] === '\\'; j--) {
        escapeCount++;
      }
      if (escapeCount % 2 === 0) {
        lastDelimIndex = i;
        break;
      }
    }
  }

  if (lastDelimIndex <= 0) {
    return { regex: s, flags: '' };
  }

  const regex = s.slice(1, lastDelimIndex);
  const flags = s.slice(lastDelimIndex + 1);
  const phpFlags = flags.replace(/g/g, ''); // g フラグ削除

  return { regex, flags: phpFlags };
}
