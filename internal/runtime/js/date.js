// PHP Date/Time functions
export function time() { return Math.floor(Date.now() / 1000); }
export function microtime(asFloat = false) {
  const now = Date.now() / 1000;
  if (asFloat) return now;
  const sec = Math.floor(now);
  const usec = (now - sec).toFixed(8);
  return `${usec} ${sec}`;
}

export function date(format, timestamp) {
  const d = timestamp ? new Date(timestamp * 1000) : new Date();
  return _formatDate(d, format);
}

export function gmdate(format, timestamp) {
  const d = timestamp ? new Date(timestamp * 1000) : new Date();
  return _formatDate(d, format, true);
}

export function mktime(h = 0, m = 0, s = 0, month = 1, day = 1, year = 1970) {
  return Math.floor(new Date(year, month - 1, day, h, m, s).getTime() / 1000);
}

export function strtotime(str, now) {
  const base = now ? new Date(now * 1000) : new Date();
  const parsed = Date.parse(str);
  if (!isNaN(parsed)) return Math.floor(parsed / 1000);
  // Handle relative dates
  const rel = str.match(/^([+-]?\d+)\s+(second|minute|hour|day|week|month|year)s?$/i);
  if (rel) {
    const n = parseInt(rel[1]);
    const unit = rel[2].toLowerCase();
    const d = new Date(base);
    switch (unit) {
      case 'second': d.setSeconds(d.getSeconds() + n); break;
      case 'minute': d.setMinutes(d.getMinutes() + n); break;
      case 'hour': d.setHours(d.getHours() + n); break;
      case 'day': d.setDate(d.getDate() + n); break;
      case 'week': d.setDate(d.getDate() + n * 7); break;
      case 'month': d.setMonth(d.getMonth() + n); break;
      case 'year': d.setFullYear(d.getFullYear() + n); break;
    }
    return Math.floor(d.getTime() / 1000);
  }
  return false;
}

function _formatDate(d, format, utc = false) {
  const g = utc ? 'getUTC' : 'get';
  const map = {
    'Y': () => d[g + 'FullYear'](),
    'y': () => String(d[g + 'FullYear']()).slice(-2),
    'm': () => String(d[g + 'Month']() + 1).padStart(2, '0'),
    'n': () => d[g + 'Month']() + 1,
    'd': () => String(d[g + 'Date']()).padStart(2, '0'),
    'j': () => d[g + 'Date'](),
    'H': () => String(d[g + 'Hours']()).padStart(2, '0'),
    'G': () => d[g + 'Hours'](),
    'i': () => String(d[g + 'Minutes']()).padStart(2, '0'),
    's': () => String(d[g + 'Seconds']()).padStart(2, '0'),
    'A': () => d[g + 'Hours']() >= 12 ? 'PM' : 'AM',
    'a': () => d[g + 'Hours']() >= 12 ? 'pm' : 'am',
    'U': () => Math.floor(d.getTime() / 1000),
    'N': () => d[g + 'Day']() || 7,
    'w': () => d[g + 'Day'](),
    'g': () => d[g + 'Hours']() % 12 || 12,
    'h': () => String(d[g + 'Hours']() % 12 || 12).padStart(2, '0'),
    'D': () => ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][d[g + 'Day']()],
    'l': () => ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'][d[g + 'Day']()],
    'M': () => ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'][d[g + 'Month']()],
    'F': () => ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'][d[g + 'Month']()],
    't': () => new Date(d[g + 'FullYear'](), d[g + 'Month']() + 1, 0).getDate(),
    'W': () => { const target = new Date(d); target.setDate(target.getDate() - ((d.getDay() + 6) % 7) + 3); const jan4 = new Date(target.getFullYear(), 0, 4); return String(1 + Math.round((target - jan4) / 86400000 / 7)).padStart(2, '0'); },
  };
  let result = '';
  for (let i = 0; i < format.length; i++) {
    if (format[i] === '\\' && i + 1 < format.length) { result += format[++i]; continue; }
    result += map[format[i]] ? map[format[i]]() : format[i];
  }
  return result;
}
