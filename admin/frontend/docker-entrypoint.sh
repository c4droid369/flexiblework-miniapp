#!/bin/sh
# Entrypoint for the admin frontend container. Templates nginx.conf with
# the actual BACKEND_PORT (default 8080) and then hands off to the
# original CMD (nginx). This means the same image works at any port:
#   docker run -e BACKEND_PORT=8081 ...
# without rebuilding.

set -eu

# Robust default handling — busybox ash is picky about `:=` in some
# contexts, so we test-then-assign explicitly. docker-compose passes
# BACKEND_PORT to the frontend container so this fallback only fires
# when running the image standalone.
if [ -z "${BACKEND_PORT:-}" ]; then
    BACKEND_PORT=8080
fi
export BACKEND_PORT

echo "[admin-frontend] proxying /api to backend:${BACKEND_PORT}"

# Substitute ONLY the env vars we declare. Passing an explicit list (vs
# bare `envsubst < template`) keeps nginx's own ${...} variables like
# $uri and $host untouched.
envsubst '${BACKEND_PORT}' \
    < /etc/nginx/templates/default.conf.template \
    > /etc/nginx/conf.d/default.conf

# Hand off to the CMD from the Dockerfile.
exec "$@"