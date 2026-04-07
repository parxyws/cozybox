# Go + SvelteKit Monolith — Chat Export

---

## Q1: Is there a way to make Go backend and React/Svelte frontend in one monolithic project?

Yes, you can combine a Go backend with a React or Svelte frontend in a single monolithic project. Here are the main approaches:

### Option 1: Embed frontend into Go binary (recommended)

Use Go's `embed` package to bundle the compiled frontend assets directly into the Go binary:

```
myapp/
├── main.go
├── server/
│   └── routes.go
├── frontend/
│   ├── src/
│   ├── package.json
│   └── dist/          ← built output, embedded by Go
└── Makefile
```

In Go, you embed the built frontend like this:

```go
import "embed"

//go:embed frontend/dist
var staticFiles embed.FS

// Then serve it:
http.Handle("/", http.FileServer(http.FS(staticFiles)))
```

Your API routes live under `/api/*`, and all other routes serve the frontend's `index.html` (for SPA routing).

### Option 2: Separate directories, same repo

Keep them side by side and run them independently in dev, but deploy together:

```
myapp/
├── backend/
│   ├── main.go
│   └── go.mod
├── frontend/
│   ├── src/
│   └── package.json
└── Makefile          ← orchestrates build + dev
```

### Tooling to Wire It Together

**Makefile** is the simplest glue:

```makefile
dev:
    cd frontend && npm run dev &
    go run ./backend

build:
    cd frontend && npm run build
    cp -r frontend/dist backend/static/
    cd backend && go build -o ../bin/app .
```

**Or use [Air](https://github.com/air-verse/air)** for Go hot reload alongside Vite's dev server.

### Dev Proxy Setup (avoid CORS headaches)

In dev, run Go on `:8080` and Vite on `:5173`, then proxy API calls in `vite.config.js`:

```js
export default {
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    }
  }
}
```

### React vs Svelte for this setup

| | React | Svelte |
|---|---|---|
| Ecosystem | Larger | Smaller but growing |
| Bundle size | Larger | Smaller |
| Learning curve | Moderate | Easier |
| Vite support | ✅ | ✅ |
| Go embed friendly | ✅ | ✅ |

**Svelte + Go** is a particularly nice pairing because Svelte compiles to vanilla JS (no runtime), keeping bundle sizes tiny.

### Recommended Stack

- **Go** + `net/http` or Chi / Gin for routing
- **Svelte or React** with **Vite** as the bundler
- **`//go:embed`** to bundle the frontend into the binary
- **Makefile** or **Task** to orchestrate builds

---

## Q2: Is it possible to use SvelteKit and embed it into Golang?

Yes, but SvelteKit needs a small configuration tweak — switch to the **static adapter**, which outputs plain HTML/CSS/JS files that Go can serve.

### 1. Install the static adapter

```bash
npm i -D @sveltejs/adapter-static
```

### 2. Configure `svelte.config.js`

```js
import adapter from '@sveltejs/adapter-static';

export default {
  kit: {
    adapter: adapter({
      pages: 'build',
      assets: 'build',
      fallback: 'index.html'
    }),
    paths: {
      base: ''
    }
  }
};
```

### 3. Disable SSR globally

Create `src/routes/+layout.js`:

```js
export const prerender = true;
export const ssr = false;
```

### Project Structure

```
myapp/
├── main.go
├── go.mod
├── server/
│   └── routes.go
├── static/          ← embedded frontend assets
│   └── (copied from frontend/build)
└── frontend/        ← SvelteKit project
    ├── src/
    ├── svelte.config.js
    └── package.json
```

### Embedding in Go

```go
package main

import (
    "embed"
    "io/fs"
    "net/http"
)

//go:embed static
var staticFiles embed.FS

func main() {
    mux := http.NewServeMux()

    // API routes
    mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"message": "hello from Go"}`))
    })

    // Serve SvelteKit static build
    staticFS, _ := fs.Sub(staticFiles, "static")
    fileServer := http.FileServer(http.FS(staticFS))

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if _, err := staticFS.Open(r.URL.Path); err != nil {
            r.URL.Path = "/"
        }
        fileServer.ServeHTTP(w, r)
    })

    http.ListenAndServe(":8080", mux)
}
```

### Makefile

```makefile
.PHONY: dev build

dev:
	cd frontend && npm run dev &
	air

build:
	cd frontend && npm run build
	rm -rf static && cp -r frontend/build static
	go build -o bin/app .

clean:
	rm -rf bin static frontend/build
```

### Dev Proxy

```js
export default {
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
};
```

### Caveats

| Issue | Solution |
|---|---|
| SvelteKit SSR features won't work | Use `ssr = false` + static adapter |
| Dynamic routes (e.g. `/post/[id]`) | Set `fallback: 'index.html'` in adapter config |
| `$env/dynamic/private` won't work | Use Go API to expose env values instead |
| Page-level prerendering | Works fine with `prerender = true` |

---

## Q3: Why is SSR globally disabled?

SSR is disabled because SvelteKit's SSR expects a **Node.js server runtime**, but Go is your server.

### What SSR Does

When SSR is enabled, SvelteKit renders pages to HTML **on the server** before sending them to the browser. This requires a running Node.js process.

```
Normal SvelteKit SSR:
  Request → Node.js (runs SvelteKit) → renders HTML → sends to browser

Your Go setup:
  Request → Go (just serves files) → browser gets static HTML/JS → browser renders
```

Go has **no idea** how to execute Svelte components. It just serves files.

### What Happens If You Don't Disable SSR

SvelteKit will try to prerender every page at build time. This breaks when a page:
- Fetches data dynamically (e.g. from your Go API)
- Uses browser-only APIs (`window`, `localStorage`)
- Has routes that depend on runtime parameters

You'd get build errors like:
```
ReferenceError: window is not defined
Error: Cannot prerender pages with dynamic data
```

### You Don't Always Need It Global

```js
// Per-page disable
export const ssr = false;

// Per-page prerender (works fine)
export const prerender = true;
```

### Mixed approach

```
src/routes/
├── +layout.js           ← ssr = false globally
├── (static)/
│   ├── about/           ← override with prerender = true
│   └── pricing/         ← override with prerender = true
└── (dynamic)/
    ├── dashboard/       ← stays ssr = false, loads from Go API
    └── profile/         ← stays ssr = false
```

### Summary

| | SSR enabled | SSR disabled |
|---|---|---|
| Renders on | Node.js server | Browser |
| Needs Node at runtime | ✅ Yes | ❌ No |
| Works with Go embed | ❌ No | ✅ Yes |
| Good for SEO | ✅ Better | ⚠️ Needs extra work |
| Dynamic data | ✅ Server `load()` | Fetch from Go API on client |

---

## Q4: How to handle dynamic data since `+page.server.js` and `+layout.server.js` won't work?

### What You Lose

```
src/routes/
├── +page.server.js   ❌ Dead — Go can't run this
├── +layout.server.js ❌ Dead — Go can't run this
└── +page.js          ✅ Works — runs in browser
```

### The Mental Shift

Think of Go as the replacement for everything `.server.js` used to do:

| SvelteKit fullstack | Go + Static SvelteKit |
|---|---|
| `+page.server.js` load | `+page.js` load + Go API |
| `+page.server.js` actions | `fetch` POST to Go API |
| `+layout.server.js` auth | Go middleware + `/api/me` |
| `$env/dynamic/private` | Go env vars, exposed via API |
| Server hooks | Go middleware |

### Handling Dynamic Data

**Go side:**
```go
mux.HandleFunc("/api/posts/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    post := db.QueryPost(id)
    json.NewEncoder(w).Encode(post)
})
```

**Svelte side (`+page.js`):**
```js
export async function load({ fetch, params }) {
  const res = await fetch(`/api/posts/${params.id}`);
  const post = await res.json();
  return { post };
}
```

```svelte
<!-- +page.svelte -->
<script>
  export let data;
</script>

<h1>{data.post.title}</h1>
<p>{data.post.body}</p>
```

### Form Actions Replacement

```svelte
<script>
  async function handleSubmit(e) {
    const form = new FormData(e.target);
    await fetch('/api/posts', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: form.get('title') })
    });
  }
</script>

<form on:submit|preventDefault={handleSubmit}>
  <input name="title" />
  <button type="submit">Create</button>
</form>
```

### Authentication

**Go — set session cookie:**
```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    http.SetCookie(w, &http.Cookie{
        Name:     "session",
        Value:    sessionToken,
        HttpOnly: true,
        Path:     "/",
    })
}

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session")
        if err != nil || !isValidSession(cookie.Value) {
            http.Error(w, "Unauthorized", 401)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**Svelte — check auth state via API:**
```js
// src/routes/+layout.js
export async function load({ fetch }) {
  const res = await fetch('/api/me');
  if (!res.ok) return { user: null };
  const user = await res.json();
  return { user };
}
```

```svelte
<!-- +layout.svelte -->
<script>
  import { goto } from '$app/navigation';
  export let data;

  $: if (!data.user) goto('/login');
</script>
```

---

*Exported on 2026-04-06*
