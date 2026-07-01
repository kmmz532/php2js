// PHP URL functions
export function urlencode(s) { return encodeURIComponent(String(s ?? '')).replace(/%20/g, '+'); }
export function urldecode(s) { return decodeURIComponent(String(s ?? '').replace(/\+/g, ' ')); }
export function rawurlencode(s) { return encodeURIComponent(String(s ?? '')); }
export function rawurldecode(s) { return decodeURIComponent(String(s ?? '')); }

export function http_build_query(data, numericPrefix = '', argSep = '&') {
  const params = [];
  const _build = (obj, prefix) => {
    for (const [key, value] of Object.entries(obj)) {
      const k = prefix ? `${prefix}[${key}]` : key;
      if (typeof value === 'object' && value !== null) _build(value, k);
      else params.push(`${encodeURIComponent(k)}=${encodeURIComponent(value ?? '')}`);
    }
  };
  _build(data, '');
  return params.join(argSep);
}

export function parse_url(url, component) {
  try {
    const u = new URL(url, 'http://localhost');
    const result = {
      scheme: u.protocol.replace(':', ''),
      host: u.hostname,
      port: u.port ? parseInt(u.port) : undefined,
      user: u.username || undefined,
      pass: u.password || undefined,
      path: u.pathname,
      query: u.search.slice(1) || undefined,
      fragment: u.hash.slice(1) || undefined,
    };
    if (component !== undefined) return Object.values(result)[component];
    return result;
  } catch { return false; }
}

export function parse_str(str, result) {
  const params = new URLSearchParams(str);
  for (const [key, value] of params.entries()) {
    if (result) result[key] = value;
  }
}

// JSON
export function json_encode(value) { try { return JSON.stringify(value); } catch { return false; } }
export function json_decode(str, assoc = false) {
  try {
    const val = JSON.parse(str);
    return val;
  } catch { return null; }
}

// Session (stub - uses simple object for now)
export async function session_start() { return true; }
export async function session_destroy() { return true; }
export function session_id(id) { return id || 'stub-session-id'; }
