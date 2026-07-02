// PHP File I/O functions -> Cloudflare R2 + Embedded Data Fallback
import { getR2 } from './index.js';

// Embedded data manifest (loaded lazily)
let _dataManifest = null;
async function getManifest() {
  if (_dataManifest === null) {
    try {
      const m = await import('../data-manifest.json');
      _dataManifest = m.default || m;
    } catch {
      _dataManifest = {};
    }
  }
  return _dataManifest;
}

// File metadata cache (simulates stat cache)
const _metaCache = new Map();

export async function file_get_contents(path) {
  const key = _normalizePath(path);
  
  // Try R2 first
  const r2 = getR2();
  if (r2) {
    try {
      const obj = await r2.get(key);
      if (obj) {
        return await obj.text();
      }
    } catch (e) {
      console.warn('R2 get error:', key, e.message);
    }
  }
  
  // Fallback to embedded manifest
  const manifest = await getManifest();
  if (key in manifest) {
    return manifest[key];
  }
  
  return false;
}

export async function file_put_contents(path, data, flags = 0) {
  const r2 = getR2();
  if (!r2) {
    // Write to manifest cache for dev mode
    const manifest = await getManifest();
    const key = _normalizePath(path);
    manifest[key] = String(data);
    return String(data).length;
  }
  const key = _normalizePath(path);
  const content = String(data);
  
  if (flags & 8) { // FILE_APPEND = 8
    const existing = await file_get_contents(path);
    if (existing !== false) {
      await r2.put(key, existing + content);
      return content.length;
    }
  }
  
  await r2.put(key, content);
  _metaCache.set(key, { size: content.length, mtime: Math.floor(Date.now() / 1000) });
  return content.length;
}

export async function file_exists(path) {
  const key = _normalizePath(path);
  
  // Check metadata cache first
  if (_metaCache.has(key)) return true;
  
  // Try R2
  const r2 = getR2();
  if (r2) {
    try {
      const obj = await r2.head(key);
      if (obj) {
        _metaCache.set(key, { size: obj.size, mtime: Math.floor(obj.uploaded?.getTime() / 1000) || Math.floor(Date.now() / 1000) });
        return true;
      }
    } catch (e) {
      // R2 error, fall through
    }
  }
  
  // Fallback to manifest
  const manifest = await getManifest();
  if (key in manifest) {
    return true;
  }
  
  return false;
}

export async function unlink(path) {
  const key = _normalizePath(path);
  _metaCache.delete(key);
  
  const r2 = getR2();
  if (r2) {
    try {
      await r2.delete(key);
    } catch {}
  }
  
  // Also remove from manifest
  const manifest = await getManifest();
  delete manifest[key];
  
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
  const prefix = _normalizePath(path);
  const results = new Set();
  
  // From R2
  const r2 = getR2();
  if (r2) {
    try {
      const pfx = prefix.endsWith('/') ? prefix : prefix + '/';
      const listed = await r2.list({ prefix: pfx, limit: 1000 });
      for (const obj of listed.objects) {
        const name = obj.key.replace(pfx, '').split('/')[0];
        if (name) results.add(name);
      }
    } catch {}
  }
  
  // From manifest
  const manifest = await getManifest();
  const pfx = prefix.endsWith('/') ? prefix : prefix + '/';
  for (const key of Object.keys(manifest)) {
    if (key.startsWith(pfx)) {
      const name = key.replace(pfx, '').split('/')[0];
      if (name) results.add(name);
    }
  }
  
  return [...results].sort();
}

export async function glob(pattern) {
  const r2 = getR2();
  const results = [];
  
  // Convert glob to prefix
  const prefix = pattern.replace(/\*.*$/, '');
  
  if (r2) {
    try {
      const listed = await r2.list({ prefix, limit: 1000 });
      results.push(...listed.objects.map(o => o.key));
    } catch {}
  }
  
  // From manifest
  const manifest = await getManifest();
  for (const key of Object.keys(manifest)) {
    if (key.startsWith(prefix) && !results.includes(key)) {
      results.push(key);
    }
  }
  
  return results;
}

export async function mkdir(path) { return true; /* R2 doesn't need directories */ }
export async function rmdir(path) { return true; }

export async function is_dir(path) {
  const prefix = _normalizePath(path);
  const pfx = prefix.endsWith('/') ? prefix : prefix + '/';
  
  // Check R2
  const r2 = getR2();
  if (r2) {
    try {
      const l = await r2.list({ prefix: pfx, limit: 1 });
      if (l.objects.length > 0) return true;
    } catch {}
  }
  
  // Check manifest
  const manifest = await getManifest();
  for (const key of Object.keys(manifest)) {
    if (key.startsWith(pfx)) return true;
  }
  
  return false;
}

export async function is_file(path) { return await file_exists(path); }

export function dirname(path) {
  const parts = String(path).split('/');
  parts.pop();
  return parts.join('/') || '.';
}

export function basename(path, suffix) {
  let base = String(path).split('/').pop() || '';
  if (suffix && base.endsWith(suffix)) base = base.slice(0, -suffix.length);
  return base;
}

export function pathinfo(path, option) {
  const s = String(path);
  const dir = dirname(s);
  const base = basename(s);
  const ext = base.includes('.') ? base.split('.').pop() : '';
  const filename = ext ? base.slice(0, -(ext.length + 1)) : base;
  const info = { dirname: dir, basename: base, extension: ext, filename };
  if (option !== undefined) {
    const vals = [info.dirname, info.basename, info.extension, info.filename];
    return vals[option] ?? info;
  }
  return info;
}

export function realpath(path) { return _normalizePath(path); }

export async function file(path, flags = 0) {
  const content = await file_get_contents(path);
  if (content === false) return false;
  return content.split('\n').map((line, i, arr) => 
    i < arr.length - 1 ? line + '\n' : (line ? line + '\n' : '')
  ).filter(l => l !== '');
}

// Filesize using metadata cache or head request
export async function filesize(path) {
  const key = _normalizePath(path);
  
  if (_metaCache.has(key)) return _metaCache.get(key).size || 0;
  
  // Check manifest
  const manifest = await getManifest();
  if (key in manifest) return manifest[key].length;
  
  // Check R2
  const r2 = getR2();
  if (r2) {
    try {
      const obj = await r2.head(key);
      if (obj) return obj.size;
    } catch {}
  }
  
  return 0;
}

export async function filemtime(path) {
  const key = _normalizePath(path);
  
  if (_metaCache.has(key)) return _metaCache.get(key).mtime || Math.floor(Date.now() / 1000);
  
  // Check R2
  const r2 = getR2();
  if (r2) {
    try {
      const obj = await r2.head(key);
      if (obj && obj.uploaded) return Math.floor(obj.uploaded.getTime() / 1000);
    } catch {}
  }
  
  return Math.floor(Date.now() / 1000);
}

export function filectime(path) { return filemtime(path); }
export function fileatime(path) { return filemtime(path); }

// File handle emulation (simplified for Workers)
const _fileHandles = new Map();
let _handleId = 0;

export async function fopen(path, mode) {
  const id = ++_handleId;
  let content = '';
  if (mode.includes('r') || mode.includes('a')) {
    content = await file_get_contents(path) || '';
  }
  _fileHandles.set(id, { 
    path: _normalizePath(path), 
    mode, 
    content: String(content), 
    pos: 0, 
    buffer: mode.includes('a') ? String(content) : '',
    dirty: false
  });
  return id;
}

export async function fclose(handle) {
  const h = _fileHandles.get(handle);
  if (h && (h.mode.includes('w') || h.mode.includes('a') || h.dirty)) {
    await file_put_contents(h.path, h.buffer);
  }
  _fileHandles.delete(handle);
  return true;
}

export function fread(handle, length) {
  const h = _fileHandles.get(handle);
  if (!h) return false;
  const data = h.content.slice(h.pos, h.pos + length);
  h.pos += length;
  return data;
}

export function fwrite(handle, data, length) {
  const h = _fileHandles.get(handle);
  if (!h) return false;
  const str = length !== undefined ? String(data).slice(0, length) : String(data);
  h.buffer += str;
  h.dirty = true;
  return str.length;
}

export function fgets(handle, length) {
  const h = _fileHandles.get(handle);
  if (!h) return false;
  if (h.pos >= h.content.length) return false;
  const nl = h.content.indexOf('\n', h.pos);
  const end = nl === -1 ? h.content.length : nl + 1;
  const line = h.content.slice(h.pos, Math.min(end, h.pos + (length || Infinity)));
  h.pos += line.length;
  return line || false;
}

export function feof(handle) {
  const h = _fileHandles.get(handle);
  return !h || h.pos >= h.content.length;
}

export function fflush(handle) { return true; }

export function ftruncate(handle, size) {
  const h = _fileHandles.get(handle);
  if (!h) return false;
  h.buffer = h.buffer.slice(0, size);
  h.content = h.content.slice(0, size);
  h.dirty = true;
  return true;
}

export function rewind(handle) {
  const h = _fileHandles.get(handle);
  if (h) h.pos = 0;
  return true;
}

export function fseek(handle, offset, whence = 0) {
  const h = _fileHandles.get(handle);
  if (!h) return -1;
  if (whence === 0) h.pos = offset;      // SEEK_SET
  else if (whence === 1) h.pos += offset; // SEEK_CUR
  else if (whence === 2) h.pos = h.content.length + offset; // SEEK_END
  return 0;
}

export function ftell(handle) {
  const h = _fileHandles.get(handle);
  return h ? h.pos : false;
}

export function fputs(handle, data, length) {
  return fwrite(handle, data, length);
}

export function flock(handle, operation) { return true; /* No-op in Workers */ }
export function set_file_buffer(handle, size) { return 0; }

export function chmod(path, mode) { return true; }
export function chown(path, user) { return true; }

export async function copy(src, dest) {
  const content = await file_get_contents(src);
  if (content === false) return false;
  await file_put_contents(dest, content);
  return true;
}

export async function touch(path, time, atime) {
  const key = _normalizePath(path);
  if (!await file_exists(path)) {
    await file_put_contents(path, '');
  }
  _metaCache.set(key, { 
    size: 0, 
    mtime: time || Math.floor(Date.now() / 1000) 
  });
  return true;
}

export function clearstatcache() {
  _metaCache.clear();
}

export function tempnam(dir, prefix) {
  return `${dir}/${prefix}${Date.now()}_${Math.random().toString(36).slice(2, 8)}.tmp`;
}

export function is_uploaded_file(path) { return true; }
export function move_uploaded_file(from, to) { return rename(from, to); }

// Directory handle emulation  
const _dirHandles = new Map();
let _dirHandleId = 0;

export async function opendir(path) {
  const entries = await scandir(path);
  const id = ++_dirHandleId;
  _dirHandles.set(id, { entries: entries || [], pos: 0 });
  return id;
}

export function readdir(handle) {
  const h = _dirHandles.get(handle);
  if (!h || h.pos >= h.entries.length) return false;
  return h.entries[h.pos++];
}

export function closedir(handle) {
  _dirHandles.delete(handle);
  return true;
}

// Normalize path for R2 key
function _normalizePath(path) {
  return String(path)
    .replace(/^\.\//, '')
    .replace(/\/+/g, '/')
    .replace(/^\//, '');
}
