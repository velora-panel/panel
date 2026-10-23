FROM node:24.19.0-alpine AS build
RUN npm install --global pnpm@11.21.0
WORKDIR /src
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY packages/ ./packages/
COPY apps/console/ ./apps/console/
COPY apps/core/package.json ./apps/core/package.json
COPY apps/cli/package.json ./apps/cli/package.json
RUN pnpm install --frozen-lockfile
RUN pnpm --filter @velora/console build
FROM nginxinc/nginx-unprivileged:1.29-alpine
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/apps/console/dist /usr/share/nginx/html
EXPOSE 8080
