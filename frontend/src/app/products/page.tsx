import type { Metadata } from "next";
import Link from "next/link";

import { getProducts } from "@/lib/api";

// Live catalog data, and the backend is not reachable during `next build`.
export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "All products",
  description: "The full ishopster product catalog.",
  alternates: { canonical: "/products" },
  openGraph: {
    title: "All products",
    description: "The full ishopster product catalog.",
    url: "/products",
  },
};

export default async function ProductsPage() {
  const products = await getProducts();

  return (
    <main>
      <h1>All products</h1>
      <ul>
        {products.map((product) => (
          <li key={product.slug}>
            <Link href={`/products/${product.slug}`}>{product.name}</Link>
            {" — "}
            {product.price}
          </li>
        ))}
      </ul>
    </main>
  );
}
