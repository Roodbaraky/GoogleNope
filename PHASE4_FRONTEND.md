# Phase 4 Angular Frontend

This document explains the Phase 4 frontend work completed in `frontend/`.

## What Changed

A new Angular 20 application now lives in `frontend/`.

The app uses standalone components, route-level composition, Angular signals for
page state, and typed API services for the backend contract.

Implemented user-facing flows:

- Session check through `GET /api/auth/me`
- OAuth login redirect through `GET /api/auth/login`
- Logout through `POST /api/auth/logout`
- Authenticated note listing through `GET /api/notes`
- Note creation through `POST /api/notes`
- Edit-on-open with save-on-close through `PATCH /api/notes/:id`
- Delete with browser confirmation through `DELETE /api/notes/:id`
- Loading, empty, auth-required, and error states
- Responsive pinned and unpinned note grids

## Structure

```text
frontend/
  src/
    app/
      core/
        api/
        auth/
      notes/
        data-access/
        feature-notes-page/
        ui/
      shared/
    environments/
```

## Runtime Configuration

The frontend only stores public runtime configuration:

```ts
apiBaseUrl: 'http://localhost:8080'
```

No OAuth client secrets, session secrets, database credentials, or other private
backend values are present in the Angular source.

The backend CORS configuration must allow the frontend origin. The existing
development default is:

```text
CORS_ALLOWED_ORIGINS=http://localhost:4200
```

## Local Commands

Run from `frontend/`:

```powershell
npm install
npm start
npm run build
```

The development URL is:

```text
http://localhost:4200
```

## Verification

The frontend production build passes on Angular 20:

```powershell
npm run build
```

The Angular dev server was also started and responded with HTTP 200 at
`http://localhost:4200`.

## What Is Not Done Yet

Remaining frontend quality work belongs in later phases:

- Unit tests for API services and page state
- Component tests for create/edit/delete behavior
- Production environment configuration strategy
- More polished transient offline detection
- Optional optimistic UI updates
