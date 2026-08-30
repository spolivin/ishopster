# Single-stage image, kept deliberately simple for the learning phase.
# A production setup would use a multi-stage build with Next.js'
# `output: "standalone"` to ship a much smaller runtime image.

FROM node:22-bookworm-slim

WORKDIR /app

# 1. Install dependencies first. This layer is only rebuilt when
#    package.json / package-lock.json change, not on every source edit.
#    `npm ci` installs exactly what the lockfile pins (incl. devDependencies,
#    which `next build` needs: typescript, eslint, @types/*).
COPY package.json package-lock.json ./
RUN npm ci

# 2. Copy the rest of the source and build the production bundle.
COPY . .
RUN npm run build

# 3. Run the production server. `next start` always runs in production mode.
ENV NODE_ENV=production
EXPOSE 3000
CMD ["npm", "run", "start"]
