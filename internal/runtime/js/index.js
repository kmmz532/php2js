// PHP Runtime for Cloudflare Workers
// Core module - manages output buffering, superglobals, headers

let _output = '';
let _headers = new Headers();
let _statusCode = 200;
let _env = null;
let _request = null;
let _ctx = null;
let _r2 = null;
let _d1 = null;

// Output buffer stack
let _obStack = [];
let _obActive = false;
let _callArgsStack = [];

// Superglobals
export const GLOBALS = {
  _LANG: { encode_hint: {}, skin: {} }
};
export const statics = {};
export const superglobals = {
  _GET: {},
  _POST: {},
  _SERVER: {},
  _COOKIE: {},
  _SESSION: {},
  _REQUEST: {},
  _FILES: {},
  _ENV: {},
  GLOBALS: GLOBALS
};

// Constants
export const CONST_PHP_EOL = '\n';
export const CONST_PHP_INT_MAX = Number.MAX_SAFE_INTEGER;
export const CONST_PHP_INT_MIN = Number.MIN_SAFE_INTEGER;
export const CONST_PHP_INT_SIZE = 8;
export const CONST_PHP_FLOAT_MAX = Number.MAX_VALUE;
export const CONST_PHP_FLOAT_MIN = Number.MIN_VALUE;
export const CONST_PHP_VERSION = '8.1.0';
export const CONST_TRUE = true;
export const CONST_FALSE = false;
export const CONST_NULL = null;
export const CONST_DIRECTORY_SEPARATOR = '/';
export const CONST_PATH_SEPARATOR = ':';
export const CONST_SORT_ASC = 4;
export const CONST_SORT_DESC = 3;
export const CONST_SORT_REGULAR = 0;
export const CONST_SORT_NUMERIC = 1;
export const CONST_SORT_STRING = 2;
export const CONST_ARRAY_FILTER_USE_BOTH = 1;
export const CONST_ARRAY_FILTER_USE_KEY = 2;
export const CONST_ENT_QUOTES = 3;
export const CONST_ENT_HTML5 = 48;
export const CONST_STR_PAD_RIGHT = 1;
export const CONST_STR_PAD_LEFT = 0;
export const CONST_STR_PAD_BOTH = 2;
export const CONST_PREG_SPLIT_NO_EMPTY = 1;
export const CONST_PREG_SET_ORDER = 2;
export const CONST_FILE_APPEND = 8;
export const CONST_LOCK_SH = 1;
export const CONST_LOCK_EX = 2;
export const CONST_LOCK_UN = 3;
export const CONST_LOCK_NB = 4;
export const CONST_SEEK_SET = 0;
export const CONST_SEEK_CUR = 1;
export const CONST_SEEK_END = 2;
export const CONST_JSON_UNESCAPED_UNICODE = 256;
export const CONST_JSON_UNESCAPED_SLASHES = 64;

// User-defined constants and some built-ins
const _constants = {
  'E_ERROR': 1,
  'E_WARNING': 2,
  'E_PARSE': 4,
  'E_NOTICE': 8,
  'E_ALL': 32767,
  'PHP_VERSION': '8.1.0',
  'PHP_EOL': '\n',
  'PREG_SPLIT_NO_EMPTY': 1,
  'PREG_SET_ORDER': 2,
  'FILE_APPEND': 8,
  'LOCK_SH': 1,
  'LOCK_EX': 2,
  'LOCK_UN': 3,
  'LOCK_NB': 4,
  'SEEK_SET': 0,
  'SEEK_CUR': 1,
  'SEEK_END': 2,
  'JSON_UNESCAPED_UNICODE': 256,
  'JSON_UNESCAPED_SLASHES': 64,
  'SORT_ASC': 4,
  'SORT_DESC': 3,
  'SORT_REGULAR': 0,
  'SORT_NUMERIC': 1,
  'SORT_STRING': 2,
  'ENT_QUOTES': 3,
  'ENT_HTML5': 48,
  'ARRAY_FILTER_USE_BOTH': 1,
  'ARRAY_FILTER_USE_KEY': 2,
  'AUTH_TYPE_NONE': 0,
  'AUTH_TYPE_BASIC': 1,
  'AUTH_TYPE_FORM': 2,
  'AUTH_TYPE_EXTERNAL': 3,
  'AUTH_TYPE_SAML': 4,
  'BACKUP_EXT': '.gz',
};

export function init(config) {
  _env = config.env;
  _request = config.request;
  _ctx = config.ctx;
  _r2 = config.r2;
  _d1 = config.d1;
}

export function reset() {
  _output = '';
  _headers = new Headers();
  _statusCode = 200;
  _obStack = [];
  _obActive = false;
  _callArgsStack = [];
  includedFiles.clear();
  superglobals._GET = {};
  superglobals._POST = {};
  superglobals._REQUEST = {};
  superglobals._COOKIE = {};
  superglobals._FILES = {};
  superglobals._SERVER = {};
}

export function getOutput() { return _output; }
export function getStatusCode() { return _statusCode; }
export function getHeaders() { return _headers; }
export function getEnv() { return _env; }
export function getR2() { return _r2; }
export function getD1() { return _d1; }

// --- Output functions ---
export function echo(value) {
  const str = String(value ?? '');
  if (_obActive && _obStack.length > 0) {
    _obStack[_obStack.length - 1] += str;
  } else {
    _output += str;
  }
}

export function print(value) {
  echo(value);
  return 1;
}

export function printf(format, ...args) {
  echo(sprintf(format, ...args));
}

// Output buffering
export function ob_start() { _obStack.push(''); _obActive = true; return true; }
export function ob_end_clean() { if (_obStack.length) { _obStack.pop(); _obActive = _obStack.length > 0; return true; } return false; }
export function ob_end_flush() { if (_obStack.length) { const buf = _obStack.pop(); _obActive = _obStack.length > 0; echo(buf); return true; } return false; }
export function ob_get_contents() { return _obStack.length ? _obStack[_obStack.length - 1] : false; }
export function ob_get_clean() { const c = ob_get_contents(); ob_end_clean(); return c; }
export function ob_get_length() { return _obStack.length ? _obStack[_obStack.length - 1].length : false; }

export function var_dump(...args) {
  for (const arg of args) echo(_varDumpStr(arg) + '\n');
}

function _varDumpStr(v, indent = 0) {
  const pad = '  '.repeat(indent);
  if (v === null || v === undefined) return `${pad}NULL`;
  if (typeof v === 'boolean') return `${pad}bool(${v})`;
  if (typeof v === 'number') {
    if (Number.isInteger(v)) return `${pad}int(${v})`;
    return `${pad}float(${v})`;
  }
  if (typeof v === 'string') return `${pad}string(${v.length}) "${v}"`;
  if (Array.isArray(v)) {
    let s = `${pad}array(${v.length}) {\n`;
    v.forEach((val, i) => { s += `${pad}  [${i}]=>\n${_varDumpStr(val, indent + 1)}\n`; });
    return s + `${pad}}`;
  }
  if (typeof v === 'object') {
    const keys = Object.keys(v);
    let s = `${pad}array(${keys.length}) {\n`;
    for (const k of keys) { s += `${pad}  ["${k}"]=>\n${_varDumpStr(v[k], indent + 1)}\n`; }
    return s + `${pad}}`;
  }
  return `${pad}${String(v)}`;
}

export function print_r(value, returnStr = false) {
  const str = _printRStr(value);
  if (returnStr) return str;
  echo(str);
}

function _printRStr(v, indent = 0) {
  const pad = '    '.repeat(indent);
  if (v === null || v === undefined) return '';
  if (typeof v !== 'object') return String(v);
  let s = 'Array\n' + pad + '(\n';
  const entries = Array.isArray(v) ? v.map((val, i) => [i, val]) : Object.entries(v);
  for (const [k, val] of entries) {
    s += `${pad}    [${k}] => ${typeof val === 'object' && val !== null ? _printRStr(val, indent + 1) : String(val)}\n`;
  }
  return s + pad + ')\n';
}

export function var_export(value, returnStr = false) {
  const str = JSON.stringify(value, null, 2);
  if (returnStr) return str;
  echo(str);
}

// --- Header / HTTP functions ---
export function header(str, replace = true) {
  const match = str.match(/^HTTP\/[\d.]+ (\d+)/);
  if (match) { _statusCode = parseInt(match[1]); return; }
  const colonIdx = str.indexOf(':');
  if (colonIdx !== -1) {
    const name = str.slice(0, colonIdx).trim();
    const value = str.slice(colonIdx + 1).trim();
    if (replace) _headers.set(name, value);
    else _headers.append(name, value);
  }
}

export function headers_sent() { return false; }

export function setcookie(name, value = '', options = {}) {
  let cookie = `${encodeURIComponent(name)}=${encodeURIComponent(value)}`;
  if (options.expires) cookie += `; Expires=${new Date(options.expires * 1000).toUTCString()}`;
  if (options.path) cookie += `; Path=${options.path}`;
  if (options.domain) cookie += `; Domain=${options.domain}`;
  if (options.secure) cookie += '; Secure';
  if (options.httponly) cookie += '; HttpOnly';
  if (options.samesite) cookie += `; SameSite=${options.samesite}`;
  _headers.append('Set-Cookie', cookie);
  return true;
}

// --- Type checking ---
export function isset(v) { return v !== undefined && v !== null; }
export function empty(v) {
  if (v === undefined || v === null || v === false || v === 0 || v === '' || v === '0') return true;
  if (Array.isArray(v)) return v.length === 0;
  if (typeof v === 'object') return Object.keys(v).length === 0;
  return false;
}
export function is_array(v) { return Array.isArray(v) || (typeof v === 'object' && v !== null && !(v instanceof RegExp)); }
export function is_string(v) { return typeof v === 'string'; }
export function is_int(v) { return typeof v === 'number' && Number.isInteger(v); }
export function is_integer(v) { return is_int(v); }
export function is_float(v) { return typeof v === 'number' && !Number.isInteger(v); }
export function is_bool(v) { return typeof v === 'boolean'; }
export function is_null(v) { return v === null || v === undefined; }
export function is_numeric(v) { return !isNaN(parseFloat(v)) && isFinite(v); }
export function is_object(v) { return typeof v === 'object' && v !== null && !Array.isArray(v); }
export function is_callable(v) { return typeof v === 'function'; }
export function gettype(v) {
  if (v === null || v === undefined) return 'NULL';
  if (typeof v === 'boolean') return 'boolean';
  if (typeof v === 'number') return Number.isInteger(v) ? 'integer' : 'double';
  if (typeof v === 'string') return 'string';
  if (Array.isArray(v)) return 'array';
  return 'object';
}
export function settype(v, type) { return true; }

export function str_repeat(str, times) { return String(str).repeat(Math.max(0, times)); }
export function ord(str) { return String(str).charCodeAt(0); }
export function chr(ascii) { return String.fromCharCode(ascii); }
export function bin2hex(str) { return Array.from(String(str)).map(c => c.charCodeAt(0).toString(16).padStart(2, '0')).join(''); }
export function hex2bin(hex) { return hex.match(/.{1,2}/g)?.map(byte => String.fromCharCode(parseInt(byte, 16))).join('') || ''; }
export function intval(v, base = 10) { return parseInt(v, base) || 0; }
export function floatval(v) { return parseFloat(v) || 0; }
export function strval(v) { return String(v ?? ''); }
export function boolval(v) { return Boolean(v); }

// --- Constants ---
export function define(name, value) { _constants[name] = value; }
export function defined(name) { return name in _constants; }
export function constant(name) { return _constants[name]; }

// --- Misc ---
export function die(msg = '') { if (msg) echo(msg); throw new Error('__PHP_EXIT__'); }
export const exit = die;

export function _resolveCallable(fn) {
  if (typeof fn === 'function') return fn;
  if (typeof fn === 'string') {
    if (globalThis[fn]) return globalThis[fn];
    // Check runtime exports
    const runtimeFn = _runtimeExports[fn];
    if (runtimeFn) return runtimeFn;
  }
  if (Array.isArray(fn) && fn.length === 2) {
    const obj = typeof fn[0] === 'string' ? globalThis[fn[0]] : fn[0];
    if (obj && typeof obj[fn[1]] === 'function') {
      return obj[fn[1]].bind(obj);
    }
  }
  return fn;
}

// Registry for runtime function lookups by string name
const _runtimeExports = {};
export function _registerRuntimeExport(name, fn) { _runtimeExports[name] = fn; }

export function class_exists(name) { return typeof globalThis[name] === 'function'; }
export function function_exists(name) { return typeof globalThis[name] === 'function' || name in _runtimeExports; }
export function func_num_args() {
  return _callArgsStack.length ? _callArgsStack[_callArgsStack.length - 1].length : 0;
}
export function func_get_args() {
  return _callArgsStack.length ? [..._callArgsStack[_callArgsStack.length - 1]] : [];
}

export async function call_user_func(fn, ...args) {
  fn = _resolveCallable(fn);
  if (typeof fn !== 'function') return undefined;
  _callArgsStack.push(args);
  try {
    return await fn(...args);
  } finally {
    _callArgsStack.pop();
  }
}
export async function call_user_func_array(fn, args) {
  fn = _resolveCallable(fn);
  if (typeof fn !== 'function') return undefined;
  const callArgs = Array.isArray(args) ? args : Object.values(args || {});
  _callArgsStack.push(callArgs);
  try {
    return await fn(...callArgs);
  } finally {
    _callArgsStack.pop();
  }
}

// --- Environment / Error ---
export function version_compare(v1, v2, op) {
  const parts1 = String(v1).split('.').map(Number);
  const parts2 = String(v2).split('.').map(Number);
  for (let i = 0; i < Math.max(parts1.length, parts2.length); i++) {
    const a = parts1[i] || 0, b = parts2[i] || 0;
    if (a !== b) {
      if (!op) return a > b ? 1 : -1;
      switch (op) {
        case '>': case 'gt': return a > b;
        case '>=': case 'ge': return a >= b;
        case '<': case 'lt': return a < b;
        case '<=': case 'le': return a <= b;
        case '==': case 'eq': return false;
        case '!=': case 'ne': return true;
      }
    }
  }
  if (!op) return 0;
  return op === '==' || op === 'eq' || op === '>=' || op === 'ge' || op === '<=' || op === 'le';
}
export function error_reporting(level) { return 0; }
export function ini_set(key, value) { return false; }
export function ini_get(key) { return ''; }
export function set_time_limit(seconds) { return true; }
export function memory_get_usage() { return 1024 * 1024; }
export function extension_loaded(name) { return name === 'mbstring'; }
export function get_magic_quotes_gpc() { return false; }
export function error_log(msg, type, dest) { console.error(msg); }
export function trigger_error(msg, type) { console.warn(msg); }

// --- File I/O stubs (overridden by file.js) ---
// These are kept as sync stubs; the async versions in file.js take precedence
export function is_readable(path) { return true; }
export function is_writable(path) { return true; }
export function chmod(path, mode) { return true; }
export function chown(path, user) { return true; }
export function clearstatcache() { }

// --- MBString ---
export function mb_language(lang) { return true; }
export function mb_regex_encoding(enc) { return true; }
export function mb_convert_kana(str, option) { return str; }
export function mb_ereg(pattern, string, regs) { return false; }
export function mb_ereg_replace(pattern, replacement, string) { return string; }
export function mb_http_output(enc) { return 'pass'; }
export function mb_detect_order(order) { return true; }
export function mb_convert_variables(to_encoding, from_encoding, vars) { return from_encoding; }

export function sleep(seconds) { /* no-op in workers */ return 0; }
export function usleep(microseconds) { return 0; }

// --- Helpers for foreach ---
export function entries(v) {
  if (Array.isArray(v)) return v.entries();
  if (v && typeof v === 'object') return Object.entries(v);
  return [][Symbol.iterator]();
}
export function values(v) {
  if (Array.isArray(v)) return v;
  if (v && typeof v === 'object') return Object.values(v);
  return [];
}

// --- Array helper ---
export function array(obj) { return obj; }
export function toArray(v) {
  if (Array.isArray(v)) return v;
  if (typeof v === 'object' && v !== null) return Object.values(v);
  return [v];
}

// --- Include/Require ---
const includedFiles = new Set();

function normalize(path) {
  let p = String(path);
  if (p.startsWith('./')) p = p.slice(2);
  if (p.startsWith('/')) p = p.slice(1);
  return p;
}

export async function include(path) {
  let phpPath = normalize(path);
  try {
    const registry = await import('../transpiled/registry.js');
    const module = await registry.default(phpPath);
    if (module && typeof module.default === 'function') {
      await module.default();
    }
    return module;
  } catch (e) {
    if (e.message === '__PHP_EXIT__') throw e;
    console.error("include error", phpPath, e.stack);
    return null;
  }
}

export async function include_once(path) {
  const p = normalize(path);
  if (includedFiles.has(p)) return true;
  const result = await include(p);
  if (result !== null) includedFiles.add(p);
  return result;
}

export async function require_once(path) {
  const p = normalize(path);
  if (includedFiles.has(p)) return true;
  const result = await include(p);
  if (result !== null) includedFiles.add(p);
  return result;
}

export async function require(path) { return require_once(path); }

// --- Object Creation ---
export async function createObject(ClassRef, ...args) {
  if (typeof ClassRef !== 'function') {
    if (typeof globalThis[ClassRef] === 'function') {
      ClassRef = globalThis[ClassRef];
    } else {
      console.warn(`Cannot instantiate: ${ClassRef}, returning stub`);
      return { filter_raw_query_string: (s) => s, get_page_from_query_string: (s) => s, get_page_uri_virtual_query: (p) => '?' + p };
    }
  }
  const obj = new ClassRef();
  if (typeof obj.__construct === 'function') {
    await obj.__construct(...args);
  }
  return obj;
}

// --- Missing PHP functions used by PukiWiki ---
export function pack(format, ...args) {
  if (format === 'H*') {
    const hex = String(args[0]);
    let result = '';
    for (let i = 0; i < hex.length; i += 2) {
      result += String.fromCharCode(parseInt(hex.substr(i, 2), 16));
    }
    return result;
  }
  return args.join('');
}

export function unpack(format, data) { return {}; }

export function uniqid(prefix = '', more_entropy = false) {
  const id = Date.now().toString(16) + Math.random().toString(16).slice(2, 10);
  return (prefix || '') + id;
}

export function preg_quote(str, delimiter) {
  str = String(str ?? '');
  let result = str.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&');
  if (delimiter) {
    result = result.split(delimiter).join('\\' + delimiter);
  }
  return result;
}

export function get_preg_u() { return 'u'; }

export function strnatcmp(a, b) {
  return String(a).localeCompare(String(b), undefined, { numeric: true, sensitivity: 'base' });
}

export function natcasesort(arr) {
  if (Array.isArray(arr)) {
    arr.sort((a, b) => String(a).localeCompare(String(b), undefined, { numeric: true, sensitivity: 'accent' }));
  }
  return true;
}

export function htmlsc(s, flags, encoding, double_encode) {
  return htmlspecialchars(s, flags, encoding, double_encode);
}

export function htmlspecialchars(s, flags, encoding, doubleEncode = true) {
  s = String(s ?? '');
  s = s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  s = s.replace(/"/g, '&quot;').replace(/'/g, '&#039;');
  return s;
}

export function htmlsc_json(v) {
  return htmlspecialchars(JSON.stringify(v));
}

export function compact(...names) { return {}; }
export function extract(obj) { return 0; }

export function ctype_digit(s) {
  return /^\d+$/.test(String(s));
}

export function chop(s) { return String(s ?? '').trimEnd(); }

export function join(glue, pieces) {
  if (pieces === undefined) { pieces = glue; glue = ''; }
  if (Array.isArray(pieces)) return pieces.join(String(glue));
  if (typeof pieces === 'object' && pieces !== null) return Object.values(pieces).join(String(glue));
  return String(pieces);
}

globalThis.join = join;

export function system(cmd) { console.warn('system() not available:', cmd); return false; }
export function popen(cmd, mode) { return false; }
export function pclose(handle) { return 0; }

// flock, rewind, etc. are in file.js but also register as globalThis
export function flock(handle, operation) { return true; }
export function ftruncate(handle, size) { return true; }
export function rewind(handle) { return true; }
export function set_file_buffer(handle, size) { return 0; }
export function fputs(handle, data, length) { return 0; }
globalThis.flock = flock;
globalThis.ftruncate = ftruncate;
globalThis.rewind = rewind;
globalThis.set_file_buffer = set_file_buffer;
globalThis.fputs = fputs;
globalThis.func_num_args = func_num_args;
globalThis.func_get_args = func_get_args;
globalThis.preg_quote = preg_quote;

// Additional missing globals
export function get_html_entity_pattern() { return '[a-zA-Z][a-zA-Z0-9]*'; }
export function pkwk_touch_file(path, time) { return true; }
export function input_filter(v) { return v; }
export function prepare_links_related(page) { }
export function links_get_related(page) { return {}; }
export function prepare_display_materials() { }
export function links_update(page) { }
export function manage_page_redirect() { return false; }
export function ensure_valid_auth_user() { }
export function check_readable(page, flag1, flag2) { return true; }
export function is_page_readable(page) { return true; }
export function get_ticketlink_jira_projects() { return []; }
export function get_auth_external_login_url(page, uri) { return '#'; }
export function get_auth_user_prefix() { return ''; }
export function pkwk_base_uri_type_stack_peek() { return 0; }
export function exist_plugin(name) { return true; }
export function exist_plugin_action(name) { return typeof globalThis[`plugin_${name}_action`] === 'function'; }
export function exist_plugin_convert(name) { return typeof globalThis[`plugin_${name}_convert`] === 'function'; }
export async function do_plugin_action(name) {
  const handler = globalThis[`plugin_${name}_action`];
  if (typeof handler !== 'function') return false;
  return await handler();
}
export async function do_plugin_convert(name) {
  const handler = globalThis[`plugin_${name}_convert`];
  if (typeof handler !== 'function') return '';
  return await handler();
}
export function attach_filelist() { return ''; }
export function make_link(str) { return str; }
export function convert_html(source) { 
  if (Array.isArray(source)) return source.join('');
  return String(source ?? '');
}
export function guess_script_absolute_uri() {
  const s = superglobals._SERVER;
  const scheme = s['HTTPS'] === 'on' ? 'https' : 'http';
  const host = s['HTTP_HOST'] || s['SERVER_NAME'] || 'localhost';
  const script = s['SCRIPT_NAME'] || '/';
  return `${scheme}://${host}${script}`;
}

// Re-export sub-modules
export * from './string.js';
export * from './array.js';
export * from './math.js';
export * from './file.js';
export * from './db.js';
export * from './date.js';
export * from './hash.js';
export * from './regex.js';
export * from './url.js';
