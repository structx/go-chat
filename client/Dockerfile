
FROM node:20-alpine as builder

WORKDIR /app

COPY . .
RUN npm install

RUN npm ci

RUN npm run build

FROM nginx:alpine

RUN rm -rf /usr/share/nginx/html/*

ENV NODE_ENV production

COPY --from=builder /app/build /usr/share/nginx/html

ENTRYPOINT ["nginx", "-g", "daemon off;"]
