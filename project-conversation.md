# Business Document Generation System — Project Architecture Chat

---

## Document Types & Scope

**Document types:** Quotes, Invoice, Debit Note, Purchase Order, Sales Order, Receipt

**Users:** External clients/customers — multi-tenant SaaS product

**Output format:** PDF only

**Deployment:** Multi-tenant SaaS

---

## System Architecture Overview

### Tech stack
- Backend: Go (intermediate team, 1–3 years experience)
- Frontend: React SPA (Vite)
- Database: PostgreSQL (row-level multi-tenancy with `tenant_id`)
- Cache/Sequences: Redis
- File storage: S3
- PDF renderer: HTML → chromedp (headless Chrome)

### Multi-tenancy strategy
Row-level isolation — single PostgreSQL database with `tenant_id` on every table. Schema-per-tenant is overengineering at < 50 tenants.

### Go project structure
```
/cmd/server
/internal
  /tenant            → tenant CRUD, branding config
  /document          → all 6 doc types, lifecycle, versioning
  /pdf               → HTML template rendering + PDF conversion
  /auth              → OAuth/SAML middleware, session handling
  /billing           → Stripe integration
  /email             → delivery via SendGrid/SES
  /storage           → S3 upload/download for archived PDFs
/web                 → React SPA (Vite build output served by Go)
/migrations          → SQL migration files (golang-migrate)
```

### Must-have integrations at launch
- Authentication / SSO (OAuth, SAML)
- Payment / billing (Stripe)
- Email delivery (SendGrid / SES)

### Success metrics (6 months)
- Fast generation time (< 3s per doc)
- Zero manual formatting errors

---

## PDF Generation: HTML → PDF vs Maroto

### Why Maroto hits a ceiling
Maroto is a Go-native library using a grid/cell model — every layout decision is baked into Go code. When tenants need to customize their quotation (move logo, change columns, add footer), there's no path forward without writing custom Go per tenant.

**The ceiling is permanent.** Maroto has no concept of "templates a non-developer can edit."

### Why HTML → PDF scales to customization

The Go layer stays clean — it injects data into a template and calls the renderer:

```go
type QuoteTemplateData struct {
    Tenant   TenantBranding
    Document QuoteDocument
    Settings TemplateSettings
}
tmpl.Execute(w, data)
// → chromedp converts the output HTML to PDF
```

Template stored in database per tenant. Branding config as JSON:

```json
{
  "primaryColor": "#1a56db",
  "logoUrl": "https://cdn.../logo.png",
  "logoPosition": "top-right",
  "showPaymentTerms": true,
  "footerNote": "Thank you for your business.",
  "font": "Inter"
}
```

### Customization roadmap
- **Phase 1 (launch):** Branding via settings form — logo, colors, footer text. CSS variables injected at render time.
- **Phase 2:** Live template preview in React — tenant sees real quotation as they adjust settings.
- **Phase 3:** Block-based template editor. Toggle sections, reorder blocks, pick layout variants. Saves `template_config` JSON per tenant.

### Managing chromedp performance
Keep 2–3 warm Chrome processes alive in Go:

```go
var browserCtx context.Context  // initialized once at startup

func GeneratePDF(html string) ([]byte, error) {
    ctx, cancel := chromedp.NewContext(browserCtx)
    defer cancel()
    // ... render and capture PDF
}
```

---

## Document Lifecycle

**States:** `draft → sent → paid → void`

**Versioning:** Append-only `document_versions` table — never mutate in place. Each save creates a new version row with `is_current` flag on latest.

**Document numbering:** Date-scoped format — `INV/2026/04/0001`

Use Redis atomic counters per `(tenant_id, doc_type, year, month)`:
```
seq:tenant123:INV:2026:04
```

---

## Key Data Model Decisions

| Decision | Recommendation |
|---|---|
| Tenancy | `tenant_id UUID` FK on all tables |
| Branding | `tenant_branding` table: logo URL (S3), primary color, font preference |
| Doc numbering | `document_sequences` table + Redis counter; resets per month |
| Versioning | Append-only `document_versions`; `is_current` flag on latest |
| Status | `status ENUM` on `documents` table with transition validation in Go |
| PDF storage | Generated PDFs stored in S3; `documents` table holds the S3 key |

---

## Google & Microsoft Calendar Integration

### Architecture
Define a `CalendarProvider` interface before writing any provider-specific code:

```go
type CalendarProvider interface {
    CreateEvent(ctx context.Context, token *OAuthToken, event CalendarEvent) (string, error)
    RefreshToken(ctx context.Context, token *OAuthToken) (*OAuthToken, error)
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*OAuthToken, error)
}

type CalendarEvent struct {
    Title          string
    Description    string
    StartTime      time.Time
    EndTime        time.Time
    AttachmentURL  string
    AttachmentName string
}
```

### OAuth Scopes

**Google:**
```
https://www.googleapis.com/auth/calendar.events
https://www.googleapis.com/auth/drive.file
```

**Microsoft Graph:**
```
Calendars.ReadWrite
Files.ReadWrite
offline_access    ← required for refresh token
```

### Token storage
```sql
CREATE TABLE calendar_connections (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL REFERENCES tenants(id),
    user_id       UUID NOT NULL REFERENCES users(id),
    provider      VARCHAR(20) NOT NULL,  -- 'google' | 'microsoft'
    access_token  BYTEA NOT NULL,        -- AES-256 encrypted
    refresh_token BYTEA NOT NULL,        -- AES-256 encrypted
    token_expiry  TIMESTAMPTZ NOT NULL,
    scopes        TEXT NOT NULL,
    connected_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, provider)
);
```

### PDF attachment strategy
Use presigned S3 URL in the event description — no Drive/OneDrive upload needed at Phase 1:

```
Invoice INV/2026/04/0012 — Due 30 April 2026
Amount: Rp 5.400.000

Download PDF: https://your-cdn.com/docs/inv-0012.pdf?token=...
```

### Calendar event triggers

| Document type | Trigger | Calendar event meaning |
|---|---|---|
| Quote | Status → `sent` | Follow-up reminder on quote expiry |
| Invoice | Status → `sent` | Payment due date reminder |
| Purchase Order | Status → `sent` | Expected delivery date |
| Sales Order | Status → `sent` | Fulfillment deadline |
| Debit Note | Status → `sent` | Payment expected date |
| Receipt | Status → `paid` | Record of payment |

### Implementation stages
1. OAuth connect/disconnect flow for Google. Store tokens. Show "connected" in UI.
2. Event creation with PDF link on document `sent`. Then add Microsoft.
3. Native Drive/OneDrive attachment if demand exists.

---

## Document Customization Canvas

### Decisions
- **Customization level:** Full free-form drag anywhere on A4 canvas
- **Draggable elements:** Logo & company header, Line items table, Signature / stamp area
- **Grid snapping:** Snap to grid only (strict)
- **Layout:** Per document type (invoice has its own, quote has its own)
- **Pipeline:** Canvas layout saved as JSON → Go renders HTML → PDF
- **Styling controls:** Font family & size, Text color & background color, Border style & radius, Padding & margin control, Show/hide on certain doc types
- **Preview:** Button-triggered only (not real-time)
- **Multi-page:** Canvas = one A4 page; line items table auto-expands at render time
- **State:** Draft and Published (no undo — Reset to Default button)

### Coordinate system
Store all block positions in grid units, not pixels. With 8px dot grid:
- A4 at 96dpi = 794 × 1123px = 99 × 140 grid units
- On drag end: `Math.round(px / gridSize)` to snap

### Layout JSON schema
```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "doc_type": "invoice",
  "state": "draft",
  "grid_size": 8,
  "canvas_width": 99,
  "canvas_height": 140,
  "blocks": [
    {
      "id": "block-1",
      "type": "logo_header",
      "x": 2, "y": 2,
      "w": 95, "h": 12,
      "visible": true,
      "styles": {
        "fontFamily": "Inter",
        "fontSize": "14px",
        "color": "#1a1a1a",
        "backgroundColor": "#ffffff",
        "borderStyle": "none",
        "borderRadius": "0px",
        "padding": "8px"
      }
    }
  ]
}
```

### Database schema
```sql
CREATE TABLE document_layouts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID,              -- NULL = system default
    doc_type      VARCHAR(30) NOT NULL,
    state         VARCHAR(10) NOT NULL,  -- 'draft' | 'published'
    grid_size     INT DEFAULT 8,
    canvas_width  INT DEFAULT 99,
    canvas_height INT DEFAULT 140,
    UNIQUE (tenant_id, doc_type, state)
);

CREATE TABLE layout_blocks (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    layout_id  UUID NOT NULL REFERENCES document_layouts(id),
    block_type VARCHAR(30) NOT NULL,  -- 'logo_header' | 'line_items_table' | 'signature_stamp'
    x          INT NOT NULL,
    y          INT NOT NULL,
    w          INT NOT NULL,
    h          INT NOT NULL,
    visible    BOOLEAN DEFAULT TRUE,
    styles     JSONB NOT NULL
);
```

### Draft → Published state machine
```
draft ──[publish]──▶ published
  ▲                      │
  └────[reset]───────────┘
       (copies system default into draft)
```

```sql
-- Publish
UPDATE document_layouts
SET blocks = (SELECT blocks FROM document_layouts
              WHERE tenant_id = $1 AND doc_type = $2 AND state = 'draft')
WHERE tenant_id = $1 AND doc_type = $2 AND state = 'published';

-- Reset to default
UPDATE document_layouts
SET blocks = (SELECT blocks FROM document_layouts
              WHERE tenant_id IS NULL AND doc_type = $2)
WHERE tenant_id = $1 AND doc_type = $2 AND state = 'draft';
```

### Implementation milestones
1. Canvas shell: A4 + dot grid + 3 draggable block placeholders + save JSON as draft
2. Preview: Go endpoint renders HTML from JSON, returns it, React shows in iframe
3. Styling panel: right sidebar with font/color/border/padding controls per block
4. Publish flow: publish button, reset to default, PDF pipeline reads published layout
5. Per document type: replicate for Quote, PO, SO, Debit Note, Receipt

---

## Canvas Library Recommendation

### Recommended: `@dnd-kit/core` + `re-resizable`

```bash
npm install @dnd-kit/core @dnd-kit/modifiers re-resizable
```

**~18 kb total. No extra dependencies.**

#### Snap to grid (built-in modifier)
```tsx
import { DndContext } from '@dnd-kit/core'
import { restrictToParentElement, createSnapModifier } from '@dnd-kit/modifiers'

const GRID_SIZE = 8
const snapToGrid = createSnapModifier(GRID_SIZE)

<DndContext
  modifiers={[snapToGrid, restrictToParentElement]}
  onDragEnd={handleDragEnd}
>
  {blocks.map(block => <DraggableBlock key={block.id} block={block} />)}
</DndContext>
```

#### Resize with snap (re-resizable)
```tsx
import { Resizable } from 're-resizable'

<Resizable
  size={{ width: block.w * GRID_SIZE, height: block.h * GRID_SIZE }}
  snap={{ x: makeGridSteps(GRID_SIZE), y: makeGridSteps(GRID_SIZE) }}
  bounds="parent"
  onResizeStop={(e, dir, ref, delta) => {
    updateBlock(block.id, {
      w: Math.round(ref.offsetWidth  / GRID_SIZE),
      h: Math.round(ref.offsetHeight / GRID_SIZE),
    })
  }}
>
  <BlockContent block={block} />
</Resizable>
```

#### Drag end handler (grid unit conversion)
```ts
const handleDragEnd = (event: DragEndEvent) => {
  const { active, delta } = event
  const block = blocks.find(b => b.id === active.id)
  if (!block) return

  const newX = Math.round((block.x * GRID_SIZE + delta.x) / GRID_SIZE)
  const newY = Math.round((block.y * GRID_SIZE + delta.y) / GRID_SIZE)

  updateBlock(block.id, {
    x: Math.max(0, Math.min(newX, CANVAS_COLS - block.w)),
    y: Math.max(0, Math.min(newY, CANVAS_ROWS - block.h)),
  })
}
```

#### A4 canvas CSS (dot grid)
```css
.a4-canvas {
  position: relative;
  width: 794px;
  height: 1123px;
  transform-origin: top left;
  transform: scale(var(--canvas-scale));
  background-color: #fff;
  background-image: radial-gradient(circle, #c0c0c0 1px, transparent 1px);
  background-size: 8px 8px;
}
```

### Why not the alternatives
- `react-draggable` — low maintenance activity
- `react-grid-layout` — column-flow only, prevents block overlap
- `Konva.js` — renders to HTML Canvas, breaks the HTML→PDF pipeline
- `tldraw` — overkill, > 500kb, hard to constrain to A4

### React component structure
```
/components/canvas/
  CanvasEditor.tsx
  DraggableBlock.tsx
  blocks/
    LogoHeaderBlock.tsx
    LineItemsBlock.tsx
    SignatureBlock.tsx
  sidebar/
    StylePanel.tsx
    BlockVisibility.tsx
  hooks/
    useLayout.ts
    useGridSnap.ts
```

---

## HTML Preview in React

### How it works
Go returns a full HTML string. React puts it inside an `<iframe>` using `srcDoc`. The iframe creates an isolated document scope — the document's CSS can't bleed into the React app's styles.

### React component
```tsx
const A4_WIDTH_PX  = 794
const A4_HEIGHT_PX = 1123

export function PreviewPanel({ layout, sampleDocumentId }: Props) {
  const [htmlString, setHtmlString] = useState<string | null>(null)
  const [loading, setLoading]       = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const [scale, setScale]           = useState(1)

  useEffect(() => {
    if (!containerRef.current) return
    setScale(containerRef.current.offsetWidth / A4_WIDTH_PX)
  }, [])

  const handlePreview = async () => {
    setLoading(true)
    const res = await fetch('/api/templates/preview', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ layout, document_id: sampleDocumentId ?? null }),
    })
    const html = await res.text()
    setHtmlString(html)
    setLoading(false)
  }

  return (
    <div>
      <button onClick={handlePreview} disabled={loading}>
        {loading ? 'Generating preview…' : 'Preview PDF'}
      </button>

      <div ref={containerRef}
        style={{ width: '100%', height: A4_HEIGHT_PX * scale, overflow: 'hidden' }}>
        {htmlString && (
          <iframe
            srcDoc={htmlString}
            sandbox="allow-same-origin"
            style={{
              width: A4_WIDTH_PX, height: A4_HEIGHT_PX,
              border: 'none', transformOrigin: 'top left',
              transform: `scale(${scale})`,
            }}
          />
        )}
      </div>
    </div>
  )
}
```

### Key rules
- Always use `srcDoc` — never `dangerouslySetInnerHTML`
- `sandbox="allow-same-origin"` — allows fonts from CDN, blocks scripts in preview
- Scale the iframe with CSS `transform: scale()` — never resize the iframe itself
- Go returns `text/html`, not JSON

---

## chromedp: HTML → PDF

### How it works
chromedp is not a PDF library. It's a Go API that controls a real headless Chrome browser. You tell it to load your HTML, wait for it to fully render, then call Chrome's built-in "Print to PDF". Chrome does all the PDF work.

### Production PDF service with warm pool

```go
func NewService(poolSize int) (*Service, error) {
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
        chromedp.Flag("no-sandbox", true),             // required in Docker
        chromedp.Flag("disable-dev-shm-usage", true),  // required in Docker
        chromedp.Flag("disable-extensions", true),
        chromedp.WindowSize(1280, 1024),
    )

    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

    svc := &Service{
        parentCtx: allocCtx,
        cancel:    cancel,
        pool:      make(chan context.Context, poolSize),
    }

    for i := 0; i < poolSize; i++ {
        tabCtx, _ := chromedp.NewContext(allocCtx)
        chromedp.Run(tabCtx, chromedp.Navigate("about:blank"))
        svc.pool <- tabCtx
    }

    return svc, nil
}

func (s *Service) GeneratePDF(htmlContent string) ([]byte, error) {
    tabCtx := <-s.pool
    defer func() { s.pool <- tabCtx }()

    ctx, cancel := context.WithTimeout(tabCtx, 10*time.Second)
    defer cancel()

    encoded := base64.StdEncoding.EncodeToString([]byte(htmlContent))
    dataURL := "data:text/html;base64," + encoded

    var pdfBytes []byte

    err := chromedp.Run(ctx,
        chromedp.Navigate(dataURL),
        chromedp.WaitReady("body", chromedp.ByQuery),
        chromedp.ActionFunc(func(ctx context.Context) error {
            var err error
            pdfBytes, _, err = page.PrintToPDF().
                WithPrintBackground(true).
                WithPaperWidth(8.27).
                WithPaperHeight(11.69).
                WithMarginTop(0).
                WithMarginBottom(0).
                WithMarginLeft(0).
                WithMarginRight(0).
                WithPreferCSSPageSize(true).
                Do(ctx)
            return err
        }),
    )

    return pdfBytes, err
}
```

### Multi-page CSS rules
```css
.line-items-table tr   { page-break-inside: avoid; }
.totals-block          { page-break-inside: avoid; }
.signature-block       { page-break-inside: avoid; }
```

`WithPreferCSSPageSize(true)` is required for Chrome to respect these rules.

### Docker requirements
```yaml
services:
  app:
    shm_size: '256mb'  # Chrome crashes without this
```

```dockerfile
RUN apt-get install -y chromium fonts-liberation
ENV CHROMIUM_PATH=/usr/bin/chromium
```

### Timing expectations (warm pool, 3 contexts)
| Step | Time |
|---|---|
| Borrow context from pool | ~0ms |
| Navigate to data URL | ~80–150ms |
| Render wait | ~100–200ms |
| PrintToPDF call | ~200–400ms |
| **Total** | **~400–750ms** |

Well inside the 3s target.

---

## PDF Storage: S3 + Database Metadata

### Decision: S3 file + Postgres metadata (not client-side, not DB blob)

**Why not client-side PDF generation:**
- `jsPDF` can't render CSS — approximates styles, breaks `position: absolute` layouts
- Output varies by browser/OS
- No audit trail — can't guarantee document integrity
- Must upload blob back to server anyway

**Why not storing PDF as BYTEA in Postgres:**
- Database grows fast, backups slow
- Can't share as a presigned URL
- Can't use S3 lifecycle policies

**Why S3 + metadata:**
- PDF is immutable once generated — freezes that version permanently
- Postgres row is cheap: UUID, S3 key, timestamp, file size
- Presigned URLs for download, email, calendar attachment
- S3 lifecycle policies per tenant tier

### S3 key naming convention
```
pdfs/{tenant_id}/{doc_type}/{year}/{month}/{doc_number}.pdf

e.g.
pdfs/t-abc123/invoice/2026/04/INV-2026-04-0012.pdf
```

### Postgres schema addition
```sql
ALTER TABLE documents ADD COLUMN
    pdf_s3_key       TEXT,
    pdf_generated_at TIMESTAMPTZ,
    pdf_size_bytes   INTEGER,
    pdf_version      INTEGER DEFAULT 1;
```

### Generation + storage flow
```go
func (s *Service) FinalizeDocument(ctx context.Context, docID string) error {
    doc, _    := s.repo.GetByID(ctx, docID)
    layout, _ := s.layoutRepo.GetPublished(ctx, doc.TenantID, doc.Type)
    htmlStr, _ := s.renderer.RenderHTML(layout, doc)
    pdfBytes, _ := s.pdfSvc.GeneratePDF(htmlStr)

    t := time.Now()
    s3Key := fmt.Sprintf("pdfs/%s/%s/%d/%02d/%s.pdf",
        doc.TenantID, doc.Type, t.Year(), t.Month(), doc.Number)

    s.storage.Put(ctx, s3Key, pdfBytes, "application/pdf")

    s.repo.SetPDFMeta(ctx, docID, storage.PDFMeta{
        S3Key:       s3Key,
        GeneratedAt: t,
        SizeBytes:   len(pdfBytes),
    })

    return nil
}
```

### Serving downloads (presigned URLs)
```go
func (s *Service) GetDownloadURL(ctx context.Context, docID, requesterTenantID string) (string, error) {
    doc, _ := s.repo.GetByID(ctx, docID)

    if doc.TenantID != requesterTenantID {
        return "", ErrUnauthorized
    }

    // 15 min presigned URL — enough to download, short enough to limit sharing
    return s.storage.PresignedGet(ctx, doc.PDFS3Key, 15*time.Minute)
}
```

### S3 object tags (set at upload)
```go
s.storage.PutWithTags(ctx, s3Key, pdfBytes, "application/pdf", map[string]string{
    "tenant_id": doc.TenantID,
    "doc_type":  doc.Type,
    "env":       "production",
})
```

Enables S3 lifecycle policies per tenant tier without code changes.

---

## Open Questions to Resolve

1. **Regional compliance** — Indonesian e-Faktur (VAT invoices) has strict XML format requirements separate from the PDF. If any early tenants are Indonesian businesses, this must be planned into the invoice data model now.

2. **Public document links** — Do clients (invoice recipients) need a public URL to view/download without logging in? Affects auth middleware design.

3. **Tenant onboarding** — Self-signup via the app, or manually provisioned for now?

---

*Exported from project architecture session — Go + React business document generation SaaS*
