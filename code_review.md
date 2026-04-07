# Code Review — Issues Found via Context7  Docs

Cross-referenced the cozybox codebase against Gin, GORM, and golang-jwt official documentation. Below are all issues found, ordered by severity.

---

## 🔴 CRITICAL — Security & Crash Risks

### 1. Password Exposed in JSON Serialization

**File:** [user_domain.go](file:///d:/Development/project/cozybox/internal/models/user_domain.go#L13)

```go
Password string `json:"password" gorm:"column:password"`  // ← LEAKS IN ANY JSON RESPONSE
```

If a `User` model is ever accidentally serialized to JSON (e.g. in a debug log, an error response, or a future endpoint), the password hash is included. The DTO layer currently strips it, but this is one mistake away from exposure.

**Fix:** Change the JSON tag to `json:"-"` to ensure the password is **never** serialized:
```go
Password string `json:"-" gorm:"column:password"`
```

---

### 2. Unsafe Context Value Type Assertions — Will Panic

**Files:** [user_service.go](file:///d:/Development/project/cozybox/internal/service/user_service.go#L401) (lines 401, 414, 447, 480, 531, 577)

```go
id := ctx.Value(contextkey.UserID).(string)  // ← PANICS if value is nil
```

If the auth middleware is bypassed, skipped, or the context key is missing for any reason, this is a **runtime panic** that crashes the goroutine. Per Go best practices, always use the comma-ok assertion:

**Fix:**
```go
id, ok := ctx.Value(contextkey.UserID).(string)
if !ok || id == "" {
    return nil, errors.New("unauthorized: user context missing")
}
```

This pattern appears **6 times** in `user_service.go`.

---

### 3. Shutdown Context Created Too Early — Graceful Shutdown Is Broken

**File:** [app.go](file:///d:/Development/project/cozybox/internal/core/app.go#L89-L90)

```go
ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout*time.Second)  // 5s timeout starts HERE
defer cancel()

// ... server runs indefinitely ...

case <-quit:
    srv.Shutdown(ctx)  // ← ctx already expired long ago!
```

The 5-second shutdown context is created **before the server starts**. By the time a SIGTERM arrives (minutes/hours later), the context has long expired. `Shutdown(ctx)` will immediately fail.

**Fix:** Create the context **inside** the quit handler:
```go
case <-quit:
    ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil { ... }
```

---

## 🟡 HIGH — Data Integrity & Framework Misuse

### 4. Using `sql.NullTime` Instead of `gorm.DeletedAt` — GORM Soft Delete Is Broken

**Files:** All models with `DeletedAt`

Per Context7/GORM docs, GORM's automatic soft delete filtering **only works** with `gorm.DeletedAt`, not `sql.NullTime`:

```go
// Current (BROKEN auto-filtering):
DeletedAt sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`

// Correct (GORM handles everything):
DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at"`
```

With `gorm.DeletedAt`:
- `db.Find(&users)` → auto-adds `WHERE deleted_at IS NULL`
- `db.Delete(&user)` → auto-sets `deleted_at` instead of hard deleting
- `db.Unscoped().Find(&users)` → includes soft-deleted records

**Impact:** Affects `User`, `Organization`, `Document`, `Contact`, `Tenant` models.

---

### 5. Manual `deleted_at IS NULL` Everywhere — Fragile & Error-Prone

**File:** [user_service.go](file:///d:/Development/project/cozybox/internal/service/user_service.go) (throughout)

Because of issue #4, every query manually adds `AND deleted_at IS NULL`:

```go
tx.Where("email = ? AND deleted_at IS NULL", req.Email).First(&userModel)
tx.Where("id = ? AND deleted_at IS NULL", id).First(&userModel)
```

This is repeated **11 times**. If `gorm.DeletedAt` is used, all these manual clauses become unnecessary — GORM adds them automatically.

---

### 6. `FirstOrCreate` Race Condition in Registration

**File:** [user_service.go](file:///d:/Development/project/cozybox/internal/service/user_service.go#L66)

```go
res := tx.Where("email = ? AND deleted_at IS NULL", req.Email).FirstOrCreate(request)
```

Per GORM docs, `FirstOrCreate` is subject to race conditions without a database-level unique constraint. If two registrations for the same email arrive simultaneously, both `FirstOrCreate` calls can pass the `WHERE` check before either inserts. Wrapping in a transaction helps, but doesn't fully prevent this.

**Fix:** Ensure the `users` table has a `UNIQUE` index on `email`:
```sql
CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
```

---

### 7. Unnecessary Transactions for Read-Only Queries

**File:** [user_service.go](file:///d:/Development/project/cozybox/internal/service/user_service.go#L207-L208)

```go
// Login — starts a transaction for reads + one small update
tx := u.db.WithContext(ctx).Begin()
defer tx.Rollback()
```

`Login` opens a transaction, reads the user, and only updates `LastLogin`. The read portion doesn't need a transaction. Similarly `VerifyEmail` and `GetProfile` patterns. Transactions hold database connections longer than necessary.

**Better:** Use `db.WithContext(ctx)` for reads, only wrap the write in a transaction if needed. Or use `db.Save()` / `db.Model().Update()` for single-field updates without a full transaction.

---

### 8. Goroutine Uses Gin/Service Context Without `c.Copy()`

**File:** [user_service.go](file:///d:/Development/project/cozybox/internal/service/user_service.go#L103-L107)

```go
go func() {
    if err := u.mailer.SendOTP(request.Email, "Verify Your Email", fmt.Sprintf("...%s", otp)); err != nil {
        fmt.Printf("failed to send otp to %s: %v", request.Email, err)
    }
}()
```

Per Gin docs (Context7), goroutines spawned from a request handler must use `c.Copy()` to avoid race conditions. While this particular goroutine only uses local variables (not the context directly), it's called from within a request scope. The bigger problem: it captures `request.Email` which is fine, but uses `fmt.Printf` instead of the logger.

Same pattern on line 521.

**Fix:** Use the logger, and if you ever need context in these goroutines, pass `c.Copy()`.

---

## 🟠 MEDIUM — Code Quality

### 9. Duplicate Context Key Type Definitions

**Files:**
- [contextkey/contextkey.go](file:///d:/Development/project/cozybox/internal/tools/contextkey/contextkey.go) — defines `ContextKey`
- [helper/context.go](file:///d:/Development/project/cozybox/internal/tools/helper/context.go#L10-L15) — defines its own `ContextKey`

Both packages define `type ContextKey string` with `RequestID` and `UserID` constants. This means values set with `contextkey.UserID` can't be retrieved with `helper.UserID` — they're different types, so `context.Value()` won't match.

**Fix:** Delete the duplicate from `helper/context.go` and import from `contextkey` package.

---

### 10. Wrong Error Message in Non-TLS Server Path

**File:** [app.go](file:///d:/Development/project/cozybox/internal/core/app.go#L125)

```go
// In the non-SSL else branch:
case err := <-serverError:
    log.Fatalf("Failed to start TLS server: %v", err)  // ← says "TLS" but this is non-TLS
```

Copy-paste error — the non-TLS path uses a TLS error message.

---

### 11. CORS Allows Wildcard Origin with Credentials

**File:** [cors_middleware.go](file:///d:/Development/project/cozybox/internal/middleware/cors_middleware.go#L11-L12)

```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
```

Browsers reject `Access-Control-Allow-Credentials: true` when `Access-Control-Allow-Origin: *`. This combination is spec-invalid. Cookies/auth headers won't be sent by the browser.

**Fix:** Either remove credentials, or reflect the specific `Origin` header instead of wildcard:
```go
origin := c.Request.Header.Get("Origin")
if origin != "" {
    c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
}
```

Also add `Authorization` to `Access-Control-Allow-Headers` — it's currently missing, so Bearer token requests will fail CORS preflight.

---

### 12. Redundant RowsAffected Check After `First()`

**File:** [user_service.go](file:///d:/Development/project/cozybox/internal/service/user_service.go#L179-L186)

```go
res := tx.Where("id = ? AND deleted_at IS NULL", user.Id).First(&userModel)
if res.Error != nil {
    return nil, fmt.Errorf("failed to get user: %w", res.Error)
}

if res.RowsAffected == 0 {  // ← redundant — First() already returns ErrRecordNotFound
    return nil, errors.New("user not found")
}
```

GORM's `First()` returns `gorm.ErrRecordNotFound` when no rows match. The `RowsAffected == 0` check after `First()` is unreachable — if no rows are found, the error check above already catches it.

---

## 🟢 LOW — Naming & Style

### 13. Filename Typo: `identificaation_gen.go`

**File:** [identificaation_gen.go](file:///d:/Development/project/cozybox/internal/tools/util/identificaation_gen.go)

Should be `identification_gen.go`.

---

### 14. Context Helper Defines Unused Duplicate Keys

**File:** [helper/context.go](file:///d:/Development/project/cozybox/internal/tools/helper/context.go#L12-L15)

```go
const (
    RequestID ContextKey = "request_id"
    UserID    ContextKey = "user_id"
)
```

These shadow the `contextkey` package and aren't used consistently. Remove them.

---

## Summary

| # | Severity | Issue | Effort |
|---|---|---|---|
| 1 | 🔴 Critical | Password in JSON tag | Trivial |
| 2 | 🔴 Critical | Unsafe context type assertions (6 places) | Small |
| 3 | 🔴 Critical | Shutdown context expires immediately | Small |
| 4 | 🟡 High | `sql.NullTime` instead of `gorm.DeletedAt` | Medium |
| 5 | 🟡 High | Manual `deleted_at IS NULL` x11 | Medium (auto-fixed by #4) |
| 6 | 🟡 High | `FirstOrCreate` race condition | Small (add unique index) |
| 7 | 🟡 High | Unnecessary transactions for reads | Small |
| 8 | 🟡 High | Goroutine context safety | Small |
| 9 | 🟠 Medium | Duplicate context key types | Trivial |
| 10 | 🟠 Medium | Wrong "TLS" error message | Trivial |
| 11 | 🟠 Medium | CORS wildcard + credentials invalid | Small |
| 12 | 🟠 Medium | Redundant RowsAffected check | Trivial |
| 13 | 🟢 Low | Filename typo | Trivial |
| 14 | 🟢 Low | Unused duplicate constants | Trivial |
