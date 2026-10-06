import { getDb } from "./db"
import { hashPassword } from "../auth"

/** Seeds the database once, on first run (users table empty). */
export function seedIfEmpty(): void {
  const db = getDb()
  const count = (db.prepare("SELECT COUNT(*) AS c FROM users").get() as {
    c: number
  }).c
  if (count > 0) return

  console.log("[vecai-api] seeding database…")

  // Every seeded account uses this demo password (shown on the login page).
  const DEMO_PASSWORD = "demo1234"
  const hash = hashPassword(DEMO_PASSWORD)

  const users: Array<[number, string, string, string, string, string, string, string]> =
    [
      [
        1,
        "david.ochieng@vecai.co.ke",
        "David Ochieng",
        "admin",
        "Project Manager",
        "DO",
        "#1e3a5f",
        "+254 722 001 234",
      ],
      [
        2,
        "amina.hassan@vecai.co.ke",
        "Amina Hassan",
        "member",
        "Site Engineer",
        "AH",
        "#0891b2",
        "+254 733 005 678",
      ],
      [
        3,
        "francis.kimani@vecai.co.ke",
        "Francis Kimani",
        "member",
        "Quantity Surveyor",
        "FK",
        "#7c3aed",
        "+254 700 009 012",
      ],
      [
        4,
        "wanjiku.mwangi@vecai.co.ke",
        "Wanjiku Mwangi",
        "member",
        "Architect",
        "WM",
        "#db2777",
        "+254 711 003 456",
      ],
      [
        5,
        "peter.otieno@vecai.co.ke",
        "Peter Otieno",
        "member",
        "Site Manager",
        "PO",
        "#059669",
        "+254 722 007 890",
      ],
      [
        6,
        "grace.wambua@greenfield.co.ke",
        "Grace Wambua",
        "viewer",
        "Project Owner",
        "GW",
        "#d97706",
        "+254 733 002 345",
      ],
      [
        7,
        "s.njoroge@vecai.co.ke",
        "Samuel Njoroge",
        "member",
        "Structural Engineer",
        "SN",
        "#dc2626",
        "+254 720 112 233",
      ],
      [
        8,
        "fatuma.ali@vecai.co.ke",
        "Fatuma Ali",
        "member",
        "Architect",
        "FA",
        "#0d9488",
        "+254 711 445 566",
      ],
      [
        9,
        "joseph.mutua@vecai.co.ke",
        "Joseph Mutua",
        "member",
        "Site Engineer",
        "JM",
        "#6d28d9",
        "+254 722 778 899",
      ],
      [
        10,
        "esther.akinyi@vecai.co.ke",
        "Esther Akinyi",
        "member",
        "Quantity Surveyor",
        "EA",
        "#be185d",
        "+254 733 990 011",
      ],
    ]
  const iu = db.prepare(
    "INSERT INTO users (id, email, password_hash, name, role, job_title, initials, color, phone) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
  )
  // Tuple layout: (id, email, name, role, job_title, initials, color, phone);
  // password_hash is spliced into its correct column position (3rd) at insert time.
  for (const u of users) iu.run(u[0], u[1], hash, ...u.slice(2))
  // Demo login hint mirrors the seeded credentials for the primary account.
  db.prepare(
    "INSERT INTO organisation (id, name) VALUES (1, 'Victoria Eagles Construct AI')",
  ).run()

  const projects: Array<[number, string, string, string, number, string, string, string, string, number, number, string, string]> =
    [
      [
        1,
        "Lakeview Residence",
        "Kisumu, Nyanza",
        "Construction",
        64,
        "2026-01-15",
        "2026-11-18",
        "David Ochieng",
        "Kamau & Sons Construction Ltd.",
        12500000,
        8400000,
        "A 4-bedroom lakeside residential property for a private client. Includes basement parking, landscaped garden, and swimming pool. Currently at superstructure stage, ground floor columns complete.",
        "https://images.unsplash.com/photo-1590725121839-892b458a74fe?w=1600&h=400&fit=crop&auto=format",
      ],
      [
        2,
        "Westgate Commercial Complex",
        "Westlands, Nairobi",
        "Foundation",
        22,
        "2026-03-01",
        "2027-09-30",
        "Greenfield Properties Ltd.",
        "Nairobi Infrastructure Ltd.",
        45000000,
        9800000,
        "Mixed-use commercial development comprising 8 floors of Grade A office space, ground floor retail, and 3 basement parking levels. Located in Westlands CBD.",
        "https://images.unsplash.com/photo-1486325212027-8081e485255e?w=1600&h=400&fit=crop&auto=format",
      ],
      [
        3,
        "Rift Valley Community Hospital — Phase 2",
        "Nakuru, Kenya",
        "Roofing",
        81,
        "2025-07-10",
        "2026-12-20",
        "Ministry of Health — Kenya",
        "East Africa Builders Corp.",
        23000000,
        18900000,
        "Phase 2 expansion of Rift Valley Community Hospital including a new 60-bed ward block, radiology unit, and staff accommodation. Government-funded project.",
        "https://images.unsplash.com/photo-1504307651254-35680f356dfd?w=1600&h=400&fit=crop&auto=format",
      ],
      [
        4,
        "Mombasa Road Interchange Upgrade",
        "Mlolongo, Machakos",
        "Tendering",
        5,
        "2026-10-15",
        "2028-04-30",
        "Kenya National Highways Authority",
        "TBD — Tender closing 14 Oct 2026",
        78000000,
        1200000,
        "Grade separation and capacity upgrade of the A104/Mombasa Road interchange at Mlolongo. Includes new flyover, service roads, and pedestrian infrastructure.",
        "https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=1600&h=400&fit=crop&auto=format",
      ],
    ]
  const ip = db.prepare(
    `INSERT INTO projects (id, name, location, status, progress, start_date, expected_completion, client, contractor, budget_total, budget_spent, description, image_url)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
  )
  for (const p of projects) ip.run(...p)

  const memberships: Array<[number, number]> = [
    [1, 1],
    [1, 2],
    [1, 3],
    [1, 4],
    [1, 5],
    [2, 6],
    [2, 7],
    [2, 8],
    [3, 9],
    [3, 10],
    [4, 1],
    [4, 3],
  ]
  const im = db.prepare(
    "INSERT INTO project_members (project_id, user_id) VALUES (?, ?)",
  )
  for (const m of memberships) im.run(...m)

  const suppliers: Array<[number, string, string, string, string, string, number, string, string, string]> =
    [
      [
        1,
        "Bamburi Cement Ltd.",
        "Cement & Concrete",
        "James Mwenda",
        "+254 020 395 2000",
        "sales@bamburi.co.ke",
        4.5,
        "active",
        "Nairobi / Mombasa",
        "2026-09-17",
      ],
      [
        2,
        "Devki Steel Mills",
        "Steel & Metals",
        "Rita Patel",
        "+254 020 238 1000",
        "orders@devki.co.ke",
        4.2,
        "active",
        "Nairobi (Ruiru)",
        "2026-09-12",
      ],
      [
        3,
        "Mabati Rolling Mills",
        "Roofing",
        "Samuel Kiptoo",
        "+254 020 690 0000",
        "orders@mariamills.co.ke",
        4.7,
        "active",
        "Thika",
        "2026-08-30",
      ],
      [
        4,
        "East Africa Timber",
        "Timber & Formwork",
        "Jane Njeru",
        "+254 722 456 789",
        "timber@eatimber.co.ke",
        3.8,
        "active",
        "Nakuru",
        "2026-09-05",
      ],
      [
        5,
        "Nairobi Aggregates Ltd.",
        "Aggregates & Sand",
        "Moses Ongeri",
        "+254 733 789 012",
        "aggregates@naiaggregates.co.ke",
        4.0,
        "active",
        "Nairobi (Kasarani)",
        "2026-09-20",
      ],
      [
        6,
        "Kisumu Brick & Tile Co.",
        "Bricks & Blocks",
        "Jane Achieng",
        "+254 720 334 455",
        "sales@kisumubtile.co.ke",
        4.3,
        "active",
        "Kisumu",
        "2026-09-10",
      ],
      [
        7,
        "Steel & Tube Holdings",
        "Steel & Metals",
        "Patrick Ndung'u",
        "+254 020 532 4000",
        "sales@steeltubeea.co.ke",
        4.6,
        "active",
        "Nairobi (Industrial Area)",
        "2026-09-08",
      ],
    ]
  const isup = db.prepare(
    "INSERT INTO suppliers (id, name, category, contact, phone, email, rating, status, location, last_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
  )
  for (const s of suppliers) isup.run(...s)

  // Materials belong to Lakeview Residence (project 1), matching the original dataset.
  const materials: Array<[number, string, number, string, number, number, number, string, number, string]> =
    [
      [
        1,
        "Portland Cement (50kg bags)",
        1,
        "bags",
        500,
        85,
        415,
        "critical",
        680,
        "2026-09-17",
      ],
      [
        2,
        "Steel Reinforcement — Y16 bars",
        2,
        "tonnes",
        2.5,
        0.3,
        2.2,
        "critical",
        145000,
        "2026-09-12",
      ],
      [3, "River Sand (Fine)", 5, "m³", 80, 18, 62, "low", 3200, "2026-09-20"],
      [
        4,
        "Hardcore Fill (Quarry Stone)",
        5,
        "m³",
        120,
        44,
        76,
        "ok",
        1800,
        "2026-09-15",
      ],
      [
        5,
        "Burnt Clay Bricks",
        6,
        "× 1,000",
        40,
        18,
        22,
        "low",
        12000,
        "2026-09-10",
      ],
      [
        6,
        "Timber Formwork (2×4 in.)",
        4,
        "pcs",
        600,
        220,
        380,
        "ok",
        450,
        "2026-09-05",
      ],
      [
        7,
        "Binding Wire (1.6mm)",
        2,
        "rolls",
        80,
        6,
        74,
        "critical",
        2800,
        "2026-09-03",
      ],
      [8, "BRC Mesh A142", 7, "sheets", 150, 60, 90, "ok", 4500, "2026-09-08"],
      [
        9,
        "Roofing Sheets — Gauge 30 (MRM)",
        3,
        "sheets",
        200,
        200,
        0,
        "ordered",
        1950,
        "2026-10-05",
      ],
      [
        10,
        "PVC Pipes — 110mm uPVC",
        5,
        "lengths",
        120,
        65,
        55,
        "ok",
        2400,
        "2026-09-14",
      ],
      [
        11,
        "Aggregate — 20mm Crushed Stone",
        5,
        "m³",
        60,
        22,
        38,
        "low",
        4100,
        "2026-09-18",
      ],
      [
        12,
        "Steel Reinforcement — Y10 bars",
        2,
        "tonnes",
        1.2,
        0.7,
        0.5,
        "ok",
        145000,
        "2026-09-12",
      ],
    ]
  const imat = db.prepare(
    `INSERT INTO materials (id, name, supplier_id, unit, required_qty, available_qty, used_qty, status, unit_cost, last_delivery, project_id)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
  )
  for (const m of materials) imat.run(...m)

  const tasks: Array<[number, number, number | null, string, string, string, string, string, string]> =
    [
      [
        1,
        1,
        5,
        "Foundation waterproofing — Block A",
        "Apply bituminous waterproofing membrane to all foundation walls below ground level. Includes pipe penetration sealing.",
        "2026-09-20",
        "2026-09-28",
        "in-progress",
        "high",
      ],
      [
        2,
        1,
        2,
        "Concrete column casting — Ground floor (C1–C6)",
        "Cast concrete for columns C1 through C6 on the ground floor. Mix design as per structural drawings Rev B.",
        "2026-09-22",
        "2026-09-30",
        "in-progress",
        "high",
      ],
      [
        3,
        1,
        1,
        "Electrical conduit installation — Level 2",
        "",
        "2026-09-27",
        "2026-10-08",
        "not-started",
        "medium",
      ],
      [
        4,
        1,
        6,
        "Plumbing rough-in — First floor bathrooms",
        "Blocked: awaiting delivery of uPVC soil pipe fittings. Supplier notified 23 Sept.",
        "2026-09-18",
        "2026-09-26",
        "blocked",
        "high",
      ],
      [
        5,
        1,
        3,
        "Submit structural drawings to county",
        "",
        "2026-09-08",
        "2026-09-20",
        "completed",
        "high",
      ],
      [
        6,
        1,
        5,
        "Install roofing structure — Phase 1",
        "",
        "2026-10-05",
        "2026-10-22",
        "not-started",
        "medium",
      ],
      [
        7,
        1,
        3,
        "BOQ review and cost reconciliation",
        "",
        "2026-09-24",
        "2026-09-27",
        "in-progress",
        "medium",
      ],
      [
        8,
        1,
        2,
        "Soil compaction test — Block B",
        "",
        "2026-09-10",
        "2026-09-20",
        "completed",
        "high",
      ],
      [
        9,
        2,
        1,
        "Procure steel reinforcement — Y16 (Phase 2)",
        "",
        "2026-09-20",
        "2026-09-30",
        "in-progress",
        "high",
      ],
      [
        10,
        3,
        9,
        "Final roofing inspection sign-off",
        "",
        "2026-09-25",
        "2026-10-03",
        "not-started",
        "medium",
      ],
      [
        11,
        1,
        5,
        "Window and door frame installation — Level 1",
        "",
        "2026-10-10",
        "2026-10-18",
        "not-started",
        "medium",
      ],
      [
        12,
        4,
        3,
        "Prepare tender submission documents",
        "",
        "2026-09-22",
        "2026-10-12",
        "in-progress",
        "high",
      ],
    ]
  const it = db.prepare(
    `INSERT INTO tasks (id, project_id, assignee_id, name, description, start_date, due_date, status, priority)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
  )
  for (const t of tasks) it.run(...t)

  const documents: Array<[number, number, number, string, string, string, string, number, string]> =
    [
      [
        1,
        1,
        4,
        "Architectural Drawings — Full Set Rev C",
        "Architectural Drawing",
        "Rev C",
        "architectural-drawings-rev-c.pdf",
        14.2 * 1024 * 1024,
        "PDF",
      ],
      [
        2,
        1,
        2,
        "Structural Design Report & Calculations",
        "Structural Drawing",
        "Rev B",
        "structural-design-report.pdf",
        8.7 * 1024 * 1024,
        "PDF",
      ],
      [
        3,
        1,
        3,
        "Bill of Quantities — Substructure & Superstructure",
        "BOQ",
        "2.1",
        "boq-substructure-superstructure.xlsx",
        2.1 * 1024 * 1024,
        "XLSX",
      ],
      [
        4,
        1,
        1,
        "NEMA Environmental Impact Assessment Permit",
        "Permit",
        "1.0",
        "nema-eia-permit.pdf",
        1.4 * 1024 * 1024,
        "PDF",
      ],
      [
        5,
        1,
        1,
        "Main Contract Agreement — Kamau & Sons",
        "Contract",
        "1.0",
        "main-contract-agreement.pdf",
        3.2 * 1024 * 1024,
        "PDF",
      ],
      [
        6,
        1,
        5,
        "Site Progress Report — Week 36",
        "Site Report",
        "W36",
        "site-progress-report-w36.pdf",
        4.8 * 1024 * 1024,
        "PDF",
      ],
      [
        7,
        1,
        2,
        "Concrete Mix Design Certificate — Bamburi C30",
        "Inspection Report",
        "1.0",
        "concrete-mix-design-certificate.pdf",
        1.1 * 1024 * 1024,
        "PDF",
      ],
      [
        8,
        1,
        3,
        "Invoice #INV-2026-0892 — Bamburi Cement Delivery",
        "Invoice",
        "—",
        "invoice-inv-2026-0892.pdf",
        0.3 * 1024 * 1024,
        "PDF",
      ],
      [
        9,
        1,
        1,
        "NCA Contractor Registration Certificate",
        "Permit",
        "1.0",
        "nca-contractor-registration.pdf",
        0.8 * 1024 * 1024,
        "PDF",
      ],
      [
        10,
        1,
        5,
        "Site Progress Photos — Week 36",
        "Photo",
        "W36",
        "site-progress-photos-w36.zip",
        22.4 * 1024 * 1024,
        "ZIP",
      ],
    ]
  const idoc = db.prepare(
    `INSERT INTO documents (id, project_id, uploaded_by, name, category, version, original_name, stored_name, size_bytes, format)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
  )
  // Seeded documents are metadata records without a stored file (no binaries exist).
  for (const d of documents)
    idoc.run(
      d[0],
      d[1],
      d[2],
      d[3],
      d[4],
      d[5],
      d[6],
      "seeded/" + d[6],
      d[7],
      d[8],
    )

  const siteUpdates: Array<[number, number, number, string, string, string, number, string, string, string, number]> =
    [
      [
        1,
        1,
        5,
        "2026-09-25",
        "Partly Cloudy",
        "24",
        28,
        "Completed casting of columns C3 and C4 on ground floor. Formwork struck for beam B1 — concrete quality satisfactory. Backfilling completed on west elevation. Surface dressing applied to internal access road.",
        "Delivery of steel Y16 bars from Devki Steel Mills delayed. Expected arrival: Tuesday 29 September. Site team adjusting schedule — working on columns C5 and C6 formwork in meantime.",
        "Client visited site at 11:00 AM. Generally satisfied with progress. Requested update on roofing material delivery timeline.",
        6,
      ],
      [
        2,
        1,
        5,
        "2026-09-24",
        "Light Rain",
        "21",
        22,
        "Plumbing rough-in continued — second floor east wing. Steel placement for columns C3 and C4 completed and checked. Carpenter team fabricating formwork for beams B4–B6.",
        "Rain halted work at 14:30. Resumed at 15:45. No structural impact.",
        "Safety inspection conducted by site engineer. All PPE compliant. Fire extinguisher near generator replaced.",
        3,
      ],
    ]
  const isu = db.prepare(
    `INSERT INTO site_updates (id, project_id, submitted_by, date, weather, temperature, workers_on_site, work_completed, issues, notes, photo_count)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
  )
  for (const s of siteUpdates) isu.run(...s)

  const isum = db.prepare(
    "INSERT INTO site_update_materials (site_update_id, name, quantity, unit) VALUES (?, ?, ?, ?)",
  )
  isum.run(1, "Portland Cement", 14, "bags")
  isum.run(1, "River Sand", 2.5, "m³")
  isum.run(1, "Steel Reinforcement Y16", 0.12, "tonnes")
  isum.run(1, "Binding Wire", 2, "rolls")
  isum.run(2, "Portland Cement", 6, "bags")
  isum.run(2, "BRC Mesh A142", 4, "sheets")

  const months = [
    ["2026-01", 850000, 820000],
    ["2026-02", 1100000, 1050000],
    ["2026-03", 1400000, 1380000],
    ["2026-04", 1600000, 1720000],
    ["2026-05", 900000, 860000],
    ["2026-06", 750000, 800000],
    ["2026-07", 680000, 710000],
    ["2026-08", 600000, 580000],
    ["2026-09", 700000, 480000],
  ]
  const ibm = db.prepare(
    "INSERT INTO budget_months (project_id, ym, planned, actual) VALUES (1, ?, ?, ?)",
  )
  for (const [ym, planned, actual] of months) ibm.run(ym, planned, actual)

  const activity: Array<[number | null, number | null, string, string, string, string]> =
    [
      [
        1,
        5,
        "Peter Otieno",
        "site-update",
        "Submitted site update — Week 36, Day 5",
        "2026-09-25 10:12:00",
      ],
      [
        1,
        4,
        "Wanjiku Mwangi",
        "document",
        "Uploaded Architectural Drawings Rev C",
        "2026-09-20 09:40:00",
      ],
      [
        1,
        3,
        "Francis Kimani",
        "task",
        'Marked "Submit structural drawings" as completed',
        "2026-09-20 16:42:00",
      ],
      [
        1,
        5,
        "Peter Otieno",
        "material",
        "Logged delivery: 6 bags Portland Cement",
        "2026-09-19 09:15:00",
      ],
      [
        1,
        1,
        "David Ochieng",
        "comment",
        'Added comment on Task "Plumbing rough-in": Supplier confirmed Thursday delivery',
        "2026-09-18 11:05:00",
      ],
      [
        1,
        3,
        "Francis Kimani",
        "document",
        "Uploaded Invoice #INV-2026-0892 — Bamburi Cement",
        "2026-09-17 14:22:00",
      ],
      [
        1,
        5,
        "Peter Otieno",
        "task",
        'Task "Plumbing rough-in" marked as Blocked',
        "2026-09-16 08:30:00",
      ],
      [
        1,
        null,
        "System",
        "material",
        "Stock alert raised: Binding Wire at 6 rolls remaining",
        "2026-09-16 08:00:00",
      ],
    ]
  const ia = db.prepare(
    "INSERT INTO activity_log (project_id, actor_id, actor_name, type, text, created_at) VALUES (?, ?, ?, ?, ?, ?)",
  )
  for (const a of activity) ia.run(...a)

  console.log("[vecai-api] seed complete")
}
