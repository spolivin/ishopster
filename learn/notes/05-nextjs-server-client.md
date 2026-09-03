# Lesson 5 — Next.js: server vs client components

Everything in lessons 1–4 was plain React running **in the browser**.
Next.js App Router adds a second place your components can run: **on the
server, during the request, before any HTML is sent**. This lesson is only
about that split.

---

## 1. Two kinds of component

|                                                  | Server Component                 | Client Component                                               |
| ------------------------------------------------ | -------------------------------- | -------------------------------------------------------------- |
| **Marker**                                       | none — this is the **default**   | `"use client"` on the first line of the file                   |
| **Runs where**                                   | on the server, once, per request | on the server once (for the initial HTML), then in the browser |
| **Ships JS to the browser?**                     | no                               | yes                                                            |
| **Can be `async` / use `await`**                 | ✅                               | ❌                                                             |
| **Can read DB / call the API directly**          | ✅                               | ❌ (must go through an API route)                              |
| **`useState`, `useEffect`, `onClick`, `window`** | ❌                               | ✅                                                             |

You do **not** import anything to make a Server Component. A file in `app/`
is a Server Component until you write `"use client"` at the top.

---

## 2. Why a Server Component can be `async`

In the browser you cannot pause rendering to wait for data — that would
freeze the page. That is why lesson 4 needed `useEffect` + `loading` state:
render first, fetch after, re-render.

On the server there is no page to freeze. Next can just wait:

```tsx
// Server Component — no "use client"
export default async function ProductsPage() {
  const products = await getProducts(); // server waits here
  return (
    <ul>
      {products.map((p) => (
        <li key={p.slug}>{p.name}</li>
      ))}
    </ul>
  );
}
```

- No `useState`, no `useEffect`, no `loading` branch.
- The server finishes the `await`, renders the finished `<ul>` to HTML,
  and sends HTML that **already contains the products**.
- A search engine and the user get the full content immediately.

The lesson-4 pattern is the browser workaround. The lesson-5 pattern is what
you use when the component runs on the server.

---

## 3. `"use client"` — the boundary

`"use client"` at the top of a file means: this file and the components in
it are Client Components. Add it only when the component needs something
from the client column of the table above — state, an effect, an event
handler, a browser API, or a client-only hook.

Rules:

- A **Server** Component may render a **Client** Component. ✅
- A **Client** Component may **not** import and render a Server Component,
  but it can receive one as `children` / props. ✅
- Everything a Client Component imports becomes client too (the boundary
  flows downward).

### In ishopster

| File                             | Kind   | Why                                                              |
| -------------------------------- | ------ | ---------------------------------------------------------------- |
| `app/products/page.tsx`          | Server | just `await getProducts()` and renders a list — no interactivity |
| `app/products/[slug]/page.tsx`   | Server | same                                                             |
| `app/categories/[slug]/page.tsx` | Server | same                                                             |
| `components/JsonLd.tsx`          | Server | renders a `<script>` tag, no state                               |
| a future "add to cart" button    | Client | needs `onClick` + state                                          |
| a future search box              | Client | needs `useState` + the typed value                               |

Catalog pages stay Server Components on purpose: less JavaScript shipped,
and the HTML is crawlable. That is the SEO argument in `CLAUDE.md`.

---

## 4. `generateMetadata`

A page file can export a second special function. Next calls it on the
server to build the `<head>`: `<title>`, `<meta name="description">`,
`<link rel="canonical">`, Open Graph tags.

```tsx
export async function generateMetadata({
  params,
  searchParams,
}): Promise<Metadata> {
  const { slug } = await params;
  const product = await getProduct(slug);
  return {
    title: product.name,
    alternates: { canonical: `/products/${product.slug}` },
  };
}
```

- It is `async` and runs on the server, like a Server Component.
- It gets `params` (the dynamic route parts, e.g. `[slug]`) and
  `searchParams` (the `?page=2` part) — both are Promises in this version,
  so you `await` them.
- It is separate from the component so Next can produce the `<head>`
  independently of the page body.
- If it fetches the same URL the page fetches, the request is **deduplicated**
  within that one request — one real call to the backend.

---

## 5. The full request path for `/products`

```
browser  ──GET /products──▶  Next.js server
                               │
                               ├─ runs generateMetadata()  ─┐
                               ├─ runs <ProductsPage/>       │  both await
                               │                             │
                               │        getProducts()  ──HTTP GET /api/products──▶ FastAPI
                               │                                                     │
                               │                                              SQLAlchemy → PostgreSQL
                               │                                                     │
                               │        products  ◀────────────JSON─────────────────┘
                               │
                               ├─ React renders the page to HTML  ◀┘
                               │
      finished HTML  ◀─────────┘   (already contains the <li> items and <head>)
        │
        ▼
   browser shows the page immediately, then "hydrates":
   downloads the JS only for Client Components and wires up their
   interactivity. Server Components ship no JS.
```

Key point: `getProducts` runs **on the Next server**, not in the browser.
`API_BASE_URL` is a server env var; the browser never sees it and never
talks to FastAPI directly.

---

## Reading exercise

Open the real file `frontend/src/app/products/page.tsx` and answer:

1. Is it a Server or a Client Component? How can you tell in one glance?
2. Where does `await getProducts(...)` actually execute — user's browser,
   Next server, or FastAPI?
3. `parsePage` and `hrefForPage` are plain functions, not components. Why
   is that fine here, and would anything change if this were a Client
   Component?
4. It uses `notFound()` and reads `searchParams`. Neither is a React hook —
   why can a Server Component use them when it cannot use `useState`?
5. The `<link rel="prev">` / `<link rel="next">` tags are returned from the
   component body, not from `generateMetadata`. What has to be true about
   how React/Next handles those tags for this to work?

---

## Answers

**1. Server Component.** No `"use client"` on the first line → it is a
Server Component by default. That is the one-glance check.

**2. On the Next.js server.** The page component runs on the server during
the request; `await getProducts()` calls the API client, which does an HTTP
`GET` to FastAPI using `API_BASE_URL` (a server-only env var the browser
never sees). FastAPI talks to PostgreSQL and returns JSON. The browser only
ever receives finished HTML.

**3. Fine because they are pure functions, not components or hooks** — they
take input, return output, touch nothing external, so they run identically
on server or client. Switching this file to a Client Component would _not_
break `parsePage` / `hrefForPage`. What _would_ break is the component
itself: a Client Component cannot be `async` and cannot `await getProducts()`
directly — you would need `useState` + `useEffect` + a `loading` branch, the
lesson-4 pattern. Portable helpers, non-portable data-fetching.

**4. Because `useState`/`useEffect` are hooks, and hooks need two things a
Server Component does not have:** the client React runtime (the render
machinery), and a component that _persists across re-renders_ so stored
state means something. A Server Component renders once and is gone.
`notFound()` and `searchParams` are not hooks:

- `notFound()` just throws a special error that Next catches to render the
  404 page — a one-way signal, nothing to persist.
- `searchParams` is a prop Next passes into the page component (a Promise
  you `await`), like `params` — just input data.

The rule is not "server can't do dynamic things"; it is "hooks need the
client runtime + a component living across renders."

**5. React must hoist document-metadata tags (`<link>`, `<meta>`, `<title>`)
out of wherever they are rendered in the tree and into `<head>`.** React 19
(used by Next 16) has built-in special handling for these elements. Without
it, a `<link>` rendered inside `<main>` would just sit in the body and a
crawler would ignore it for canonical / pagination. Older React could not do
this — you needed `next/head` or the Metadata API for every tag.

---

Next: go through `frontend/src/app/products/page.tsx` line by line.
