// PHP Math functions
export function abs(n) { return Math.abs(n); }
export function ceil(n) { return Math.ceil(n); }
export function floor(n) { return Math.floor(n); }
export function round(n, p = 0) { const f = 10 ** p; return Math.round(n * f) / f; }
export function max(...args) { const a = args.length === 1 && Array.isArray(args[0]) ? args[0] : args; return Math.max(...a); }
export function min(...args) { const a = args.length === 1 && Array.isArray(args[0]) ? args[0] : args; return Math.min(...a); }
export function pow(base, exp) { return Math.pow(base, exp); }
export function sqrt(n) { return Math.sqrt(n); }
export function rand(min = 0, max = 2147483647) { return Math.floor(Math.random() * (max - min + 1)) + min; }
export const mt_rand = rand;
export function log(n) { return Math.log(n); }
export function log10(n) { return Math.log10(n); }
export function fmod(x, y) { return x % y; }
export function pi() { return Math.PI; }
export function bindec(s) { return parseInt(String(s), 2); }
export function octdec(s) { return parseInt(String(s), 8); }
export function hexdec(s) { return parseInt(String(s), 16); }
export function decbin(n) { return (n >>> 0).toString(2); }
export function decoct(n) { return (n >>> 0).toString(8); }
export function dechex(n) { return (n >>> 0).toString(16); }
