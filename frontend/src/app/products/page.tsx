import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { getProducts } from "@/lib/api";

// Live catalog data, and the backend is not reachable during `next build`.
export const dynamic = "force-dynamic";

const PAGE_SIZE = 20;

// Next passes the same props object to both generateMetadata and the page
// component. `searchParams` is a Promise in this version of Next, so it is
// awaited before use. We only read `page`, hence the narrow shape.
type Props = {
  searchParams: Promise<{ page?: string }>;
};

// Parse ?page= into a 1-based integer. Anything invalid falls back to 1.
function parsePage(raw: string | undefined): number {
  const n = Number(raw);
  return Number.isInteger(n) && n >= 1 ? n : 1;
}

const hrefForPage = (page: number) =>
  page === 1 ? "/products" : `/products?page=${page}`;

export async function generateMetadata({
  searchParams,
}: Props): Promise<Metadata> {
  const page = parsePage((await searchParams).page);
  const path = hrefForPage(page);
  const title = page === 1 ? "All products" : `All products – page ${page}`;
  const description = "The full ishopster product catalog.";

  return {
    title,
    description,
    // Each page is its own canonical: pointing page 2+ at page 1 would hide
    // the products only listed further down from search engines.
    alternates: { canonical: path },
    openGraph: { title, description, url: path },
  };
}

export default async function ProductsPage({ searchParams }: Props) {
  const page = parsePage((await searchParams).page);
  const offset = (page - 1) * PAGE_SIZE;

  const { items, total } = await getProducts({ limit: PAGE_SIZE, offset });

  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  // A crawler (or a user) asking for a page past the end gets a real 404
  // rather than a soft "Page 5 of 1" with an empty list.
  if (page > pageCount) notFound();

  const hasPrev = page > 1;
  const hasNext = page < pageCount;

  return (
    <main>
      {/* Hoisted into <head> by React. rel=prev/next describe the sequence
          to crawlers. */}
      {hasPrev && <link rel="prev" href={hrefForPage(page - 1)} />}
      {hasNext && <link rel="next" href={hrefForPage(page + 1)} />}

      <h1>All products</h1>

      {items.length === 0 ? (
        <p>No products on this page.</p>
      ) : (
        <ul>
          {items.map((product) => (
            <li key={product.slug}>
              <Link href={`/products/${product.slug}`}>{product.name}</Link>
              {" — "}
              {product.price}
            </li>
          ))}
        </ul>
      )}

      <nav aria-label="Pagination">
        {hasPrev && (
          <Link href={hrefForPage(page - 1)} rel="prev">
            ← Previous
          </Link>
        )}{" "}
        <span>
          Page {page} of {pageCount}
        </span>{" "}
        {hasNext && (
          <Link href={hrefForPage(page + 1)} rel="next">
            Next →
          </Link>
        )}
      </nav>
    </main>
  );
}
