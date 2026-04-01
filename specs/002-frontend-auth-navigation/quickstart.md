# Quickstart: Frontend Authentication & Navigation System

**Feature**: 002-frontend-auth-navigation

## For Developers

### Prerequisites

- Frontend development environment set up (Node.js 18+, npm/yarn)
- Backend running with BFF service (`make backend/bff/dev`)
- Default user seeded in backend database

### Quick Start

```bash
# Install dependencies
cd frontend && npm install

# Start the frontend development server
make frontend/dev

# Opens http://localhost:3000

# The login screen loads automatically with:
# - Username: demo (or configured via VITE_DEFAULT_USERNAME)
# - Password: demo123 (or configured via VITE_DEFAULT_PASSWORD)
# Note: defaults come from .env.local and are non-production only

# Click "Sign in" to authenticate
```

### Environment Setup (Dev)

Create `frontend/.env.local`:

```bash
VITE_DEFAULT_USERNAME=demo
VITE_DEFAULT_PASSWORD=demo123
```

### Testing the Feature

```bash
# Run all frontend tests (unit + integration + a11y)
cd frontend && npm test

# Run tests with coverage
npm test -- --coverage

# Run integration tests with backend
make test/integration
```

### Step-by-Step Test Scenarios

**Login flow**
1. Open `http://localhost:3000`
2. Verify login screen is displayed as the first view
3. Verify username and password are pre-filled from env vars
4. Click "Sign in" — observe skeleton loading state briefly
5. Verify redirect to `/dashboard` occurs after successful auth

**Sidebar navigation**
1. Log in successfully
2. Verify sidebar appears on the left with 6 items: Dashboard, Documents, Bills, Payments, Analytics, Settings
3. Click each item — verify URL changes and active item is highlighted
4. Use browser Back/Forward — verify correct page loads

**Mobile responsive**
1. Open DevTools → set viewport to 375px wide
2. Verify sidebar is hidden; hamburger button is visible (top-left)
3. Click hamburger → verify sidebar slides in
4. Click a nav item → verify sidebar collapses and page loads

**Token refresh (silent)**
1. Log in successfully
2. In DevTools Network tab, filter for `/api/auth/refresh`
3. Wait for ~75% of `expires_in` to elapse
4. Observe a refresh request is made silently without interrupting the session
5. Verify subsequent API calls succeed with the new token

**Session persistence**
1. Log in successfully
2. Close the browser tab
3. Open `http://localhost:3000` again
4. Verify the session is restored if the `cfa:session` in localStorage is not expired

**Logout and session clear**
1. Log out from the application
2. Open DevTools → Application → Local Storage
3. Verify `cfa:session` key has been removed

**Error handling**
1. Enter invalid credentials and click "Sign in"
2. Verify an error message appears below the form
3. Submit 5+ times → verify lockout message with countdown appears
4. Verify the button is disabled during lockout

**Draft state restoration**
1. Log in and navigate to a page with a form
2. Trigger a session expiry (or use `localStorage.removeItem('cfa:session')` + refresh)
3. Log in again
4. Verify the draft restore modal appears if the application saved draft state

### Browser DevTools Checklist

- [ ] Application → Cookies: verify `access_token` has `HttpOnly`, `SameSite=Strict`
- [ ] Network: verify `/api/auth/login` POST on login
- [ ] Network: verify `/api/auth/refresh` POST fires at ~75% of token lifetime
- [ ] Application → Local Storage: verify `cfa:session` contains only metadata (no tokens)
- [ ] Accessibility: Tab through login form — username → password → Sign in button

## For QA / Testers

### Login Flow

1. Open the application in your browser
2. Verify the login screen displays
3. Verify username and password are pre-filled
4. Click "Sign in"
5. Verify the dashboard loads within 5 seconds
6. Verify the sidebar is visible with navigation items

### Navigation Flow

1. Click each sidebar item one by one
2. Verify the correct page loads for each item
3. Verify the URL changes to match the page
4. Verify the active menu item is highlighted
5. Use browser back/forward buttons and verify navigation works

### Mobile Responsive

1. Resize browser to mobile width (< 768px)
2. Verify sidebar is hidden
3. Verify hamburger menu button appears
4. Click hamburger to show sidebar
5. Click a navigation item
6. Verify sidebar hides and page loads

## For Product Managers

### User Value

- **Login is frictionless**: Default credentials eliminate manual entry for development
- **Navigation is discoverable**: Sidebar shows all available features in one place
- **Session is seamless**: Token refresh prevents unexpected logout
- **Design is responsive**: Works on phones, tablets, and desktops

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Login fails with 401 | Verify BFF is running and default user is seeded in DB |
| Login fails with 429 | Wait for lockout window to expire; verify failed-attempt policy and server clock |
| Sidebar doesn't appear | Check browser console for errors; verify React Router is initialized |
| Auth not persisted | Check cookie domain/path/secure flags and `credentials: include` on client requests |
| Mobile menu not working | Resize browser to trigger media query; check viewport meta tag |

## Next Steps

- Run `/speckit.plan` to generate detailed design and task breakdown
- Implement components according to the design phase output
- Execute tasks from tasks.md
- Run integration tests before marking feature complete
