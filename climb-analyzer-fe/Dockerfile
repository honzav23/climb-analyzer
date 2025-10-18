
ARG NODE_VERSION=24.7.0-alpine

FROM node:${NODE_VERSION} AS builder

WORKDIR /app/FE

COPY ./climb-analyzer-fe/package.json ./climb-analyzer-fe/package-lock.json ./

RUN --mount=type=cache,target=/root/.npm npm ci

COPY ./climb-analyzer-fe .

RUN npm run build


FROM nginx:alpine AS runner

# Use a built-in non-root user for security best practices
USER root

# Copy custom Nginx config
COPY ./climb-analyzer-fe/nginx.conf /etc/nginx/nginx.conf

# Copy the static build output from the build stage to Nginx's default HTML serving directory
COPY --chown=nginx:nginx --from=builder /app/FE/dist/*/browser /usr/share/nginx/html

EXPOSE 80

# Start Nginx directly with custom config
ENTRYPOINT ["nginx", "-c", "/etc/nginx/nginx.conf"]
CMD ["-g", "daemon off;"]