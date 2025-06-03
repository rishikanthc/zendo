# 1. Build stage: install dependencies & run SvelteKit build
FROM node:18-alpine AS builder
WORKDIR /app

# Copy package manifest and lockfile for caching
COPY package.json package-lock.json ./
RUN npm ci

# Copy source and produce a build
COPY . .
RUN npm run build

# 2. Runtime stage: install only production deps & copy build output
FROM node:18-alpine AS runtime
WORKDIR /app

# Set your DATABASE_URL at build‐time so it’s baked into the image
ENV DATABASE_URL="file:local.db"

# Copy only package files, install prod deps
COPY package.json package-lock.json ./
RUN npm ci --production

# Copy the compiled build artifacts from the builder stage
COPY --from=builder /app/build ./build

# Expose the port your SvelteKit app listens on (default: 3000)
EXPOSE 3000

# Launch the Node server from the build folder
CMD ["node", "build"]
