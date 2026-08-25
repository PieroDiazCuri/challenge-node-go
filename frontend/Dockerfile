FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Etapa final: solo servir la app compilada
FROM node:18-alpine
WORKDIR /app

# Copiar la carpeta dist desde el builder
COPY --from=builder /app/dist ./dist

# Install serve to run the built app
RUN npm install -g serve

# Expose port
EXPOSE 5173

# Comando de arranque
CMD ["serve", "-s", "dist", "-l", "5173"]
