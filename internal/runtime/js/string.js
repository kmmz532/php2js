// PHP String functions for JS runtime

export function strlen(s) { return s == null ? 0 : String(s).length; }
export function mb_strlen(s) { return s == null ? 0 : [...String(s)].length; }

export function strpos(haystack, needle, offset = 0) {
  const idx = String(haystack).indexOf(String(needle), offset);
  return idx === -1 ? false : idx;
}
export function mb_strpos(h, n, o = 0) { return strpos(h, n, o); }

export function strrpos(haystack, needle, offset = 0) {
  const idx = String(haystack).lastIndexOf(String(needle), offset || undefined);
  return idx === -1 ? false : idx;
}

export function substr(s, start, length) {
  s = String(s ?? '');
  if (start < 0) start = Math.max(0, s.length + start);
  if (length === undefined) return s.slice(start);
  if (length < 0) return s.slice(start, s.length + length);
  return s.slice(start, start + length);
}
export function mb_substr(s, start, length) { return substr(s, start, length); }

export function str_replace(search, replace, subject) {
  if (Array.isArray(search)) {
    let result = String(subject);
    for (let i = 0; i < search.length; i++) {
      const rep = Array.isArray(replace) ? (replace[i] ?? '') : replace;
      result = result.split(String(search[i])).join(String(rep));
    }
    return result;
  }
  return String(subject).split(String(search)).join(String(replace));
}

export function explode(delimiter, string, limit) {
  const s = String(string ?? '');
  if (limit === undefined) return s.split(String(delimiter));
  const parts = s.split(String(delimiter));
  if (limit > 0) {
    if (parts.length <= limit) return parts;
    const result = parts.slice(0, limit - 1);
    result.push(parts.slice(limit - 1).join(delimiter));
    return result;
  }
  return parts;
}

export function implode(glue, pieces) {
  if (pieces === undefined) { pieces = glue; glue = ''; }
  if (Array.isArray(pieces)) return pieces.join(String(glue));
  if (typeof pieces === 'object') return Object.values(pieces).join(String(glue));
  return String(pieces);
}

export function trim(s, chars) {
  s = String(s ?? '');
  if (!chars) return s.trim();
  const escaped = chars.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&');
  return s.replace(new RegExp(`^[${escaped}]+|[${escaped}]+$`, 'g'), '');
}

export function ltrim(s, chars) {
  s = String(s ?? '');
  if (!chars) return s.trimStart();
  const escaped = chars.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&');
  return s.replace(new RegExp(`^[${escaped}]+`, 'g'), '');
}

export function rtrim(s, chars) {
  s = String(s ?? '');
  if (!chars) return s.trimEnd();
  const escaped = chars.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&');
  return s.replace(new RegExp(`[${escaped}]+$`, 'g'), '');
}

export function strtolower(s) { return String(s ?? '').toLowerCase(); }
export function strtoupper(s) { return String(s ?? '').toUpperCase(); }
export function mb_strtolower(s) { return strtolower(s); }
export function mb_strtoupper(s) { return strtoupper(s); }
export function ucfirst(s) { s = String(s ?? ''); return s.charAt(0).toUpperCase() + s.slice(1); }
export function lcfirst(s) { s = String(s ?? ''); return s.charAt(0).toLowerCase() + s.slice(1); }

export function sprintf(format, ...args) {
  let i = 0;
  return String(format ?? '').replace(/%([+\-0 ]*)(\d+)?(?:\.(\d+))?([sdfeEgGxXobc%])/g,
    (match, flags, width, precision, type) => {
      if (type === '%') return '%';
      const arg = args[i++];
      let result;
      switch (type) {
        case 's': result = String(arg ?? ''); break;
        case 'd': result = String(parseInt(arg) || 0); break;
        case 'f': result = (parseFloat(arg) || 0).toFixed(precision !== undefined ? parseInt(precision) : 6); break;
        case 'e': case 'E': result = (parseFloat(arg) || 0).toExponential(precision !== undefined ? parseInt(precision) : 6); break;
        case 'x': result = (parseInt(arg) || 0).toString(16); break;
        case 'X': result = (parseInt(arg) || 0).toString(16).toUpperCase(); break;
        case 'o': result = (parseInt(arg) || 0).toString(8); break;
        case 'b': result = (parseInt(arg) || 0).toString(2); break;
        case 'c': result = String.fromCharCode(parseInt(arg) || 0); break;
        default: result = String(arg ?? '');
      }
      if (width) {
        const w = parseInt(width);
        const padChar = flags.includes('0') ? '0' : ' ';
        if (flags.includes('-')) result = result.padEnd(w, padChar);
        else result = result.padStart(w, padChar);
      }
      return result;
    });
}

export function htmlspecialchars(s, flags, encoding, doubleEncode = true) {
  s = String(s ?? '');
  s = s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  s = s.replace(/"/g, '&quot;').replace(/'/g, '&#039;');
  return s;
}

export function htmlspecialchars_decode(s) {
  return String(s ?? '').replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"').replace(/&#039;/g, "'");
}

export function htmlentities(s) { return htmlspecialchars(s); }

export function strip_tags(s, allowed) {
  s = String(s ?? '');
  if (!allowed) return s.replace(/<[^>]*>/g, '');
  const allowedTags = String(allowed).match(/<\w+>/g) || [];
  const tagPattern = allowedTags.map(t => t.slice(1, -1)).join('|');
  return s.replace(new RegExp(`<(?!\\/?(${tagPattern})\\b)[^>]*>`, 'gi'), '');
}

export function nl2br(s) { return String(s ?? '').replace(/\n/g, '<br />\n'); }
export function wordwrap(s, width = 75, brk = '\n', cutLongWords = false) {
  s = String(s ?? '');
  if (!cutLongWords) {
    const regex = new RegExp(`(.{1,${width}})(\\s|$)`, 'g');
    return s.replace(regex, `$1${brk}`).trim();
  }
  const regex = new RegExp(`.{1,${width}}`, 'g');
  return (s.match(regex) || []).join(brk);
}

export function str_pad(input, length, padStr = ' ', padType = 1) {
  input = String(input ?? '');
  if (input.length >= length) return input;
  if (padType === 0) return input.padStart(length, padStr); // STR_PAD_LEFT
  if (padType === 2) { // STR_PAD_BOTH
    const diff = length - input.length;
    const left = Math.floor(diff / 2);
    return input.padStart(input.length + left, padStr).padEnd(length, padStr);
  }
  return input.padEnd(length, padStr); // STR_PAD_RIGHT
}

export function str_repeat(s, times) { return String(s ?? '').repeat(Math.max(0, times)); }
export function str_split(s, length = 1) {
  s = String(s ?? '');
  const result = [];
  for (let i = 0; i < s.length; i += length) result.push(s.slice(i, i + length));
  return result.length ? result : [''];
}

export function chunk_split(body, chunklen = 76, end = '\r\n') {
  body = String(body ?? '');
  let result = '';
  for (let i = 0; i < body.length; i += chunklen) result += body.slice(i, i + chunklen) + end;
  return result;
}

export function ord(s) { return String(s ?? '').charCodeAt(0) || 0; }
export function chr(code) { return String.fromCharCode(code); }

export function mb_convert_encoding(s, to, from) { return String(s ?? ''); }
export function mb_detect_encoding(s) { return 'UTF-8'; }
export function mb_internal_encoding(enc) { return enc ? true : 'UTF-8'; }

export function number_format(num, decimals = 0, decPoint = '.', thousandsSep = ',') {
  num = parseFloat(num) || 0;
  const fixed = num.toFixed(decimals);
  const [intPart, decPart] = fixed.split('.');
  const formatted = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, thousandsSep);
  return decPart ? formatted + decPoint + decPart : formatted;
}

export function str_contains(haystack, needle) { return String(haystack).includes(String(needle)); }
export function str_starts_with(haystack, needle) { return String(haystack).startsWith(String(needle)); }
export function str_ends_with(haystack, needle) { return String(haystack).endsWith(String(needle)); }
export function substr_count(haystack, needle) { return String(haystack).split(String(needle)).length - 1; }
export function str_word_count(s) { return String(s ?? '').trim().split(/\s+/).filter(Boolean).length; }
export function strtolower_first(s) { return lcfirst(s); }
