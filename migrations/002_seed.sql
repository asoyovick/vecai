-- users
INSERT OR IGNORE INTO users VALUES(1,'david.ochieng@vecai.co.ke','scrypt:12ca40480567f50305269bf2e2fd5bd5:f121a17ee06fdf59277527002e40adedb0965507410d03febce080229d06ccfc61dad59f1ddfc5cf1b9dd374a5db2842f39492643152879ab319af760515039c','David Ochieng','admin','Project Manager','DO','#1e3a5f','+254 722 001 234','2026-10-05 13:00:38','2026-10-05 13:05:33');
INSERT OR IGNORE INTO users VALUES(2,'amina.hassan@vecai.co.ke','scrypt:75b728618dde3111d4c20ec845da80bf:66f430e348c52b58a63290206d0ad47210053c2e4af0600215e716d1378ccd786bb56992c4a5d1bf94e9654c88a87a420be83b30958f8a26f773191337aad0b3','Amina Hassan','member','Site Engineer','AH','#0891b2','+254 733 005 678','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO users VALUES(3,'francis.kimani@vecai.co.ke','scrypt:75b728618dde3111d4c20ec845da80bf:66f430e348c52b58a63290206d0ad47210053c2e4af0600215e716d1378ccd786bb56992c4a5d1bf94e9654c88a87a420be83b30958f8a26f773191337aad0b3','Francis Kimani','member','Quantity Surveyor','FK','#7c3aed','+254 700 009 012','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO users VALUES(4,'wanjiku.mwangi@vecai.co.ke','scrypt:75b728618dde3111d4c20ec845da80bf:66f430e348c52b58a63290206d0ad47210053c2e4af0600215e716d1378ccd786bb56992c4a5d1bf94e9654c88a87a420be83b30958f8a26f773191337aad0b3','Wanjiku Mwangi','member','Architect','WM','#db2777','+254 711 003 456','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO users VALUES(5,'peter.otieno@vecai.co.ke','scrypt:75b728618dde3111d4c20ec845da80bf:66f430e348c52b58a63290206d0ad47210053c2e4af0600215e716d1378ccd786bb56992c4a5d1bf94e9654c88a87a420be83b30958f8a26f773191337aad0b3','Peter Otieno','member','Site Manager','PO','#059669','+254 722 007 890','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO users VALUES(6,'grace.wambua@greenfield.co.ke','scrypt:75b728618dde3111d4c20ec845da80bf:66f430e348c52b58a63290206d0ad47210053c2e4af0600215e716d1378ccd786bb56992c4a5d1bf94e9654c88a87a420be83b30958f8a26f773191337aad0b3','Grace Wambua','viewer','Project Owner','GW','#d97706','+254 733 002 345','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO users VALUES(7,'s.njoroge@vecai.co.ke','scrypt:75b728618dde3111d4c20ec845da80bf:66f430e348c52b58a63290206d0ad47210053c2e4af0600215e716d1378ccd786bb56992c4a5d1bf94e9654c88a87a420be83b30958f8a26f773191337aad0b3','Samuel Njoroge','member','Structural Engineer','SN','#dc2626','+254 720 112 233','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO users VALUES(8,'fatuma.ali@vecai.co.ke','scrypt:75b728618dde3111d4c20ec845da80bf:66f430e348c52b58a63290206d0ad47210053c2e4af0600215e716d1378ccd786bb56992c4a5d1bf94e9654c88a87a420be83b30958f8a26f773191337aad0b3','Fatuma Ali','member','Architect','FA','#0d9488','+254 711 445 566','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO users VALUES(9,'joseph.mutua@vecai.co.ke','scrypt:75b728618dde3111d4c20ec845da80bf:66f430e348c52b58a63290206d0ad47210053c2e4af0600215e716d1378ccd786bb56992c4a5d1bf94e9654c88a87a420be83b30958f8a26f773191337aad0b3','Joseph Mutua','member','Site Engineer','JM','#6d28d9','+254 722 778 899','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO users VALUES(10,'esther.akinyi@vecai.co.ke','scrypt:75b728618dde3111d4c20ec845da80bf:66f430e348c52b58a63290206d0ad47210053c2e4af0600215e716d1378ccd786bb56992c4a5d1bf94e9654c88a87a420be83b30958f8a26f773191337aad0b3','Esther Akinyi','member','Quantity Surveyor','EA','#be185d','+254 733 990 011','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO users VALUES(11,'test.user@vecai.co.ke','scrypt:c9f4cba93d0c2b25808ab18c6fa2bee4:f09c09b08651d1f10e0b82c2975649937f46814c90b0357103f168580d95339ee8dade811d8651d0b0d22774c66c41b9f05f594f1051b0e71bbc147f70861fb2','Test User','member','Engineer','TU','#db2777','','2026-10-05 13:04:15','2026-10-05 13:04:15');
-- organisation
INSERT OR IGNORE INTO organisation VALUES(1,'Victoria Eagles Construct AI','Construction Company','Kenya','KES — Kenyan Shilling');
-- projects
INSERT OR IGNORE INTO projects VALUES(1,'Lakeview Residence','Kisumu, Nyanza','Construction',64,'2026-01-15','2026-11-18','David Ochieng','Kamau & Sons Construction Ltd.',12500000,8400000,'A 4-bedroom lakeside residential property for a private client. Includes basement parking, landscaped garden, and swimming pool. Currently at superstructure stage, ground floor columns complete.','https://images.unsplash.com/photo-1590725121839-892b458a74fe?w=1600&h=400&fit=crop&auto=format','2026-10-05 13:00:38','2026-10-05 14:11:54');
INSERT OR IGNORE INTO projects VALUES(2,'Westgate Commercial Complex','Westlands, Nairobi','Foundation',22,'2026-03-01','2027-09-30','Greenfield Properties Ltd.','Nairobi Infrastructure Ltd.',45000000,9800000,'Mixed-use commercial development comprising 8 floors of Grade A office space, ground floor retail, and 3 basement parking levels. Located in Westlands CBD.','https://images.unsplash.com/photo-1486325212027-8081e485255e?w=1600&h=400&fit=crop&auto=format','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO projects VALUES(3,'Rift Valley Community Hospital — Phase 2','Nakuru, Kenya','Roofing',81,'2025-07-10','2026-12-20','Ministry of Health — Kenya','East Africa Builders Corp.',23000000,18900000,'Phase 2 expansion of Rift Valley Community Hospital including a new 60-bed ward block, radiology unit, and staff accommodation. Government-funded project.','https://images.unsplash.com/photo-1504307651254-35680f356dfd?w=1600&h=400&fit=crop&auto=format','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO projects VALUES(4,'Mombasa Road Interchange Upgrade','Mlolongo, Machakos','Tendering',5,'2026-10-15','2028-04-30','Kenya National Highways Authority','TBD — Tender closing 14 Oct 2026',78000000,1200000,'Grade separation and capacity upgrade of the A104/Mombasa Road interchange at Mlolongo. Includes new flyover, service roads, and pedestrian infrastructure.','https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=1600&h=400&fit=crop&auto=format','2026-10-05 13:00:38','2026-10-05 13:00:38');
-- project_members
INSERT OR IGNORE INTO project_members VALUES(1,1);
INSERT OR IGNORE INTO project_members VALUES(1,2);
INSERT OR IGNORE INTO project_members VALUES(1,3);
INSERT OR IGNORE INTO project_members VALUES(1,4);
INSERT OR IGNORE INTO project_members VALUES(1,5);
INSERT OR IGNORE INTO project_members VALUES(2,6);
INSERT OR IGNORE INTO project_members VALUES(2,7);
INSERT OR IGNORE INTO project_members VALUES(2,8);
INSERT OR IGNORE INTO project_members VALUES(3,9);
INSERT OR IGNORE INTO project_members VALUES(3,10);
INSERT OR IGNORE INTO project_members VALUES(4,1);
INSERT OR IGNORE INTO project_members VALUES(4,3);
-- suppliers
INSERT OR IGNORE INTO suppliers VALUES(1,'Bamburi Cement Ltd.','Cement & Concrete','James Mwenda','+254 020 395 2000','sales@bamburi.co.ke',4.5,'active','Nairobi / Mombasa','2026-09-17','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO suppliers VALUES(2,'Devki Steel Mills','Steel & Metals','Rita Patel','+254 020 238 1000','orders@devki.co.ke',4.20000000000000017,'active','Nairobi (Ruiru)','2026-09-12','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO suppliers VALUES(3,'Mabati Rolling Mills','Roofing','Samuel Kiptoo','+254 020 690 0000','orders@mariamills.co.ke',4.70000000000000017,'active','Thika','2026-08-30','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO suppliers VALUES(4,'East Africa Timber','Timber & Formwork','Jane Njeru','+254 722 456 789','timber@eatimber.co.ke',3.79999999999999982,'active','Nakuru','2026-09-05','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO suppliers VALUES(5,'Nairobi Aggregates Ltd.','Aggregates & Sand','Moses Ongeri','+254 733 789 012','aggregates@naiaggregates.co.ke',4.0,'active','Nairobi (Kasarani)','2026-09-20','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO suppliers VALUES(6,'Kisumu Brick & Tile Co.','Bricks & Blocks','Jane Achieng','+254 720 334 455','sales@kisumubtile.co.ke',4.29999999999999982,'active','Kisumu','2026-09-10','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO suppliers VALUES(7,'Steel & Tube Holdings','Steel & Metals','Patrick Ndung''u','+254 020 532 4000','sales@steeltubeea.co.ke',4.59999999999999964,'active','Nairobi (Industrial Area)','2026-09-08','2026-10-05 13:00:38','2026-10-05 13:00:38');
-- materials
INSERT OR IGNORE INTO materials VALUES(1,1,1,'Portland Cement (50kg bags)','bags',500.0,185.0,415.0,'critical',680,'2026-10-05','2026-10-05 13:00:38','2026-10-05 13:03:23');
INSERT OR IGNORE INTO materials VALUES(2,1,2,'Steel Reinforcement — Y16 bars','tonnes',2.5,0.299999999999999988,2.20000000000000017,'critical',145000,'2026-09-12','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO materials VALUES(3,1,5,'River Sand (Fine)','m³',80.0,18.0,62.0,'low',3200,'2026-09-20','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO materials VALUES(4,1,5,'Hardcore Fill (Quarry Stone)','m³',120.0,44.0,76.0,'ok',1800,'2026-09-15','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO materials VALUES(5,1,6,'Burnt Clay Bricks','× 1,000',40.0,18.0,22.0,'low',12000,'2026-09-10','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO materials VALUES(6,1,4,'Timber Formwork (2×4 in.)','pcs',600.0,220.0,380.0,'ok',450,'2026-09-05','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO materials VALUES(7,1,2,'Binding Wire (1.6mm)','rolls',80.0,6.0,74.0,'critical',2800,'2026-09-03','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO materials VALUES(8,1,7,'BRC Mesh A142','sheets',150.0,60.0,90.0,'ok',4500,'2026-09-08','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO materials VALUES(9,1,3,'Roofing Sheets — Gauge 30 (MRM)','sheets',200.0,200.0,0.0,'ordered',1950,'2026-10-05','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO materials VALUES(10,1,5,'PVC Pipes — 110mm uPVC','lengths',120.0,65.0,55.0,'ok',2400,'2026-09-14','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO materials VALUES(11,1,5,'Aggregate — 20mm Crushed Stone','m³',60.0,22.0,38.0,'low',4100,'2026-09-18','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO materials VALUES(12,1,2,'Steel Reinforcement — Y10 bars','tonnes',1.19999999999999995,0.699999999999999955,0.5,'ok',145000,'2026-09-12','2026-10-05 13:00:38','2026-10-05 13:00:38');
-- material_deliveries
INSERT OR IGNORE INTO material_deliveries VALUES(1,1,100.0,'test',1,'2026-10-05 13:03:23');
-- tasks
INSERT OR IGNORE INTO tasks VALUES(1,1,5,'Foundation waterproofing — Block A','Apply bituminous waterproofing membrane to all foundation walls below ground level. Includes pipe penetration sealing.','2026-09-20','2026-09-28','in-progress','high','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(2,1,2,'Concrete column casting — Ground floor (C1–C6)','Cast concrete for columns C1 through C6 on the ground floor. Mix design as per structural drawings Rev B.','2026-09-22','2026-09-30','in-progress','high','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(3,1,1,'Electrical conduit installation — Level 2','','2026-09-27','2026-10-08','not-started','medium','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(4,1,6,'Plumbing rough-in — First floor bathrooms','Blocked: awaiting delivery of uPVC soil pipe fittings. Supplier notified 23 Sept.','2026-09-18','2026-09-26','blocked','high','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(5,1,3,'Submit structural drawings to county','','2026-09-08','2026-09-20','completed','high','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(6,1,5,'Install roofing structure — Phase 1','','2026-10-05','2026-10-22','not-started','medium','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(7,1,3,'BOQ review and cost reconciliation','','2026-09-24','2026-09-27','in-progress','medium','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(8,1,2,'Soil compaction test — Block B','','2026-09-10','2026-09-20','completed','high','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(9,2,1,'Procure steel reinforcement — Y16 (Phase 2)','','2026-09-20','2026-09-30','in-progress','high','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(10,3,9,'Final roofing inspection sign-off','','2026-09-25','2026-10-03','not-started','medium','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(11,1,5,'Window and door frame installation — Level 1','','2026-10-10','2026-10-18','not-started','medium','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(12,4,3,'Prepare tender submission documents','','2026-09-22','2026-10-12','in-progress','high','2026-10-05 13:00:38','2026-10-05 13:00:38');
INSERT OR IGNORE INTO tasks VALUES(13,1,2,'Test task from API','',NULL,'2026-10-20','completed','high','2026-10-05 13:03:23','2026-10-05 13:04:15');
INSERT OR IGNORE INTO tasks VALUES(14,1,NULL,'SMOKE 1791207993948','',NULL,'2026-12-01','completed','low','2026-10-05 13:46:33','2026-10-05 13:46:34');
INSERT OR IGNORE INTO tasks VALUES(15,1,NULL,'SMOKE 1791209269153','',NULL,'2026-12-01','completed','low','2026-10-05 14:07:49','2026-10-05 14:07:49');
INSERT OR IGNORE INTO tasks VALUES(16,1,NULL,'SMOKE 1791209514325','',NULL,'2026-12-01','completed','low','2026-10-05 14:11:54','2026-10-05 14:11:54');
-- documents
INSERT OR IGNORE INTO documents VALUES(1,1,4,'Architectural Drawings — Full Set Rev C','Architectural Drawing','Rev C','architectural-drawings-rev-c.pdf','seeded/architectural-drawings-rev-c.pdf','application/octet-stream',14889779.1999999992,'PDF','2026-10-05 13:00:38');
INSERT OR IGNORE INTO documents VALUES(2,1,2,'Structural Design Report & Calculations','Structural Drawing','Rev B','structural-design-report.pdf','seeded/structural-design-report.pdf','application/octet-stream',9122611.19999999925,'PDF','2026-10-05 13:00:38');
INSERT OR IGNORE INTO documents VALUES(3,1,3,'Bill of Quantities — Substructure & Superstructure','BOQ','2.1','boq-substructure-superstructure.xlsx','seeded/boq-substructure-superstructure.xlsx','application/octet-stream',2202009.60000000009,'XLSX','2026-10-05 13:00:38');
INSERT OR IGNORE INTO documents VALUES(4,1,1,'NEMA Environmental Impact Assessment Permit','Permit','1.0','nema-eia-permit.pdf','seeded/nema-eia-permit.pdf','application/octet-stream',1468006.3999999999,'PDF','2026-10-05 13:00:38');
INSERT OR IGNORE INTO documents VALUES(5,1,1,'Main Contract Agreement — Kamau & Sons','Contract','1.0','main-contract-agreement.pdf','seeded/main-contract-agreement.pdf','application/octet-stream',3355443.20000000018,'PDF','2026-10-05 13:00:38');
INSERT OR IGNORE INTO documents VALUES(6,1,5,'Site Progress Report — Week 36','Site Report','W36','site-progress-report-w36.pdf','seeded/site-progress-report-w36.pdf','application/octet-stream',5033164.79999999981,'PDF','2026-10-05 13:00:38');
INSERT OR IGNORE INTO documents VALUES(7,1,2,'Concrete Mix Design Certificate — Bamburi C30','Inspection Report','1.0','concrete-mix-design-certificate.pdf','seeded/concrete-mix-design-certificate.pdf','application/octet-stream',1153433.60000000009,'PDF','2026-10-05 13:00:38');
INSERT OR IGNORE INTO documents VALUES(8,1,3,'Invoice #INV-2026-0892 — Bamburi Cement Delivery','Invoice','—','invoice-inv-2026-0892.pdf','seeded/invoice-inv-2026-0892.pdf','application/octet-stream',314572.799999999988,'PDF','2026-10-05 13:00:38');
INSERT OR IGNORE INTO documents VALUES(9,1,1,'NCA Contractor Registration Certificate','Permit','1.0','nca-contractor-registration.pdf','seeded/nca-contractor-registration.pdf','application/octet-stream',838860.800000000046,'PDF','2026-10-05 13:00:38');
INSERT OR IGNORE INTO documents VALUES(10,1,5,'Site Progress Photos — Week 36','Photo','W36','site-progress-photos-w36.zip','seeded/site-progress-photos-w36.zip','application/octet-stream',23488102.3999999985,'ZIP','2026-10-05 13:00:38');
INSERT OR IGNORE INTO documents VALUES(11,1,1,'Test Document','Site Report','1.1','testdoc.txt','1791205403423-0d31b6afc0f0.txt','text/plain',13,'TXT','2026-10-05 13:03:23');
-- site_updates
INSERT OR IGNORE INTO site_updates VALUES(1,1,5,'2026-09-25','Partly Cloudy','24',28,'Completed casting of columns C3 and C4 on ground floor. Formwork struck for beam B1 — concrete quality satisfactory. Backfilling completed on west elevation. Surface dressing applied to internal access road.','Delivery of steel Y16 bars from Devki Steel Mills delayed. Expected arrival: Tuesday 29 September. Site team adjusting schedule — working on columns C5 and C6 formwork in meantime.','Client visited site at 11:00 AM. Generally satisfied with progress. Requested update on roofing material delivery timeline.',6,'2026-10-05 13:00:38');
INSERT OR IGNORE INTO site_updates VALUES(2,1,5,'2026-09-24','Light Rain','21',22,'Plumbing rough-in continued — second floor east wing. Steel placement for columns C3 and C4 completed and checked. Carpenter team fabricating formwork for beams B4–B6.','Rain halted work at 14:30. Resumed at 15:45. No structural impact.','Safety inspection conducted by site engineer. All PPE compliant. Fire extinguisher near generator replaced.',3,'2026-10-05 13:00:38');
INSERT OR IGNORE INTO site_updates VALUES(3,1,1,'2026-10-05','Sunny','27',30,'Cast column C5.','','',0,'2026-10-05 13:04:15');
INSERT OR IGNORE INTO site_updates VALUES(4,1,1,'2026-10-05','Sunny','26',25,'SMOKE TEST update 1791207994064','','',0,'2026-10-05 13:46:34');
INSERT OR IGNORE INTO site_updates VALUES(5,1,1,'2026-10-05','Sunny','26',25,'SMOKE TEST update 1791209269366','','',0,'2026-10-05 14:07:49');
INSERT OR IGNORE INTO site_updates VALUES(6,1,1,'2026-10-05','Sunny','26',25,'SMOKE TEST update 1791209514481','','',0,'2026-10-05 14:11:54');
-- site_update_materials
INSERT OR IGNORE INTO site_update_materials VALUES(1,1,'Portland Cement',14.0,'bags');
INSERT OR IGNORE INTO site_update_materials VALUES(2,1,'River Sand',2.5,'m³');
INSERT OR IGNORE INTO site_update_materials VALUES(3,1,'Steel Reinforcement Y16',0.119999999999999995,'tonnes');
INSERT OR IGNORE INTO site_update_materials VALUES(4,1,'Binding Wire',2.0,'rolls');
INSERT OR IGNORE INTO site_update_materials VALUES(5,2,'Portland Cement',6.0,'bags');
INSERT OR IGNORE INTO site_update_materials VALUES(6,2,'BRC Mesh A142',4.0,'sheets');
INSERT OR IGNORE INTO site_update_materials VALUES(7,3,'Cement',10.0,'bags');
INSERT OR IGNORE INTO site_update_materials VALUES(8,4,'Test Sand',3.0,'m³');
INSERT OR IGNORE INTO site_update_materials VALUES(9,5,'Test Sand',3.0,'m³');
INSERT OR IGNORE INTO site_update_materials VALUES(10,6,'Test Sand',3.0,'m³');
-- budget_months
INSERT OR IGNORE INTO budget_months VALUES(1,'2026-01',850000,820000);
INSERT OR IGNORE INTO budget_months VALUES(1,'2026-02',1100000,1050000);
INSERT OR IGNORE INTO budget_months VALUES(1,'2026-03',1400000,1380000);
INSERT OR IGNORE INTO budget_months VALUES(1,'2026-04',1600000,1720000);
INSERT OR IGNORE INTO budget_months VALUES(1,'2026-05',900000,860000);
INSERT OR IGNORE INTO budget_months VALUES(1,'2026-06',750000,800000);
INSERT OR IGNORE INTO budget_months VALUES(1,'2026-07',680000,710000);
INSERT OR IGNORE INTO budget_months VALUES(1,'2026-08',600000,580000);
INSERT OR IGNORE INTO budget_months VALUES(1,'2026-09',700000,480000);
-- activity_log
INSERT OR IGNORE INTO activity_log VALUES(1,1,5,'Peter Otieno','site-update','Submitted site update — Week 36, Day 5','2026-09-25 10:12:00');
INSERT OR IGNORE INTO activity_log VALUES(2,1,4,'Wanjiku Mwangi','document','Uploaded Architectural Drawings Rev C','2026-09-20 09:40:00');
INSERT OR IGNORE INTO activity_log VALUES(3,1,3,'Francis Kimani','task','Marked "Submit structural drawings" as completed','2026-09-20 16:42:00');
INSERT OR IGNORE INTO activity_log VALUES(4,1,5,'Peter Otieno','material','Logged delivery: 6 bags Portland Cement','2026-09-19 09:15:00');
INSERT OR IGNORE INTO activity_log VALUES(5,1,1,'David Ochieng','comment','Added comment on Task "Plumbing rough-in": Supplier confirmed Thursday delivery','2026-09-18 11:05:00');
INSERT OR IGNORE INTO activity_log VALUES(6,1,3,'Francis Kimani','document','Uploaded Invoice #INV-2026-0892 — Bamburi Cement','2026-09-17 14:22:00');
INSERT OR IGNORE INTO activity_log VALUES(7,1,5,'Peter Otieno','task','Task "Plumbing rough-in" marked as Blocked','2026-09-16 08:30:00');
INSERT OR IGNORE INTO activity_log VALUES(8,1,NULL,'System','material','Stock alert raised: Binding Wire at 6 rolls remaining','2026-09-16 08:00:00');
INSERT OR IGNORE INTO activity_log VALUES(9,1,1,'David Ochieng','task','Created task “Test task from API”','2026-10-05 13:03:23');
INSERT OR IGNORE INTO activity_log VALUES(10,1,1,'David Ochieng','material','Logged delivery: 100 bags Portland Cement (50kg bags)','2026-10-05 13:03:23');
INSERT OR IGNORE INTO activity_log VALUES(11,1,1,'David Ochieng','document','Uploaded “Test Document”','2026-10-05 13:03:23');
INSERT OR IGNORE INTO activity_log VALUES(12,1,1,'David Ochieng','task','Marked “Test task from API” as Completed','2026-10-05 13:04:15');
INSERT OR IGNORE INTO activity_log VALUES(13,1,1,'David Ochieng','site-update','Submitted site update for Lakeview Residence — 2026-10-05','2026-10-05 13:04:15');
INSERT OR IGNORE INTO activity_log VALUES(14,1,1,'David Ochieng','task','Created task “SMOKE 1791207993948”','2026-10-05 13:46:33');
INSERT OR IGNORE INTO activity_log VALUES(15,1,1,'David Ochieng','task','Marked “SMOKE 1791207993948” as In progress','2026-10-05 13:46:34');
INSERT OR IGNORE INTO activity_log VALUES(16,1,1,'David Ochieng','task','Marked “SMOKE 1791207993948” as Completed','2026-10-05 13:46:34');
INSERT OR IGNORE INTO activity_log VALUES(17,1,1,'David Ochieng','site-update','Submitted site update for Lakeview Residence — 2026-10-05','2026-10-05 13:46:34');
INSERT OR IGNORE INTO activity_log VALUES(18,1,1,'David Ochieng','task','Created task “SMOKE 1791209269153”','2026-10-05 14:07:49');
INSERT OR IGNORE INTO activity_log VALUES(19,1,1,'David Ochieng','task','Marked “SMOKE 1791209269153” as In progress','2026-10-05 14:07:49');
INSERT OR IGNORE INTO activity_log VALUES(20,1,1,'David Ochieng','task','Marked “SMOKE 1791209269153” as Completed','2026-10-05 14:07:49');
INSERT OR IGNORE INTO activity_log VALUES(21,1,1,'David Ochieng','site-update','Submitted site update for Lakeview Residence — 2026-10-05','2026-10-05 14:07:49');
INSERT OR IGNORE INTO activity_log VALUES(22,1,1,'David Ochieng','task','Created task “SMOKE 1791209514325”','2026-10-05 14:11:54');
INSERT OR IGNORE INTO activity_log VALUES(23,1,1,'David Ochieng','task','Marked “SMOKE 1791209514325” as In progress','2026-10-05 14:11:54');
INSERT OR IGNORE INTO activity_log VALUES(24,1,1,'David Ochieng','task','Marked “SMOKE 1791209514325” as Completed','2026-10-05 14:11:54');
INSERT OR IGNORE INTO activity_log VALUES(25,1,1,'David Ochieng','site-update','Submitted site update for Lakeview Residence — 2026-10-05','2026-10-05 14:11:54');
-- user_preferences
INSERT OR IGNORE INTO user_preferences VALUES(1,'{"document-uploads":true}');

-- Keep AUTOINCREMENT counters ahead of seeded ids
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM users) WHERE name = 'users';
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM projects) WHERE name = 'projects';
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM suppliers) WHERE name = 'suppliers';
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM materials) WHERE name = 'materials';
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM material_deliveries) WHERE name = 'material_deliveries';
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM tasks) WHERE name = 'tasks';
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM documents) WHERE name = 'documents';
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM site_updates) WHERE name = 'site_updates';
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM site_update_materials) WHERE name = 'site_update_materials';
UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM activity_log) WHERE name = 'activity_log';
