# syntax=docker/dockerfile:1

FROM node:22-alpine AS builder
WORKDIR /app

COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

COPY . .
ARG VITE_API_BASE_URL=/api
ARG VITE_INGEST_BASE_URL=
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_INGEST_BASE_URL=${VITE_INGEST_BASE_URL}
RUN npm run build

FROM nginx:1.27-alpine
# nginx's official entrypoint renders this template with API_UPSTREAM when
# the container starts, so the frontend image does not depend on a Compose
# service name from the backend project.
ENV API_UPSTREAM=http://host.docker.internal:2910
ENV NGINX_ENVSUBST_FILTER=API_UPSTREAM
COPY nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=15s --timeout=5s --retries=5 CMD wget -qO- http://127.0.0.1/healthz || exit 1
