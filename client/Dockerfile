
FROM node:latest as builder

WORKDIR /app

COPY . .
RUN npm install --legacy-peer-deps

RUN npm run build

FROM nginx:alpine

RUN rm -rf /usr/share/nginx/html/*

ENV NODE_ENV production

COPY --from=builder /app/build /usr/share/nginx/html

EXPOSE 80

ENTRYPOINT ["nginx", "-g", "daemon off;"]
