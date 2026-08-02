# ----- build stage -----
FROM node:20-alpine AS build

WORKDIR /src
COPY package.json package-lock.json* ./
RUN npm install --no-audit --no-fund

COPY . .
RUN npm run build

# ----- runtime stage -----
FROM nginx:1.27-alpine

# envsubst (from gettext) is used in docker-entrypoint.sh to template
# nginx.conf with the actual BACKEND_PORT at container start, so the
# proxy_pass URLs follow whatever BACKEND_PORT the user passed at run time.
RUN apk add --no-cache gettext

# Treat nginx.conf as a template — moved to /etc/nginx/templates/ and
# rendered into /etc/nginx/conf.d/default.conf on every container start.
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/templates/default.conf.template
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

COPY --from=build /src/dist /usr/share/nginx/html

EXPOSE 80

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]

# No HEALTHCHECK: nginx:alpine ships without wget/curl, so any probe that
# calls them returns 127 → "unhealthy" forever. Verify liveness by hitting
# the URL instead, or add `apk add --no-cache curl` + `curl -sf` if a real
# healthcheck is needed (adds ~1MB to the image).