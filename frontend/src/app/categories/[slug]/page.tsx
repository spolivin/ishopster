import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { JsonLd } from "@/components/JsonLd";
import { getCategory, getProducts } from "@/lib/api";
import { SITE_URL } from "@/lib/site";

type Params = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Params): Promise<Metadata> {
  const { slug } = await params;
  const category = await getCategory(slug);
  if (!category) return {};

  const description =
    category.description ?? `Products in the ${category.name} category.`;
  const path = `/categories/${category.slug}`;

  return {
    title: category.name,
    description,
    alternates: { canonical: path },
    openGraph: { title: category.name, description, url: path },
  };
}

export default async function CategoryPage({ params }: Params) {
  const { slug } = await params;

  const category = await getCategory(slug);
  if (!category) notFound();

  const products = await getProducts(slug);

  const breadcrumbs = {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: [
      { "@type": "ListItem", position: 1, name: "Home", item: SITE_URL },
      {
        "@type": "ListItem",
        position: 2,
        name: "Categories",
        item: `${SITE_URL}/categories`,
      },
      {
        "@type": "ListItem",
        position: 3,
        name: category.name,
        item: `${SITE_URL}/categories/${category.slug}`,
      },
    ],
  };

  return (
    <main>
      <JsonLd data={breadcrumbs} />

      <nav aria-label="Breadcrumb">
        <Link href="/">Home</Link>
        {" / "}
        <Link href="/categories">Categories</Link>
        {" / "}
        {category.name}
      </nav>

      <h1>{category.name}</h1>
      {category.description && <p>{category.description}</p>}

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
