import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { JsonLd } from "@/components/JsonLd";
import { getProduct } from "@/lib/api";
import { SITE_URL } from "@/lib/site";

type Params = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Params): Promise<Metadata> {
  const { slug } = await params;
  const product = await getProduct(slug);
  if (!product) return {};

  const description = product.description ?? product.name;
  const path = `/products/${product.slug}`;

  return {
    title: product.name,
    description,
    alternates: { canonical: path },
    openGraph: { title: product.name, description, url: path },
  };
}

export default async function ProductPage({ params }: Params) {
  const { slug } = await params;

  const product = await getProduct(slug);
  if (!product) notFound();

  // TODO: priceCurrency should come from the backend once the catalog
  // stores a currency. Hardcoded to USD for now.
  const productLd = {
    "@context": "https://schema.org",
    "@type": "Product",
    name: product.name,
    description: product.description ?? undefined,
    offers: {
      "@type": "Offer",
      price: product.price,
      priceCurrency: "USD",
      availability: "https://schema.org/InStock",
      url: `${SITE_URL}/products/${product.slug}`,
    },
  };

  const breadcrumbLd = {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: [
      { "@type": "ListItem", position: 1, name: "Home", item: SITE_URL },
      {
        "@type": "ListItem",
        position: 2,
        name: product.category.name,
        item: `${SITE_URL}/categories/${product.category.slug}`,
      },
      {
        "@type": "ListItem",
        position: 3,
        name: product.name,
        item: `${SITE_URL}/products/${product.slug}`,
      },
    ],
  };

  return (
    <main>
      <JsonLd data={productLd} />
      <JsonLd data={breadcrumbLd} />

      <nav aria-label="Breadcrumb">
        <Link href="/">Home</Link>
        {" / "}
        <Link href={`/categories/${product.category.slug}`}>
          {product.category.name}
        </Link>
        {" / "}
        {product.name}
      </nav>

      <article>
        <h1>{product.name}</h1>
        <p>{product.price}</p>
        {product.description && <p>{product.description}</p>}
      </article>
    </main>
  );
}
