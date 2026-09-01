/**
 * Renders a JSON-LD structured-data block. Search engines read the
 * script contents; it is invisible to users.
 *
 * `dangerouslySetInnerHTML` is required because React escapes text nodes,
 * which would corrupt the JSON. The value is our own serialized object,
 * not user input.
 */
export function JsonLd({ data }: { data: Record<string, unknown> }) {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }}
    />
  );
}
