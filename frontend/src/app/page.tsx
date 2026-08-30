// Server Component (no "use client"): this runs on the Next.js server,
// so the fetch below is a server-to-server request. No CORS, and the
// result is baked into the HTML that reaches the browser / crawlers.

type Health = { status: string };

// Render this page on every request (never prerender at build time), so
// the backend is checked live and API_BASE_URL is read from the runtime
// environment, not baked into the build.
export const dynamic = "force-dynamic";

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8000";

async function getBackendHealth(endpoint: string): Promise<Health | null> {
  try {
    // fetch is not cached by default in this Next.js version, so each
    // request re-checks the backend.
    const res = await fetch(`${API_BASE_URL}${endpoint}`);
    if (!res.ok) return null;
    return (await res.json()) as Health;
  } catch {
    return null;
  }
}

export default async function Home() {
  const [health, healthDb] = await Promise.all([
    getBackendHealth("/health"),
    getBackendHealth("/health/db"),
  ]);

  return (
    <main>
      <h1>ishopster</h1>
      <p>Backend health: {health ? health.status : "unavailable"}</p>
      <p>Backend DB health: {healthDb ? healthDb.status : "unavailable"}</p>
    </main>
  );
}
