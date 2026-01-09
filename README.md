# Underground Rap Rankings (Prototype)

Lightweight prototype for community-driven ELO rankings of underground rap artists. No auth, no tracking — just vote, update ratings, and view leaderboards.

## Stack

- Frontend: React + TypeScript + Vite + Tailwind
- Backend: Go 1.25 + Gin (in-memory store simulating Cassandra/Redis/Kafka flows)
- Infra (docker-compose): Cassandra, Redis, Kafka + Zookeeper (not wired yet but scaffolded)

## Quick start (local)

1. Backend: `cd backend && go run ./...`
2. Frontend: `cd frontend && npm install && npm run dev`
3. Open `http://localhost:5173` (proxy hits `http://localhost:8080`).

## Docker

- `make docker-up` to build/run api + frontend + infra
- `make docker-down` to stop and clean volumes

## API

- `GET /api/matchup` → two artists
- `POST /api/vote` with `{ "winner_id": "a1", "loser_id": "a2" }`
- `GET /api/leaderboard?type=overall|velocity`
- `GET /api/trending?period=daily|weekly|momentum`

## Notes

- Data lives in-memory; restart resets stats.
- ELO K-factor is 32; smart pairing favors nearest ELO.
- Redis/Cassandra/Kafka are present in compose for future wiring but unused in code.
