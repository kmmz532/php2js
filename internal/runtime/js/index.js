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

// Superglobals
export const GLOBALS = {
  _LANG: { encode_hint: {}, skin: {} }
};
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

// User-defined constants and some built-ins
const _constants = {
  'E_ERROR': 1,
  'E_WARNING': 2,
  'E_PARSE': 4,
  'E_NOTICE': 8,
  'E_ALL': 32767
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
}

export function getOutput() { return _output; }
export function getStatusCode() { return _statusCode; }
export function getHeaders() { return _headers; }
export function getEnv() { return _env; }
export function getR2() { return _r2; }
export function getD1() { return _d1; }

// --- Output functions ---
export function echo(value) {
  _output += String(value ?? '');
}

export function print(value) {
  _output += String(value ?? '');
  return 1;
}

export function printf(format, ...args) {
  _output += sprintf(format, ...args);
}

export function var_dump(...args) {
  for (const arg of args) {
    _output += _varDumpStr(arg) + '\n';
  }
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
  _output += str;
}

function _printRStr(v, indent = 0) {
  const pad = '    '.repeat(indent);
  if (v === null || v === undefined) return '';
  if (typeof v !== 'object') return String(v);
  const isArr = Array.isArray(v);
  let s = 'Array\n' + pad + '(\n';
  const entries = isArr ? v.map((val, i) => [i, val]) : Object.entries(v);
  for (const [k, val] of entries) {
    s += `${pad}    [${k}] => ${typeof val === 'object' && val !== null ? _printRStr(val, indent + 1) : String(val)}\n`;
  }
  return s + pad + ')\n';
}

export function var_export(value, returnStr = false) {
  const str = JSON.stringify(value, null, 2);
  if (returnStr) return str;
  _output += str;
}

// --- Header / HTTP functions ---
export function header(str, replace = true) {
  const match = str.match(/^HTTP\/[\d.]+ (\d+)/);
  if (match) {
    _statusCode = parseInt(match[1]);
    return;
  }
  const colonIdx = str.indexOf(':');
  if (colonIdx !== -1) {
    const name = str.slice(0, colonIdx).trim();
    const value = str.slice(colonIdx + 1).trim();
    if (replace) {
      _headers.set(name, value);
    } else {
      _headers.append(name, value);
    }
  }
}

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
export function is_array(v) { return Array.isArray(v) || (typeof v === 'object' && v !== null); }
export function is_string(v) { return typeof v === 'string'; }
export function is_int(v) { return typeof v === 'number' && Number.isInteger(v); }
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
export function intval(v, base = 10) { return parseInt(v, base) || 0; }
export function floatval(v) { return parseFloat(v) || 0; }
export function strval(v) { return String(v ?? ''); }

// --- Constants ---
export function define(name, value) { _constants[name] = value; }
export function defined(name) { return name in _constants; }
export function constant(name) { return _constants[name]; }

// --- Misc ---
export function die(msg = '') { if (msg) echo(msg); throw new Error('__PHP_EXIT__'); }
export const exit = die;
export function _resolveCallable(fn) {
  if (typeof fn === 'string') return globalThis[fn];
  if (Array.isArray(fn) && fn.length === 2) {
    const obj = typeof fn[0] === 'string' ? globalThis[fn[0]] : fn[0];
    if (obj && typeof obj[fn[1]] === 'function') {
      return obj[fn[1]].bind(obj);
    }
  }
  return fn;
}

export function class_exists(name) { return typeof globalThis[name] === 'function'; }
export function function_exists(name) { return typeof globalThis[name] === 'function'; }
export function call_user_func(fn, ...args) { 
  fn = _resolveCallable(fn);
  return typeof fn === 'function' ? fn(...args) : undefined; 
}
export function call_user_func_array(fn, args) { 
  fn = _resolveCallable(fn);
  return typeof fn === 'function' ? fn(...args) : undefined; 
}

// --- Environment / Error ---
export function version_compare(v1, v2, op) { return false; }
export function error_reporting(level) { return 0; }
export function ini_set(key, value) { return false; }
export function ini_get(key) { return ''; }
export function set_time_limit(seconds) { return true; }
export function memory_get_usage() { return 1024 * 1024; }
export function extension_loaded(name) { return true; }
export function get_magic_quotes_gpc() { return false; }

// --- File I/O ---
export async function file_exists(path) { return true; }
export function is_readable(path) { return true; }
export function is_writable(path) { return true; }

// --- MBString ---
export function mb_language(lang) { return true; }
export function mb_regex_encoding(enc) { return true; }
export function mb_convert_kana(str, option) { return str; }
export function mb_ereg(pattern, string, regs) { return false; }
export function mb_ereg_replace(pattern, replacement, string) { return string; }
export function mb_http_output(enc) { return 'pass'; }
export function mb_detect_order(order) { return true; }
export function mb_convert_variables(to_encoding, from_encoding, vars) { return from_encoding; }

export function sleep(seconds) { return new Promise(r => setTimeout(r, seconds * 1000)); }
export function usleep(microseconds) { return new Promise(r => setTimeout(r, microseconds / 1000)); }

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
export function array(obj) {
  // Creates a PHP-style associative array (plain object)
  return obj;
}

export function toArray(v) {
  if (Array.isArray(v)) return v;
  if (typeof v === 'object' && v !== null) return Object.values(v);
  return [v];
}

export async function include(path) {
  try {
    let phpPath = String(path);
    if (phpPath.startsWith('./')) phpPath = phpPath.slice(2);
    if (phpPath.startsWith('/')) phpPath = phpPath.slice(1);
    
    // Import the registry which has static imports mapped
    const registry = await import('../transpiled/registry.js');
    const module = await registry.default(phpPath);
    if (!module) {
      console.warn(`[runtime] Module not found in registry: ${phpPath}`);
    } else if (typeof module.default === 'function') {
      await module.default();
    }
    return module;
  } catch (e) {
    console.warn(`[runtime] Dynamic include failed for ${path}: ${e.message}`);
    return null;
  }
}
export async function include_once(path) { return include(path); }
export async function require(path) { return include(path); }
export async function require_once(path) { return include(path); }

// --- Object Creation ---
export async function createObject(ClassRef, ...args) {
  if (typeof ClassRef !== 'function') {
    if (typeof globalThis[ClassRef] === 'function') {
      ClassRef = globalThis[ClassRef];
    } else {
      throw new Error(`Cannot instantiate non-class: ${ClassRef}`);
    }
  }
  const obj = new ClassRef();
  if (typeof obj.__construct === 'function') {
    await obj.__construct(...args);
  }
  return obj;
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
