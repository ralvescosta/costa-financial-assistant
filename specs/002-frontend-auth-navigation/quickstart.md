# Quickstart: Frontend Authentication & Navigation System

**Feature**: 002-frontend-auth-navigation

## For Developers

### Prerequisites

- Frontend development environment set up (Node.js 18+, npm/yarn)
- Backend running with BFF service (`make backend/bff/dev`)
- Default user seeded in backend database

### Quick Start

```bash
# Start the frontend development server
make frontend/dev

# Opens http://localhost:5173

# The login screen loads automatically with:
# - Username: demo (or configured default)
# - Password: demo123 (or configured default)
# Note: defaults come from frontend env config and are non-production only

# Click "Login" to authenticate
```

### Testing the Feature

```bash
# Run frontend tests
npm run test

# Run integration tests with backend
make test/integration

# Test token refresh behavior
# 1. Log in successfully
# 2. Wait for token refresh (check console logs)
# 3. Make an authenticated API call
# 4. Verify it succeeds without re-login
```

### Browser DevTools Checklist

- [ ] Check Application > Cookies for auth cookies set by BFF (`HttpOnly`, `SameSite=Strict`)
- [ ] Check Network tab for login request and response
- [ ] Check Console for authentication logs
- [ ] Test sidebar navigation by clicking items
- [ ] Test responsive design by resizing the browser window
- [ ] Confirm skeleton placeholders appear during login/protected-page loading

## For QA / Testers

### Login Flow

1. Open the application in your browser
2. Verify the login screen displays
3. Verify username and password are pre-filled
4. Click "Login"
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
