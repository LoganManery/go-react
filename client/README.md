```
Go Gin + Vite React App
This project is a full-stack application using Go with Gin on the backend and React with Vite on the frontend.
Project Structure
Copygo-react-app/
├── main.go                 # Main Go server file
├── go.mod                  # Go module definition
├── go.sum                  # Go dependencies checksums
├── client/                 # Vite React app directory
│   ├── public/             # Public assets
│   ├── src/                # React source code
│   │   ├── App.jsx
│   │   ├── App.css
│   │   ├── main.jsx
│   │   └── ...
│   ├── index.html          # HTML template
│   ├── vite.config.js      # Vite configuration
│   ├── package.json        # NPM dependencies
│   └── ...
└── README.md               # Project documentation
Setup and Installation
Prerequisites

Go (1.19 or newer)
Node.js (14 or newer) and npm

Backend Setup

Install Go dependencies:
bashCopygo mod download

Run the Go server:
bashCopygo run main.go


The server will start on port 8080 (default) or the port specified by the PORT environment variable.
Frontend Setup

Navigate to the client directory:
bashCopycd client

Install npm dependencies:
bashCopynpm install

For development:
bashCopynpm run dev

For production build:
bashCopynpm run build


Development Workflow
Option 1: Separate Development Servers (Recommended for Development)
For active development, it's best to run the Vite dev server and Go server separately:

Run the Go server:
bashCopygo run main.go

In another terminal, run the Vite development server:
bashCopycd client
npm run dev

The Vite dev server will automatically proxy API requests to the Go server based on the configuration in vite.config.js
Visit http://localhost:5173 to view the React app with hot module replacement

Option 2: Using the Go Server for Everything (Recommended for Production)

Build the React app:
bashCopycd client
npm run build

Run the Go server which will serve the built React app:
bashCopygo run main.go

Visit http://localhost:8080 to view the application

API Endpoints

GET /api/hello: Returns a JSON message from the server

Environment Variables

PORT: The port on which the server will run (default: 8080)
Vite environment variables can be added with the VITE_ prefix in a .env file in the client directory

Production Deployment

Build the React app:
bashCopycd client
npm run build

Build the Go application:
bashCopygo build -o server

Run the server:
bashCopy./server


Make sure both the server binary and the client/dist directory are included when deploying.
Differences from Create React App
Vite offers several advantages over Create React App:

Faster development server startup - Vite uses ES modules natively
Hot Module Replacement (HMR) - Better hot reloading for faster development
Smaller bundle size - More efficient build output
More flexible configuration - Easier to customize with vite.config.js
Better performance - Both in development and production

Key Vite-specific Files

vite.config.js - Contains configuration for the Vite build tool and development server
index.html - The root HTML template (in the client directory, not in public/)
client/src/main.jsx - The entry point file
```
