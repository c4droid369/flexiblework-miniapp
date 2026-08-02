# ----- build stage -----
FROM node:20-alpine AS build

WORKDIR /src
COPY package.json package-lock.json* ./
RUN npm install --no-audit --no-fund

COPY . .
RUN npm run build

# ----- runtime stage -----
FROM nginx:1.27-alpine

# Custom nginx config with /api reverse proxy to backend service.
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=build /src/dist /usr/share/nginx/html

EXPOSE 80

# No HEALTHCHECK: nginx:alpine ships without wget/curl, so any probe that
# calls them returns 127 → "unhealthy" forever. Verify liveness by hitting
# the URL instead, or add `apk add --no-cache curl` + `curl -sf` if a real
# healthcheck is needed (adds ~1MB to the image).