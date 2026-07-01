// PHP File I/O functions -> Cloudflare R2
import { getR2 } from './index.js';

export async function file_get_contents(path) {
  const r2 = getR2();
  if (!r2) { console.warn('R2 not available'); return false; }
  const key = _normalizePath(path);
  const obj = await r2.get(key);
  if (!obj) return false;
  return await obj.text();
}

export async function file_put_contents(path, data, flags = 0) {
  const r2 = getR2();
  if (!r2) return false;
  const key = _normalizePath(path);
  await r2.put(key, String(data));
  return String(data).length;
}

export async function file_exists(path) {
  const r2 = getR2();
  if (!r2) return false;
  const key = _normalizePath(path);
  const obj = await r2.head(key);
  return obj !== null;
}

export async function unlink(path) {
  const r2 = getR2();
  if (!r2) return false;
  await r2.delete(_normalizePath(path));
  return true;
}

export async function rename(from, to) {
  const content = await file_get_contents(from);
  if (content === false) return false;
  await file_put_contents(to, content);
  await unlink(from);
  return true;
}

export async function scandir(path) {
  const r2 = getR2();
  if (!r2) return false;
  const prefix = _normalizePath(path);
  const listed = await r2.list({ prefix: prefix + '/' });
  return listed.objects.map(o => o.key.replace(prefix + '/', '').split('/')[0]).filter((v, i, a) => a.indexOf(v) === i);
}

export async function glob(pattern) {
  // Simplified glob using R2 list
  const r2 = getR2();
  if (!r2) return [];
  const prefix = pattern.replace(/\*.*$/, '');
  const listed = await r2.list({ prefix });
  return listed.objects.map(o => o.key);
}

export async function mkdir(path) { return true; /* R2 doesn't need directories */ }
export async function rmdir(path) { return true; }
export async function is_dir(path) { const r2 = getR2(); if(!r2) return false; const l = await r2.list({prefix: _normalizePath(path) + '/', limit: 1}); return l.objects.length > 0; }
export async function is_file(path) { return await file_exists(path); }

export function dirname(path) { const parts = String(path).split('/'); parts.pop(); return parts.join('/') || '.'; }
export function basename(path, suffix) { let base = String(path).split('/').pop() || ''; if(suffix && base.endsWith(suffix)) base = base.slice(0, -suffix.length); return base; }
export function pathinfo(path, option) {
  const s = String(path);
  const dir = dirname(s);
  const base = basename(s);
  const ext = base.includes('.') ? base.split('.').pop() : '';
  const filename = ext ? base.slice(0, -(ext.length + 1)) : base;
  const info = { dirname: dir, basename: base, extension: ext, filename };
  return option !== undefined ? Object.values(info)[option] : info;
}
export function realpath(path) { return _normalizePath(path); }

export async function file(path, flags = 0) {
  const content = await file_get_contents(path);
  if (content === false) return false;
  return content.split('\n');
}

// File handle emulation (simplified)
const _fileHandles = new Map();
let _handleId = 0;

export async function fopen(path, mode) {
  const id = ++_handleId;
  const content = mode.includes('r') ? (await file_get_contents(path) || '') : '';
  _fileHandles.set(id, { path: _normalizePath(path), mode, content, pos: 0, buffer: '' });
  return id;
}

export async function fclose(handle) { const h = _fileHandles.get(handle); if(h && h.mode.includes('w')) await file_put_contents(h.path, h.buffer); _fileHandles.delete(handle); return true; }
export function fread(handle, length) { const h = _fileHandles.get(handle); if(!h) return false; const data = h.content.slice(h.pos, h.pos + length); h.pos += length; return data; }
export function fwrite(handle, data) { const h = _fileHandles.get(handle); if(!h) return false; h.buffer += data; return data.length; }
export function fgets(handle, length) { const h = _fileHandles.get(handle); if(!h) return false; const nl = h.content.indexOf('\n', h.pos); const end = nl === -1 ? h.content.length : nl + 1; const line = h.content.slice(h.pos, Math.min(end, h.pos + (length || Infinity))); h.pos += line.length; return line || false; }
export function feof(handle) { const h = _fileHandles.get(handle); return !h || h.pos >= h.content.length; }

function _normalizePath(path) {
  return String(path).replace(/^\.\//, '').replace(/\/+/g, '/').replace(/^\//, '');
}
