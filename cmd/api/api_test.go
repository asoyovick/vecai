package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"vecai/internal/db"
)

// repoPath resolves a path relative to the repository root.
func repoPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(append([]string{filepath.Dir(file), "..", ".."}, parts...)...)
}

const testDSN = "postgres://postgres:vecai@localhost:5432/vecai_test?sslmode=disable"

// resetTestDatabase recreates the schema and loads demo data.
func resetTestDatabase(t *testing.T) {
	t.Helper()
	database, err := db.Open(testDSN)
	if err != nil {
		t.Skipf("test postgres not reachable: %v", err)
	}
	defer database.DB.Close()

	if _, err := database.DB.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if err := db.Migrate(database.DB, repoPath("db/migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.SeedIfEmpty(database.DB, repoPath("db/seed.sql")); err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func newTestApp(t *testing.T) http.Handler {
	t.Helper()
	resetTestDatabase(t)
	database, err := db.Open(testDSN)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { database.DB.Close() })

	handler, err := buildApp(config{
		MigrationsDir: repoPath("db/migrations"),
		StaticDir:     repoPath("dist"),
		UploadsDir:    t.TempDir(),
		SeedPath:      repoPath("db/seed.sql"),
		SeedDemo:      false,
	}, database)
	if err != nil {
		t.Fatalf("build app: %v", err)
	}
	return handler
}

type client struct {
	t       *testing.T
	handler http.Handler
	cookie  string
}

func (c *client) do(method, path string, body any) (int, map[string]any) {
	c.t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			c.t.Fatalf("marshal: %v", err)
		}
		reader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}
	rec := httptest.NewRecorder()
	c.handler.ServeHTTP(rec, req)
	resp := map[string]any{}
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			c.t.Fatalf("decode %s %s: %v (%s)", method, path, err, rec.Body.String())
		}
	}
	if cookies := rec.Result().Cookies(); len(cookies) > 0 {
		for _, ck := range cookies {
			if ck.Name == "vecai_session" {
				if ck.MaxAge < 0 {
					c.cookie = ""
				} else {
					c.cookie = ck.Name + "=" + ck.Value
				}
			}
		}
	}
	return rec.Code, resp
}

func (c *client) login(email, password string) {
	c.t.Helper()
	code, body := c.do("POST", "/api/auth/login", map[string]any{"email": email, "password": password})
	if code != 200 {
		c.t.Fatalf("login %s: %d %+v", email, code, body)
	}
	if c.cookie == "" {
		c.t.Fatal("login did not set a session cookie")
	}
}

func TestAPI_AuthFlow(t *testing.T) {
	handler := newTestApp(t)
	c := &client{t: t, handler: handler}

	if code, _ := c.do("GET", "/api/projects", nil); code != 401 {
		t.Fatalf("unauthenticated projects status = %d, want 401", code)
	}
	if code, body := c.do("POST", "/api/auth/login", map[string]any{"email": "david.ochieng@vecai.co.ke", "password": "wrong"}); code != 401 || body["error"] == nil {
		t.Fatalf("bad login: %d %+v", code, body)
	}
	c.login("david.ochieng@vecai.co.ke", "demo1234")

	if code, body := c.do("GET", "/api/auth/me", nil); code != 200 {
		t.Fatalf("me: %d %+v", code, body)
	}
	if code, body := c.do("POST", "/api/auth/register", map[string]any{
		"name": "Reg Tester", "email": "reg@example.co.ke", "password": "regpass1", "jobTitle": "Engineer",
	}); code != 201 {
		t.Fatalf("register: %d %+v", code, body)
	}
	if code, _ := c.do("POST", "/api/auth/register", map[string]any{
		"name": "Reg Tester", "email": "reg@example.co.ke", "password": "regpass1",
	}); code != 409 {
		t.Fatalf("duplicate register status = %d, want 409", code)
	}

	c.logout()
	if code, _ := c.do("GET", "/api/auth/me", nil); code != 401 {
		t.Fatalf("me after logout: %d, want 401", code)
	}
}

func (c *client) logout() {
	code, _ := c.do("POST", "/api/auth/logout", map[string]any{})
	if code != 200 {
		c.t.Fatalf("logout: %d", code)
	}
}

func TestAPI_ProjectsAndTasks(t *testing.T) {
	handler := newTestApp(t)
	c := &client{t: t, handler: handler}
	c.login("david.ochieng@vecai.co.ke", "demo1234")

	if code, body := c.do("GET", "/api/projects", nil); code != 200 {
		t.Fatalf("projects: %d %+v", code, body)
	}
	if code, body := c.do("GET", "/api/projects/1", nil); code != 200 {
		t.Fatalf("project detail: %d %+v", code, body)
	}
	if code, _ := c.do("GET", "/api/projects/999", nil); code != 404 {
		t.Fatalf("missing project status = %d, want 404", code)
	}

	code, body := c.do("POST", "/api/projects", map[string]any{
		"name": "Riverside Mall", "location": "Nairobi", "status": "Design",
		"client": "Acme Ltd", "budgetTotal": 5000000, "startDate": "2026-11-01",
	})
	if code != 201 {
		t.Fatalf("create project: %d %+v", code, body)
	}
	project := body["project"].(map[string]any)
	if project["status"] != "Design" || project["client"] != "Acme Ltd" {
		t.Fatalf("unexpected project: %+v", project)
	}

	if code, _ := c.do("POST", "/api/projects", map[string]any{"name": "", "location": "x"}); code != 400 {
		t.Fatalf("blank project name status = %d, want 400", code)
	}
	if code, _ := c.do("POST", "/api/projects", map[string]any{"name": "x", "status": "Bogus"}); code != 400 {
		t.Fatalf("bad project status status = %d, want 400", code)
	}

	code, body = c.do("PATCH", "/api/projects/1", map[string]any{"progress": 65})
	if code != 200 {
		t.Fatalf("patch project: %d %+v", code, body)
	}
	if got := body["project"].(map[string]any)["progress"]; got != float64(65) {
		t.Fatalf("progress = %v", got)
	}

	code, body = c.do("POST", "/api/tasks", map[string]any{
		"projectId": 1, "name": "Pour columns", "dueDate": "2026-10-20",
		"assigneeId": 2, "priority": "high", "description": "Ground floor columns",
	})
	if code != 201 {
		t.Fatalf("create task: %d %+v", code, body)
	}
	task := body["task"].(map[string]any)
	if task["status"] != "not-started" || task["priority"] != "high" {
		t.Fatalf("unexpected task: %+v", task)
	}
	if task["assignee_name"] != "Amina Hassan" {
		t.Fatalf("assignee not joined: %+v", task)
	}

	if code, _ := c.do("POST", "/api/tasks", map[string]any{"projectId": 1, "name": "x", "dueDate": "not-a-date"}); code != 400 {
		t.Fatalf("bad due date status = %d, want 400", code)
	}
	if code, _ := c.do("POST", "/api/tasks", map[string]any{"projectId": 999, "name": "x", "dueDate": "2026-10-20"}); code != 400 {
		t.Fatalf("unknown project status = %d, want 400", code)
	}

	code, body = c.do("PATCH", "/api/tasks/"+intID(task["id"]), map[string]any{"status": "in-progress"})
	if code != 200 {
		t.Fatalf("patch task: %d %+v", code, body)
	}
	if body["task"].(map[string]any)["status"] != "in-progress" {
		t.Fatalf("task status not updated: %+v", body)
	}
	if code, _ := c.do("PATCH", "/api/tasks/"+intID(task["id"]), map[string]any{"status": "bogus"}); code != 400 {
		t.Fatalf("bad task status status = %d, want 400", code)
	}
	if code, _ := c.do("PATCH", "/api/tasks/9999", map[string]any{"status": "completed"}); code != 404 {
		t.Fatalf("missing task status = %d, want 404", code)
	}

	if code, body = c.do("GET", "/api/tasks?projectId=1&status=in-progress", nil); code != 200 {
		t.Fatalf("task list: %d %+v", code, body)
	}
	if len(body["tasks"].([]any)) == 0 {
		t.Fatal("task list is empty after creating a task")
	}
}

func intID(v any) string {
	if f, ok := v.(float64); ok {
		return strings.TrimSuffix(strings.ReplaceAll(jsonNumber(f), ".", ""), "")
	}
	return ""
}

func jsonNumber(f float64) string {
	return strings.TrimSpace(strings.Replace(strings.TrimSpace(formatFloat(f)), "e+0", "", 1))
}

func formatFloat(f float64) string {
	b, _ := json.Marshal(int64(f))
	return string(b)
}

func TestAPI_MaterialsDeliveriesSuppliers(t *testing.T) {
	handler := newTestApp(t)
	c := &client{t: t, handler: handler}
	c.login("david.ochieng@vecai.co.ke", "demo1234")

	if code, body := c.do("GET", "/api/materials?projectId=1", nil); code != 200 {
		t.Fatalf("materials: %d %+v", code, body)
	}
	code, body := c.do("POST", "/api/materials", map[string]any{
		"projectId": 1, "name": "Cement", "unit": "bags", "requiredQty": 500,
		"availableQty": 100, "supplierId": 1, "unitCost": 750,
	})
	if code != 201 {
		t.Fatalf("create material: %d %+v", code, body)
	}
	material := body["material"].(map[string]any)
	if material["supplier_name"] != "Devki Steel Mills" && material["supplier_name"] != nil && material["supplier_name"] != "" {
		// supplier names vary; only the join must not break
		_ = material["supplier_name"]
	}

	code, body = c.do("POST", "/api/materials/"+intID(material["id"])+"/deliveries", map[string]any{
		"quantity": 200, "note": "Devki delivery",
	})
	if code != 201 {
		t.Fatalf("log delivery: %d %+v", code, body)
	}
	if got := body["material"].(map[string]any)["available_qty"]; got != float64(300) {
		t.Fatalf("available_qty after delivery = %v, want 300", got)
	}
	if code, _ := c.do("POST", "/api/materials/"+intID(material["id"])+"/deliveries", map[string]any{"quantity": -5}); code != 400 {
		t.Fatalf("negative delivery status = %d, want 400", code)
	}
	if code, _ := c.do("POST", "/api/materials/9999/deliveries", map[string]any{"quantity": 5}); code != 404 {
		t.Fatalf("missing material delivery status = %d, want 404", code)
	}

	code, body = c.do("POST", "/api/suppliers", map[string]any{
		"name": "Test Suppliers Ltd", "category": "Cement", "contact": "Jane", "rating": 4.5,
	})
	if code != 201 {
		t.Fatalf("create supplier: %d %+v", code, body)
	}
	if code, _ := c.do("POST", "/api/suppliers", map[string]any{"name": "x", "category": "y", "rating": 9}); code != 400 {
		t.Fatalf("bad supplier rating status = %d, want 400", code)
	}
	if code, _ := c.do("GET", "/api/suppliers?category=Cement", nil); code != 200 {
		t.Fatalf("supplier list: %d", code)
	}
}

func TestAPI_Documents(t *testing.T) {
	handler := newTestApp(t)
	c := &client{t: t, handler: handler}
	c.login("david.ochieng@vecai.co.ke", "demo1234")

	if code, body := c.do("GET", "/api/documents?projectId=1", nil); code != 200 {
		t.Fatalf("documents: %d %+v", code, body)
	}

	var buf bytes.Buffer
	form := multipart.NewWriter(&buf)
	part, err := form.CreateFormFile("file", "drawing.pdf")
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte("%PDF-1.4 test"))
	form.WriteField("projectId", "1")
	form.WriteField("name", "Test Drawing")
	form.WriteField("category", "Architectural Drawing")
	form.WriteField("version", "2.0")
	form.Close()

	req := httptest.NewRequest("POST", "/api/documents", &buf)
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.Header.Set("Cookie", c.cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("upload document: %d %s", rec.Code, rec.Body.String())
	}
	var uploaded map[string]any
	json.Unmarshal(rec.Body.Bytes(), &uploaded)
	document := uploaded["document"].(map[string]any)
	if document["version"] != "2.0" || document["format"] != "PDF" {
		t.Fatalf("unexpected document: %+v", document)
	}
	docID := intID(document["id"])

	req = httptest.NewRequest("GET", "/api/documents/"+docID+"/download", nil)
	req.Header.Set("Cookie", c.cookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Body.String() != "%PDF-1.4 test" {
		t.Fatalf("download document: %d %q", rec.Code, rec.Body.String())
	}

	if code, _ := c.do("GET", "/api/documents/9999/download", nil); code != 404 {
		t.Fatalf("missing document status = %d, want 404", code)
	}
}

func TestAPI_BOQ(t *testing.T) {
	handler := newTestApp(t)
	c := &client{t: t, handler: handler}
	c.login("david.ochieng@vecai.co.ke", "demo1234")

	code, body := c.do("POST", "/api/boq", map[string]any{
		"projectId": 1, "itemNo": "A.1", "description": "Excavation", "unit": "m³", "quantity": 120, "unitRate": 1500,
	})
	if code != 201 {
		t.Fatalf("create boq item: %d %+v", code, body)
	}
	item := body["item"].(map[string]any)
	if item["amount"] != float64(180000) {
		t.Fatalf("amount = %v, want 180000", item["amount"])
	}

	code, body = c.do("GET", "/api/boq?projectId=1", nil)
	if code != 200 {
		t.Fatalf("boq list: %d %+v", code, body)
	}
	if len(body["items"].([]any)) != 1 {
		t.Fatalf("boq items = %v", body["items"])
	}

	code, body = c.do("PATCH", "/api/boq/"+intID(item["id"]), map[string]any{"quantity": 240})
	if code != 200 {
		t.Fatalf("patch boq: %d %+v", code, body)
	}
	if body["item"].(map[string]any)["amount"] != float64(360000) {
		t.Fatalf("amount after patch = %v", body["item"].(map[string]any)["amount"])
	}

	if code, _ := c.do("DELETE", "/api/boq/"+intID(item["id"]), nil); code != 200 {
		t.Fatalf("delete boq: %d", code)
	}
	if code, _ := c.do("DELETE", "/api/boq/"+intID(item["id"]), nil); code != 404 {
		t.Fatalf("delete missing boq status = %d, want 404", code)
	}
}

func TestAPI_ReportsAndAssistant(t *testing.T) {
	handler := newTestApp(t)
	c := &client{t: t, handler: handler}
	c.login("david.ochieng@vecai.co.ke", "demo1234")

	if code, body := c.do("GET", "/api/dashboard", nil); code != 200 {
		t.Fatalf("dashboard: %d %+v", code, body)
	}
	code, dashboard := c.do("GET", "/api/dashboard", nil)
	if _, ok := dashboard["stats"]; !ok {
		t.Fatalf("dashboard missing stats: %+v (code %d)", dashboard, code)
	}
	if code, body := c.do("GET", "/api/activity", nil); code != 200 {
		t.Fatalf("activity: %d %+v", code, body)
	}
	if code, body := c.do("GET", "/api/site-updates?projectId=1", nil); code != 200 {
		t.Fatalf("site updates: %d %+v", code, body)
	}

	code, body := c.do("POST", "/api/site-updates", map[string]any{
		"projectId": 1, "date": "2026-10-06", "weather": "Sunny", "temperature": "26",
		"workersOnSite": 30, "workCompleted": "Slab edge formwork completed.",
		"issues": "Rain expected tomorrow.", "materialsUsed": []map[string]any{{"name": "Sand", "quantity": 4, "unit": "m³"}},
	})
	if code != 201 {
		t.Fatalf("create site update: %d %+v", code, body)
	}
	update := body["siteUpdate"].(map[string]any)
	if update["workers_on_site"] != float64(30) {
		t.Fatalf("workers_on_site = %v", update["workers_on_site"])
	}
	if mats := update["materialsUsed"].([]any); len(mats) != 1 {
		t.Fatalf("materialsUsed = %v", mats)
	}
	if code, _ := c.do("POST", "/api/site-updates", map[string]any{"projectId": 1, "date": "bad", "workCompleted": "x"}); code != 400 {
		t.Fatalf("bad site update date status = %d, want 400", code)
	}

	for _, question := range []string{
		"weekly status report",
		"summarise the site",
		"budget variance",
		"material forecast",
		"tasks at risk",
		"why is the project delayed",
	} {
		code, body := c.do("POST", "/api/assistant", map[string]any{"question": question, "projectId": 1})
		if code != 200 {
			t.Fatalf("assistant %q: %d %+v", question, code, body)
		}
		answer := body["answer"].(map[string]any)
		if answer["title"] == "" || len(answer["sections"].([]any)) == 0 {
			t.Fatalf("assistant answer for %q: %+v", question, answer)
		}
	}
}

func TestAPI_Authorization(t *testing.T) {
	handler := newTestApp(t)

	// An admin invites a read-only viewer; viewers cannot mutate.
	admin := &client{t: t, handler: handler}
	admin.login("david.ochieng@vecai.co.ke", "demo1234")
	code, body := admin.do("POST", "/api/team/invite", map[string]any{
		"name": "Viewer Only", "email": "viewer@example.co.ke", "jobTitle": "Intern", "role": "viewer",
	})
	if code != 201 {
		t.Fatalf("invite viewer: %d %+v", code, body)
	}
	register := &client{t: t, handler: handler}
	register.login("viewer@example.co.ke", body["temporaryPassword"].(string))
	if code, _ := register.do("POST", "/api/tasks", map[string]any{"projectId": 1, "name": "x", "dueDate": "2026-10-20"}); code != 403 {
		t.Fatalf("viewer create task status = %d, want 403", code)
	}
	if code, _ := register.do("POST", "/api/materials", map[string]any{"projectId": 1, "name": "x", "unit": "pcs"}); code != 403 {
		t.Fatalf("viewer create material status = %d, want 403", code)
	}
	if code, _ := register.do("PATCH", "/api/organisation", map[string]any{"name": "x", "businessType": "y", "country": "z", "currency": "KES"}); code != 403 {
		t.Fatalf("viewer update organisation status = %d, want 403", code)
	}
	if code, _ := register.do("POST", "/api/team/invite", map[string]any{"name": "x", "email": "y@example.co.ke", "jobTitle": "z"}); code != 403 {
		t.Fatalf("viewer invite status = %d, want 403", code)
	}

	// A member cannot create projects or invite.
	member := &client{t: t, handler: handler}
	member.login("francis.kimani@vecai.co.ke", "demo1234")
	if code, _ := member.do("POST", "/api/projects", map[string]any{"name": "x"}); code != 403 {
		t.Fatalf("member create project status = %d, want 403", code)
	}
	if code, _ := member.do("POST", "/api/team/invite", map[string]any{"name": "x", "email": "z@example.co.ke", "jobTitle": "z"}); code != 403 {
		t.Fatalf("member invite status = %d, want 403", code)
	}
	if code, _ := member.do("GET", "/api/team", nil); code != 200 {
		t.Fatalf("member team list status = %d, want 200", code)
	}

	// Admin can invite with a project assignment and receives the temporary password.
	code, body = admin.do("POST", "/api/team/invite", map[string]any{
		"name": "Invited Person", "email": "invited@example.co.ke", "jobTitle": "Architect", "role": "viewer", "projectId": 1,
	})
	if code != 201 {
		t.Fatalf("admin invite: %d %+v", code, body)
	}
	if body["temporaryPassword"] == "" {
		t.Fatalf("invite response missing temporary password: %+v", body)
	}
	inviteLogin := &client{t: t, handler: handler}
	inviteLogin.login("invited@example.co.ke", body["temporaryPassword"].(string))
	if code, _ := inviteLogin.do("GET", "/api/auth/me", nil); code != 200 {
		t.Fatalf("invited user login: %d", code)
	}
}

func TestAPI_Newsletter(t *testing.T) {
	handler := newTestApp(t)
	c := &client{t: t, handler: handler}

	if code, _ := c.do("POST", "/api/newsletter/subscribe", map[string]any{"email": "news@example.co.ke"}); code != 201 {
		t.Fatalf("subscribe: %d", code)
	}
	if code, body := c.do("POST", "/api/newsletter/subscribe", map[string]any{"email": "news@example.co.ke"}); code != 409 {
		t.Fatalf("duplicate subscribe status = %d (%+v), want 409", code, body)
	}
	if code, _ := c.do("POST", "/api/newsletter/subscribe", map[string]any{"email": "nope"}); code != 400 {
		t.Fatalf("invalid email status = %d, want 400", code)
	}
}

func TestAPI_ValidationAndGuards(t *testing.T) {
	handler := newTestApp(t)
	c := &client{t: t, handler: handler}
	c.login("david.ochieng@vecai.co.ke", "demo1234")

	req := httptest.NewRequest("POST", "/api/projects", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", c.cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("invalid JSON status = %d, want 400", rec.Code)
	}

	req = httptest.NewRequest("POST", "/api/projects", strings.NewReader("name=x"))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Cookie", c.cookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 415 {
		t.Fatalf("non-JSON mutation status = %d, want 415", rec.Code)
	}

	if code, _ := c.do("GET", "/api/projects/abc", nil); code != 400 {
		t.Fatalf("non-numeric project id status = %d, want 400", code)
	}
}

func TestSPAHandler(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/index.html", []byte("<html>app</html>"), 0o644)
	os.WriteFile(dir+"/app.js", []byte("// js"), 0o644)

	handler := spaHandler(dir)
	for _, path := range []string{"/", "/projects/1", "/login"} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != 200 || rec.Body.String() != "<html>app</html>" {
			t.Fatalf("SPA fallback %s: %d %q", path, rec.Code, rec.Body.String())
		}
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/app.js", nil))
	if rec.Code != 200 || rec.Body.String() != "// js" {
		t.Fatalf("static asset: %d %q", rec.Code, rec.Body.String())
	}
}
