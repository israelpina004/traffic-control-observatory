# Traffic Control Observatory 🔬

The **Traffic Control Observatory** is a real-time, interactive simulation system that visualizes load balancing strategies across multiple servers. It provides a visual dashboard to understand how different load balancing algorithms distribute traffic under simulated concurrent loads.

## Features

- **Real-Time Visualization**: Watch requests flow to 10 different backend servers in real-time via WebSockets.
- **Multiple Load Balancing Strategies**:
  - **Random**: Distributes traffic completely at random.
  - **Round Robin**: Distributes traffic sequentially across all available servers.
  - **Least Connections**: Routes traffic to the server with the fewest active requests.
  - **Peak EWMA (Exponentially Weighted Moving Average)**: Accounts for latency and active connections to route to the healthiest server.
  - **Power of Two Choices (P2C)**: Picks two random servers and routes to the one with fewer active connections.
- **Interactive Dashboard**: Control the simulation (start/stop workers) and switch load balancing algorithms on the fly to see how traffic adapts.
- **Live Metrics**: View live charts of completed, routed, and errored requests, as well as per-server latency metrics.

## Architecture

The project consists of two main components:

1. **Go Backend (`/backend`)**:
   - A custom TCP Load Balancer implementation.
   - 10 mock backend servers that simulate latency and process requests.
   - A high-concurrency worker pool simulating incoming user traffic.
   - An API Server (REST + WebSockets) that broadcasts telemetry events (EventStart, EventBackendSent, EventEnd, EventError) to connected clients.
2. **React Dashboard (`/dashboard`)**:
   - Built with React, Vite, and TypeScript.
   - Connects to the backend via WebSockets to render a real-time traffic canvas.
   - Uses React hooks to manage historical data and graph rendering.

## Getting Started

### Prerequisites
- Go 1.25+
- Node.js & npm

### Running the Backend Locally
1. Navigate to the backend directory:
   ```bash
   cd backend
   ```
2. If you want to use the script, make sure it's executable:
   ```bash
   chmod +x start.sh
   ```
3. Run the script:
   ```bash
   ./start.sh
   ```
   *(Alternatively, run the `mock_backends` and `main.go` separately).*

### Running the Dashboard Locally
1. Open a new terminal and navigate to the dashboard directory:
   ```bash
   cd dashboard
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the Vite development server:
   ```bash
   npm run dev
   ```
4. Open your browser to `http://localhost:5173`.

## Deployment

- **Frontend**: Deployed to Vercel. Live URL: https://lb-simulator.vercel.app
- **Backend**: Deployed to Render. Live URL: https://lb-simulator-backend.onrender.com

## Tech Stack
- **Backend**: Go, WebSockets (`gorilla/websocket`), TCP Sockets
- **Frontend**: React, TypeScript, Vite, Vanilla CSS

## Current Bugs
- Upon refresh the website indicates that the workers aren't running and that a strategy hasn't been chosen even if the user has already started the workers and chosen a strategy.
- How one user uses the website will affect another user's experience on their end. In other words, if another user simultaneously using the app starts the workers, then the other users will also have their workers started, which may not be what they want. This is because the backend is a single process, and the frontend isn't aware of any other users using the app.
