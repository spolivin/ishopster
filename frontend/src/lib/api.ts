// Server-side client for the FastAPI backend. Every function here runs on
// the Next.js server (Server Components), so API_BASE_URL is read from the
// runtime environment and never shipped to the browser.

export type Category = {
  id: number;
  name: string;
  slug: string;
  description: string | null;
};

export type Product = {
  id: number;
  name: string;
  slug: string;
  description: string | null;
  price: string; // Decimal serialized as a string by the backend
  category: Category;
};

/** One page of products, matching the backend `ProductList` schema. */
export type ProductList = {
  items: Product[];
  total: number; // total matching products, ignoring limit/offset
  limit: number;
  offset: number;
};

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8000";

/** Thrown when the backend answers with a non-2xx status. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

/**
 * Fetch `path` from the backend and parse the JSON body.
 * Throws `ApiError` on any non-2xx response; lets network errors propagate.
 */
async function apiFetch<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`);
  if (!res.ok) {
    throw new ApiError(res.status, `GET ${path} → ${res.status}`);
  }
  return (await res.json()) as T;
}

/** All categories. Any failure propagates (renders the error page). */
export function getCategories(): Promise<Category[]> {
  return apiFetch<Category[]>("/api/categories");
}

/** One category by slug, or `null` if the backend returned 404. */
export async function getCategory(slug: string): Promise<Category | null> {
  try {
    return await apiFetch<Category>(`/api/categories/${slug}`);
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}

/**
 * One page of products, optionally filtered to one category by its slug.
 * Any failure propagates.
 */
export function getProducts(
  opts: { category?: string; limit?: number; offset?: number } = {},
): Promise<ProductList> {
  const params = new URLSearchParams();
  if (opts.category) params.set("category", opts.category);
  if (opts.limit !== undefined) params.set("limit", String(opts.limit));
  if (opts.offset !== undefined) params.set("offset", String(opts.offset));
  const query = params.toString();
  return apiFetch<ProductList>(`/api/products${query ? `?${query}` : ""}`);
}

/**
 * Every product, following pagination to the end. Used by the sitemap,
 * which must list all URLs. Fine while the catalog is small; revisit if it
 * grows into the thousands.
 */
export async function getAllProducts(): Promise<Product[]> {
  const pageSize = 100;
  const all: Product[] = [];
  for (let offset = 0; ; offset += pageSize) {
    const page = await getProducts({ limit: pageSize, offset });
    all.push(...page.items);
    if (offset + pageSize >= page.total) return all;
  }
}

/** One product by slug, or `null` if the backend returned 404. */
export async function getProduct(slug: string): Promise<Product | null> {
  try {
    return await apiFetch<Product>(`/api/products/${slug}`);
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}
