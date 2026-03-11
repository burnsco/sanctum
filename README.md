# Sanctum

Sanctum is a social platform centered on hobbies, creativity, and interest-based communities. It combines a Reddit-style feed with realtime chat, multiplayer features, and a Go + React full-stack architecture.

## Highlights

- Community feed and threaded posts
- Direct messages, chatrooms, and realtime social features
- Multiplayer games integrated into the platform
- Go backend with PostgreSQL, Redis, and monitoring support

## Repo layout

- `backend/`: Go API and core services
- `frontend/`: React app
- `docs/`: project and feature documentation

## Quick start

```bash
make setup-local
make dev
```

Default local services:

- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8375`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
