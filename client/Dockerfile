# Stage 1: Build the application
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package.json and package-lock.json to the workspace
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy the entire project to the workspace
COPY . .

# Build the project
RUN npm run build

# Stage 2: Create the production image
FROM nginx:latest

# Copy the built files from the previous stage to the NGINX web root directory
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose port 80 (default for NGINX)
EXPOSE 80

# Start NGINX
CMD ["nginx", "-g", "daemon off;"]