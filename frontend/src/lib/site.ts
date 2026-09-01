// Public origin of the site, used for absolute URLs in metadata,
// sitemap.xml and robots.txt. Set via the SITE_URL environment variable;
// falls back to the local dev origin.
export const SITE_URL = process.env.SITE_URL ?? "http://localhost:3000";
