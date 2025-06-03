# ┌─────────────────────────────────────────────────────────────────────────────┐
# │ 1. Build stage: install deps + run SvelteKit build                         │
# └─────────────────────────────────────────────────────────────────────────────┘
FROM node:18-alpine AS builder
WORKDIR /app

# 1a. Copy package.json + lockfile so 'npm ci' can cache layers
COPY package.json package-lock.json ./
RUN npm ci

# 1b. Copy the rest of your source (including drizzle.config.json, migrations/)
COPY . .
RUN npm run build

# ┌─────────────────────────────────────────────────────────────────────────────┐
# │ 2. Runtime stage: install deps (including Drizzle CLI), copy build + config │
# └─────────────────────────────────────────────────────────────────────────────┘
FROM node:18-alpine AS runtime
WORKDIR /app

# 2a. Copy package.json & lockfile, then install ALL dependencies (so Drizzle CLI exists)
COPY package.json package-lock.json ./
RUN npm ci

# 2b. Copy drizzle config + migrations so drizzle-kit can see them at runtime
COPY drizzle.config.ts ./
COPY src ./src

# 2c. Bake DATABASE_URL into the image (optional if you set via docker-compose)
ENV DATABASE_URL="file:local.db"

# 2d. Copy the built SvelteKit output from the builder stage
COPY --from=builder /app/build ./build

# 2e. Copy our entrypoint script (runs migrations first)
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# 2f. Expose the port and set entrypoint + default command
EXPOSE 3000
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["node", "build/index.js"]
