
# frontend/Dockerfile


# Use Node image
FROM node:18

# Set working directory
WORKDIR /app

# Copy dependency definitions
COPY package.json package-lock.json ./
RUN npm install

# Copy entire project
COPY . .

# Start Vite dev server
CMD ["npm", "run", "dev"]

# Expose Vite's default port
EXPOSE 5173

