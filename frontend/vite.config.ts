import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

const manualChunkPackages = {
  "react-vendor": ["react", "react-dom", "react-router-dom"],
  "ui-vendor": [
    "@radix-ui/react-avatar",
    "@radix-ui/react-dropdown-menu",
    "@radix-ui/react-label",
    "@radix-ui/react-scroll-area",
    "@radix-ui/react-slot",
    "@radix-ui/react-tabs",
  ],
  "query-vendor": ["@tanstack/react-query"],
  "form-vendor": ["@tanstack/react-form", "zod"],
  "utils-vendor": [
    "date-fns",
    "lucide-react",
    "next-themes",
    "clsx",
    "tailwind-merge",
    "class-variance-authority",
  ],
} as const;

function matchesPackage(id: string, pkg: string) {
  return id.includes(`/node_modules/${pkg}/`) || id.includes(`\\node_modules\\${pkg}\\`);
}

function getManualChunk(id: string) {
  for (const [chunkName, packages] of Object.entries(manualChunkPackages)) {
    if (packages.some((pkg) => matchesPackage(id, pkg))) {
      return chunkName;
    }
  }

  return undefined;
}

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": "/src",
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/health": process.env.VITE_BACKEND_URL || "http://localhost:8375",
      "/ping": process.env.VITE_BACKEND_URL || "http://localhost:8375",
      "/api": {
        target: process.env.VITE_BACKEND_URL || "http://localhost:8375",
        changeOrigin: true,
        ws: true,
        configure: (proxy, _options) => {
          proxy.on("upgrade", (req, socket, head) => {
            console.log("[vite] WebSocket upgrade request:", req.url);
          });
          proxy.on("proxyReqWs", (_proxyReq, _req, socket, _options, _head) => {
            console.log("[vite] Proxying WebSocket connection");
            if (socket && !("destroySoon" in socket)) {
              (socket as any).destroySoon = (socket as any).destroy;
            }
          });
          proxy.on("error", (err, _req, res) => {
            console.error("[vite] Proxy error:", err);
            // res can be either http.ServerResponse or net.Socket
            if (res && "writeHead" in res && !res.headersSent) {
              res.writeHead(500, {
                "Content-Type": "application/json",
              });
              res.end(JSON.stringify({ error: "Proxy error", message: err.message }));
            }
          });
        },
      },
      "/media": {
        target: process.env.VITE_BACKEND_URL || "http://localhost:8375",
        changeOrigin: true,
      },
      "/live": {
        target: "http://media-server:80",
        changeOrigin: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        // Keep chunk grouping stable without relying on Rollup's object-form typing.
        manualChunks: getManualChunk,
      },
    },
    chunkSizeWarningLimit: 600, // Slightly higher limit since we're chunking
  },
});
