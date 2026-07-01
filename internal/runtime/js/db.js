// PHP Database functions -> Cloudflare D1
import { getD1 } from './index.js';

// Connection emulation
let _dbConnected = false;
const _resultSets = new Map();
let _resultId = 0;

export async function mysql_connect() { _dbConnected = true; return true; }
export const mysqli_connect = mysql_connect;

export async function mysql_query(sql, link) {
  const d1 = getD1();
  if (!d1) { console.warn('D1 not available'); return false; }
  try {
    const result = await d1.prepare(sql).all();
    const id = ++_resultId;
    _resultSets.set(id, { results: result.results || [], index: 0 });
    return id;
  } catch (e) { console.error('Query error:', e); return false; }
}
export const mysqli_query = mysql_query;

export function mysql_fetch_array(result) {
  const rs = _resultSets.get(result);
  if (!rs || rs.index >= rs.results.length) return null;
  const row = rs.results[rs.index++];
  // Return both numeric and assoc keys
  const combined = { ...row };
  Object.values(row).forEach((v, i) => { combined[i] = v; });
  return combined;
}

export function mysql_fetch_assoc(result) {
  const rs = _resultSets.get(result);
  if (!rs || rs.index >= rs.results.length) return null;
  return rs.results[rs.index++];
}
export const mysqli_fetch_assoc = mysql_fetch_assoc;

export function mysql_num_rows(result) {
  const rs = _resultSets.get(result);
  return rs ? rs.results.length : 0;
}
export const mysqli_num_rows = mysql_num_rows;

export function mysql_free_result(result) { _resultSets.delete(result); return true; }
export function mysql_close() { _dbConnected = false; return true; }
export function mysql_real_escape_string(s) { return String(s).replace(/'/g, "''"); }
export const mysqli_real_escape_string = mysql_real_escape_string;
export function mysql_error() { return ''; }
export function mysql_affected_rows() { return 0; }
export function mysql_insert_id() { return 0; }

// PDO emulation (basic)
export class PDO {
  constructor(dsn, user, pass) { this._d1 = getD1(); }
  async prepare(sql) { return new PDOStatement(this._d1, sql); }
  async exec(sql) { if (!this._d1) return 0; await this._d1.prepare(sql).run(); return 1; }
  async query(sql) { return this.prepare(sql); }
  quote(s) { return `'${String(s).replace(/'/g, "''")}'`; }
}

export class PDOStatement {
  constructor(d1, sql) { this._d1 = d1; this._sql = sql; this._params = []; this._results = null; this._index = 0; }
  bindValue(param, value) { this._params.push([param, value]); }
  async execute(params) {
    if (!this._d1) return false;
    let stmt = this._d1.prepare(this._sql);
    const binds = params || this._params.map(([, v]) => v);
    if (binds.length) stmt = stmt.bind(...binds);
    const result = await stmt.all();
    this._results = result.results || [];
    this._index = 0;
    return true;
  }
  fetch() {
    if (!this._results || this._index >= this._results.length) return false;
    return this._results[this._index++];
  }
  fetchAll() { return this._results || []; }
  rowCount() { return this._results ? this._results.length : 0; }
}
