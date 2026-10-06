import { DatabaseSync } from "node:sqlite"
import fs from "node:fs"
import path from "node:path"
import { fileURLToPath } from "node:url"

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export const DATA_DIR = process.env.DATABASE_PATH
  ? path.dirname(process.env.DATABASE_PATH)
  : path.resolve(__dirname, "../../data")
export const DB_PATH = path.join(DATA_DIR, "vecai.db")
export const UPLOADS_DIR = path.join(DATA_DIR, "uploads")

let db: DatabaseSync | null = null

export function getDb(): DatabaseSync {
  if (db) return db
  fs.mkdirSync(DATA_DIR, { recursive: true })
  fs.mkdirSync(UPLOADS_DIR, { recursive: true })
  db = new DatabaseSync(DB_PATH)
  db.exec("PRAGMA journal_mode = WAL;")
  db.exec("PRAGMA foreign_keys = ON;")
  migrate(db)
  return db
}

function migrate(database: DatabaseSync): void {
  database.exec(`
    CREATE TABLE IF NOT EXISTS schema_migrations (
      version    TEXT PRIMARY KEY,
      applied_at TEXT NOT NULL DEFAULT (datetime('now'))
    );
  `)
  const applied = new Set(
    database
      .prepare("SELECT version FROM schema_migrations")
      .all()
      .map((r) => r.version),
  )
  const dir = path.join(__dirname, "migrations")
  const files = fs
    .readdirSync(dir)
    .filter((f) => f.endsWith(".sql"))
    .sort()
  for (const file of files) {
    if (applied.has(file)) continue
    const sql = fs.readFileSync(path.join(dir, file), "utf8")
    database.exec("BEGIN")
    try {
      database.exec(sql)
      database
        .prepare("INSERT INTO schema_migrations (version) VALUES (?)")
        .run(file)
      database.exec("COMMIT")
      console.log(`[vecai-api] applied migration ${file}`)
    } catch (err) {
      database.exec("ROLLBACK")
      throw err
    }
  }
}
