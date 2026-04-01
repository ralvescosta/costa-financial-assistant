# Data Model: Frontend Authentication & Navigation System

**Feature**: 002-frontend-auth-navigation  
**Created**: 2026-04-01

## Frontend State Model

### AuthenticationContext (React Context / State)

Centralized authentication state managed by the frontend application:

```typescript
interface AuthenticationContext {
  // Authentication Status
  isAuthenticated: boolean
  isLoading: boolean
  error?: string

  // Session Metadata (tokens are HTTP-only cookies managed by BFF)
  expiresIn?: number
  expiryTimestamp?: number
  refreshAtTimestamp?: number
  csrfToken?: string
  lockoutUntil?: number

  // User Information
  user?: {
    id: string
    username: string
    email?: string
    fullName?: string
    avatar?: string
  }

  // Active Project Context (Multi-Tenancy)
  activeProject?: {
    id: string
    name: string
    role: 'read_only' | 'update' | 'write'
  }

  // Actions
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  refreshAccessToken: () => Promise<void>
  setActiveProject: (projectId: string) => void
}
```

### Navigation State

Route and sidebar state:

```typescript
interface NavigationState {
  currentRoute: string
  activeMenuItem?: string
  sidebarOpen: boolean
  breadcrumbs: BreadcrumbItem[]
}

interface BreadcrumbItem {
  label: string
  route: string
  icon?: string
}
```

### Session Storage Schema

Data persisted in browser storage:

```json
{
  "auth": {
    "expiresIn": 3600,
    "expiryTimestamp": 1682592000,
    "refreshAtTimestamp": 1682591100,
    "lockoutUntil": null,
    "user": {
      "id": "user-123",
      "username": "demo"
    },
    "activeProject": {
      "id": "project-456",
      "name": "Personal Finance"
    }
  },
  "ui": {
    "theme": "light",
    "sidebarOpen": true
  }
}
```

## API Contracts (BFF Endpoints)

### Login Endpoint Request

```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "demo",
  "password": "demo123"
}
```

### Login Endpoint Response (Success)

```json
{
  "statusCode": 200,
  "data": {
    "expiresIn": 3600,
    "refreshAt": 2700,
    "csrfToken": "csrf_opaque_token",
    "user": {
      "id": "user-123",
      "username": "demo",
      "email": "demo@example.com"
    },
    "activeProject": {
      "id": "project-456",
      "name": "Personal Finance",
      "role": "write"
    }
  }
}
```

### Login Endpoint Response (Failure)

```json
{
  "statusCode": 401,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid username or password"
  }
}
```

### Login Endpoint Response (Lockout)

```json
{
  "statusCode": 429,
  "error": {
    "code": "AUTH_LOCKED",
    "message": "Too many failed attempts. Try again later.",
    "lockoutUntil": "2026-04-01T12:30:00Z",
    "remainingSeconds": 540
  }
}
```

### Refresh Token Endpoint Request

```http
POST /api/auth/refresh
Content-Type: application/json
```

### Refresh Token Endpoint Response

```json
{
  "statusCode": 200,
  "data": {
    "expiresIn": 3600,
    "refreshAt": 2700,
    "csrfToken": "rotated_csrf_token"
  }
}
```

## Navigation Routes

### Core Routes

| Route | Component | Authentication | Purpose |
|-------|-----------|-----------------|---------|
| `/` | Redirect | - | Redirects to `/dashboard` if authenticated |
| `/login` | LoginScreen | No | Login page for unauthenticated users |
| `/dashboard` | DashboardPage | Yes | Main application dashboard |
| `/documents` | DocumentsPage | Yes | Document upload and management |
| `/payments` | PaymentDashboard | Yes | Bill payment dashboard |
| `/analytics` | AnalyticsPage | Yes | Financial history and reports |
| `/settings` | SettingsPage | Yes | User and project settings |

### Route Guard Implementation

All authenticated routes require:
1. Valid authentication cookies transmitted with credentials-enabled requests
2. Current project context set in user state
3. Proper role permissions for the action

## Component Hierarchy

```
App
├── AuthProvider (wraps entire app)
├── Router (react-router-dom)
│   ├── LoginRoute
│   │   └── LoginScreen
│   └── ProtectedLayout
│       ├── Sidebar (navigation)
│       ├── HamburgerMenu (mobile)
│       └── MainContent (outlet for child routes)
│           ├── DashboardPage
│           ├── DocumentsPage
│           ├── PaymentDashboard
│           ├── AnalyticsPage
│           └── SettingsPage
├── ThemeProvider
└── ErrorBoundary
```

## Local Storage Considerations

### Data Persisted

- Session metadata (expiry, refresh timing, lockout state)
- One-time draft restore payload with short TTL
- User preference (theme: light/dark)
- Active project ID

### Data NOT Persisted

- Unauthenticated navigation state
- Temporary form state (cleared on page reload)
- Error messages (cleared on navigation)

## Browser API Usage

| API | Purpose | Storage | Duration |
|-----|---------|---------|----------|
| sessionStorage | Store credentials for current session | Browser memory | Until browser closed |
| localStorage | Store persistent preferences (theme) | Browser disk | Indefinite |
| Cookies (HTTP-only) | Store refresh token securely | Browser disk + HTTP transmission | Expires on server-side |

## Performance Considerations

- Minimize re-renders by memoizing context value
- Lazy-load route components for code splitting
- Cache BFF responses when appropriate (e.g., user profile)
- Debounce token refresh to avoid thundering herd
