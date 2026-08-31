# Single-stage image, kept deliberately simple for the learning phase.
# A production setup would use a multi-stage build with Next.js'
# `output: "standalone"` to ship a much smaller runtime image.

FROM node:22-bookworm-slim

# pnpm, pinned to the same version as package.json "packageManager".
RUN npm install -g pnpm@11.10.0

WORKDIR /app

# 1. Install dependencies first. This layer is only rebuilt when the
#    manifest / lockfile / workspace config change, not on every source
#    edit. `--frozen-lockfile` fails if pnpm-lock.yaml is out of sync
#    with package.json (the equivalent of `npm ci`).
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

# 2. Copy the rest of the source and build the production bundle.
COPY . .
RUN pnpm run build

# 3. Run the production server. `next start` always runs in production mode.
ENV NODE_ENV=production
EXPOSE 3000
CMD ["pnpm", "run", "start"]
