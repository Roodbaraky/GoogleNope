# GoogleNope Frontend

Angular 20 frontend for GoogleNope.

## Setup

Install dependencies:

```powershell
npm install
```

Public runtime configuration lives in `src/environments/environment.ts`.
Do not put OAuth secrets, database credentials, or session secrets in frontend
environment files.

## Development Server

```powershell
npm start
```

Open `http://localhost:4200/`.

## Build

```powershell
npm run build
```

## Tests

```powershell
npm test -- --watch=false --browsers=ChromeHeadless
```

or:

```powershell
npm run test:ci
```
