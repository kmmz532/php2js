// PHP Array functions for JS runtime

export function count(v) {
  if (v == null) return 0;
  if (Array.isArray(v)) return v.length;
  if (typeof v === 'object') return Object.keys(v).length;
  return 0;
}

export function array_push(arr, ...values) {
  if (Array.isArray(arr)) { arr.push(...values); return arr.length; }
  return 0;
}

export function array_pop(arr) { return Array.isArray(arr) ? arr.pop() : null; }
export function array_shift(arr) { return Array.isArray(arr) ? arr.shift() : null; }
export function array_unshift(arr, ...values) {
  if (Array.isArray(arr)) { arr.unshift(...values); return arr.length; }
  return 0;
}

export function array_merge(...arrays) {
  const result = [];
  for (const arr of arrays) {
    if (Array.isArray(arr)) result.push(...arr);
    else if (typeof arr === 'object' && arr !== null) Object.assign(result, arr);
  }
  return result;
}

export function array_slice(arr, offset, length, preserveKeys = false) {
  if (Array.isArray(arr)) {
    if (length === undefined) return arr.slice(offset);
    return arr.slice(offset, offset + length);
  }
  const entries = Object.entries(arr || {});
  const sliced = length === undefined ? entries.slice(offset) : entries.slice(offset, offset + length);
  if (preserveKeys) return Object.fromEntries(sliced);
  return sliced.map(([, v]) => v);
}

export function array_splice(arr, start, deleteCount, ...items) {
  if (!Array.isArray(arr)) return [];
  return arr.splice(start, deleteCount ?? arr.length, ...items);
}

export function array_keys(arr) {
  if (Array.isArray(arr)) return arr.map((_, i) => i);
  return Object.keys(arr || {});
}

export function array_values(arr) {
  if (Array.isArray(arr)) return [...arr];
  return Object.values(arr || {});
}

export function array_flip(arr) {
  const result = {};
  const entries = Array.isArray(arr) ? arr.map((v, i) => [i, v]) : Object.entries(arr || {});
  for (const [k, v] of entries) result[v] = k;
  return result;
}

export function array_reverse(arr, preserveKeys = false) {
  if (Array.isArray(arr)) return [...arr].reverse();
  const entries = Object.entries(arr || {}).reverse();
  return Object.fromEntries(entries);
}

export function array_map(callback, arr, ...extra) {
  if (callback === null) {
    // array_map(null, $a, $b) -> zip
    const arrays = [arr, ...extra];
    const maxLen = Math.max(...arrays.map(a => Array.isArray(a) ? a.length : Object.keys(a || {}).length));
    const result = [];
    for (let i = 0; i < maxLen; i++) result.push(arrays.map(a => Array.isArray(a) ? a[i] : undefined));
    return result;
  }
  if (Array.isArray(arr)) return arr.map((v, i) => callback(v, ...extra.map(a => a?.[i])));
  const result = {};
  for (const [k, v] of Object.entries(arr || {})) result[k] = callback(v);
  return result;
}

export function array_filter(arr, callback, flag = 0) {
  if (Array.isArray(arr)) {
    if (!callback) return arr.filter(Boolean);
    if (flag === 2) return arr.filter((_, i) => callback(i)); // ARRAY_FILTER_USE_KEY
    if (flag === 1) return arr.filter((v, i) => callback(v, i)); // ARRAY_FILTER_USE_BOTH
    return arr.filter(callback);
  }
  const result = {};
  for (const [k, v] of Object.entries(arr || {})) {
    const keep = !callback ? Boolean(v) : (flag === 2 ? callback(k) : flag === 1 ? callback(v, k) : callback(v));
    if (keep) result[k] = v;
  }
  return result;
}

export function array_reduce(arr, callback, initial = null) {
  const values = Array.isArray(arr) ? arr : Object.values(arr || {});
  return values.reduce(callback, initial);
}

export function array_walk(arr, callback) {
  if (Array.isArray(arr)) arr.forEach((v, i) => callback(v, i));
  else for (const [k, v] of Object.entries(arr || {})) callback(v, k);
  return true;
}

export function array_search(needle, haystack, strict = false) {
  if (Array.isArray(haystack)) {
    const idx = strict ? haystack.findIndex(v => v === needle) : haystack.findIndex(v => v == needle);
    return idx === -1 ? false : idx;
  }
  for (const [k, v] of Object.entries(haystack || {})) {
    if (strict ? v === needle : v == needle) return k;
  }
  return false;
}

export function in_array(needle, haystack, strict = false) {
  if (Array.isArray(haystack)) return strict ? haystack.includes(needle) : haystack.some(v => v == needle);
  return Object.values(haystack || {}).some(v => strict ? v === needle : v == needle);
}

export function array_key_exists(key, arr) {
  if (Array.isArray(arr)) return key >= 0 && key < arr.length;
  return arr != null && key in arr;
}

export function array_unique(arr) {
  if (Array.isArray(arr)) return [...new Set(arr)];
  const seen = new Set();
  const result = {};
  for (const [k, v] of Object.entries(arr || {})) {
    if (!seen.has(v)) { seen.add(v); result[k] = v; }
  }
  return result;
}

export function sort(arr) { if (Array.isArray(arr)) arr.sort((a, b) => a < b ? -1 : a > b ? 1 : 0); return true; }
export function rsort(arr) { if (Array.isArray(arr)) arr.sort((a, b) => a > b ? -1 : a < b ? 1 : 0); return true; }
export function asort(obj) { return true; /* preserve key sorting - no-op for objects */ }
export function arsort(obj) { return true; }
export function ksort(obj) { return true; }
export function krsort(obj) { return true; }
export function usort(arr, fn) { if (Array.isArray(arr)) arr.sort(fn); return true; }
export function uksort(obj, fn) { return true; }
export function uasort(obj, fn) { return true; }

export function array_diff(...arrays) {
  const [first, ...rest] = arrays;
  const restValues = new Set(rest.flatMap(a => Array.isArray(a) ? a : Object.values(a || {})));
  if (Array.isArray(first)) return first.filter(v => !restValues.has(v));
  const result = {};
  for (const [k, v] of Object.entries(first || {})) if (!restValues.has(v)) result[k] = v;
  return result;
}

export function array_intersect(...arrays) {
  const [first, ...rest] = arrays;
  const restValues = rest.map(a => new Set(Array.isArray(a) ? a : Object.values(a || {})));
  if (Array.isArray(first)) return first.filter(v => restValues.every(s => s.has(v)));
  const result = {};
  for (const [k, v] of Object.entries(first || {})) if (restValues.every(s => s.has(v))) result[k] = v;
  return result;
}

export function range(start, end, step = 1) {
  const result = [];
  if (typeof start === 'string' && typeof end === 'string') {
    const s = start.charCodeAt(0), e = end.charCodeAt(0);
    for (let i = s; s <= e ? i <= e : i >= e; i += s <= e ? step : -step) result.push(String.fromCharCode(i));
  } else {
    start = Number(start); end = Number(end);
    for (let i = start; start <= end ? i <= end : i >= end; i += start <= end ? step : -step) result.push(i);
  }
  return result;
}

export function compact(...names) {
  // This needs access to the calling scope - simplified version
  console.warn('compact() has limited support in transpiled code');
  return {};
}

export function extract(obj) {
  console.warn('extract() has limited support in transpiled code');
  return 0;
}

export function list(...vars) {
  console.warn('list() has limited support in transpiled code');
}

export function array_combine(keys, values) {
  const result = {};
  const k = Array.isArray(keys) ? keys : Object.values(keys || {});
  const v = Array.isArray(values) ? values : Object.values(values || {});
  for (let i = 0; i < k.length; i++) result[k[i]] = v[i] ?? null;
  return result;
}

export function array_chunk(arr, size, preserveKeys = false) {
  const values = Array.isArray(arr) ? arr : Object.values(arr || {});
  const result = [];
  for (let i = 0; i < values.length; i += size) result.push(values.slice(i, i + size));
  return result;
}

export function array_fill(startIndex, num, value) {
  const result = [];
  for (let i = 0; i < num; i++) result[startIndex + i] = value;
  return result;
}

export function array_pad(arr, size, value) {
  const a = Array.isArray(arr) ? [...arr] : Object.values(arr || {});
  while (a.length < Math.abs(size)) {
    if (size > 0) a.push(value); else a.unshift(value);
  }
  return a;
}
