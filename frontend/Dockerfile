# core/Dockerfile


# Use Node image
FROM node:18

# Set working directory
WORKDIR /app

# Copy dependency definitions
COPY package.json package-lock.json ./
RUN npm install

# Copy entire project
COPY . .

# Start the dev server
CMD ["npm", "start"]

# Expose port 3000
EXPOSE 3000
