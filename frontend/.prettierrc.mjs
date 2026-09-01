/** @type {import("prettier").Config} */
const config = {
  plugins: ["@ianvs/prettier-plugin-sort-imports"],
  importOrder: [
    "^(react|react-dom)$", // React first
    "^next(/.*)?$", // Next.js
    "<THIRD_PARTY_MODULES>", // other node_modules
    "", // blank line
    "^@/(.*)$", // internal alias imports
    "",
    "^[./]", // relative imports
  ],
  importOrderTypeScriptVersion: "5.0.0",
};

export default config;
