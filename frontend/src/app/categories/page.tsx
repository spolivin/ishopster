import type { Metadata } from "next";
import Link from "next/link";

import { getCategories } from "@/lib/api";

// Live catalog data, and the backend is not reachable during `next build`.
export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Categories",
  description: "Browse all product categories.",
  alternates: { canonical: "/categories" },
  openGraph: {
    title: "Categories",
    description: "Browse all product categories.",
    url: "/categories",
  },
};

export default async function CategoriesPage() {
  const categories = await getCategories();

  return (
    <main>
      <h1>Categories</h1>
      <ul>
        {categories.map((category) => (
          <li key={category.slug}>
            <Link href={`/categories/${category.slug}`}>{category.name}</Link>
          </li>
        ))}
      </ul>
    </main>
  );
}
