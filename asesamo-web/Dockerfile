# ==============================================================================
# ASESAMO WEB — Production Multi-Stage Dockerfile
# Service: asesamo-web (Next.js 14 App Router)
# ==============================================================================

# ------------------------------------------------------------------------------
# Stage 1: Dependency Installation
# ------------------------------------------------------------------------------
FROM node:20-alpine AS deps

# Install libc6-compat for compatibility with native library binaries
RUN apk add --no-cache libc6-compat

WORKDIR /app

# Copy dependency manifests
COPY package.json package-lock.json ./
RUN npm ci

# ------------------------------------------------------------------------------
# Stage 2: Application Build
# ------------------------------------------------------------------------------
FROM node:20-alpine AS builder

WORKDIR /app

# Copy node_modules from deps stage
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Build argument for Next.js public client-side URL
ARG NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}

# Optimize build performance and disable telemetry
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production

# Ensure public folder exists so subsequent COPY in runner never fails if repo has no public directory
RUN mkdir -p /app/public

# Build Next.js application with standalone output enabled
RUN npm run build

# ------------------------------------------------------------------------------
# Stage 3: Minimal Hardened Runtime
# ------------------------------------------------------------------------------
FROM node:20-alpine AS runner

WORKDIR /app

# Production environment variables
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"
ENV NEXT_TELEMETRY_DISABLED=1

# Install curl for healthcheck capabilities
RUN apk add --no-cache curl

# Create unprivileged system group and user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Set correct ownership for runtime execution
# Copy standalone server build output
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
# Copy static assets (required by Next.js standalone server)
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
# Copy public assets
COPY --from=builder --chown=nextjs:nodejs /app/public ./public

# Switch to non-root user
USER nextjs

# Expose web port
EXPOSE 3000

# Healthcheck probe matching docker-compose expectations
HEALTHCHECK --interval=5s --timeout=5s --start-period=15s --retries=5 \
  CMD curl -f http://localhost:3000 || exit 1

# Launch Next.js standalone Node.js server
CMD ["node", "server.js"]
