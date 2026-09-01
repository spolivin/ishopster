import type { Metadata } from "next";

import { SITE_URL } from "@/lib/site";
import "./globals.css";

export const metadata: Metadata = {
  // Base for resolving relative URLs in metadata (canonical, OG images).
  metadataBase: new URL(SITE_URL),
  title: {
    default: "ishopster",
    // Page titles become "<page> — ishopster".
    template: "%s — ishopster",
  },
  description: "A learning e-commerce project.",
  openGraph: {
    siteName: "ishopster",
    type: "website",
  },
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
