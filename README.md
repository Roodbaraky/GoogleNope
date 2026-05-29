# GoogleNope

GoogleNope is a Google Keep-style notes app with a Go/Gin API, MongoDB storage,
OAuth-backed sessions, and an Angular frontend.

## Prerequisites

- Go 1.25.4 or newer
- Node.js and npm compatible with Angular 20
- MongoDB, either local or through Docker
- OAuth/OIDC client credentials for sign-in

## Local MongoDB

The optional compose file starts a local MongoDB instance:

```powershell
docker compose up -d mongo
```

The backend default development connection string is:

```text
mongodb://localhost:27017
```

## Backend Setup

Create a local backend environment file:

```powershell
Copy-Item web-service-gin\.env.example web-service-gin\.env
```

Then edit `web-service-gin\.env` and set at least:

```text
SESSION_SECRET=replace-with-a-long-random-secret
OAUTH_CLIENT_ID=your-client-id
OAUTH_CLIENT_SECRET=your-client-secret
OAUTH_REDIRECT_URL=http://localhost:8080/api/auth/callback
CORS_ALLOWED_ORIGINS=http://localhost:4200,http://127.0.0.1:4200
```

Run the API:

```powershell
cd web-service-gin
go run ./cmd/api
```

The API listens on `http://localhost:8080` by default.

## Frontend Setup

The frontend public API URL is configured in
`frontend/src/environments/environment.ts`. Use
`frontend/src/environments/environment.example.ts` as the public configuration
shape.

Run the Angular app:

```powershell
cd frontend
npm install
npm start
```

The frontend listens on `http://localhost:4200` by default.

## Tests And Quality Checks

Backend:

```powershell
cd web-service-gin
gofmt -w .
go test ./...
go vet ./...
```

Frontend:

```powershell
cd frontend
npm test -- --watch=false --browsers=ChromeHeadless
npm run build
```

Angular lint is not configured in this repository yet. Add an ESLint setup before
treating lint as a required CI check.

## Secrets

Do not commit `.env`, OAuth secrets, database credentials, session secrets, or
other private configuration. Frontend environment files must contain public
runtime configuration only.
