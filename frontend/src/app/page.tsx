import type { Metadata } from "next";
import Link from "next/link";

import { getCategories } from "@/lib/api";

// Live catalog data, and the backend is not reachable during `next build`.
export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  // The home page title should be exactly "ishopster", not "… — ishopster".
  title: "ishopster",
  description: "Browse laptops, smartphones and accessories.",
  alternates: { canonical: "/" },
};

export default async function Home() {
  const categories = await getCategories();

  return (
    <main>
      <h1>ishopster</h1>

      <nav aria-label="Categories">
        <ul>
          {categories.map((category) => (
            <li key={category.slug}>
              <Link href={`/categories/${category.slug}`}>{category.name}</Link>
            </li>
          ))}
        </ul>
      </nav>

      <p>
        <Link href="/products">See all products</Link>
      </p>
    </main>
  );
}
