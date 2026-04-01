# Feature Specification: Frontend Authentication & Navigation System

**Feature Branch**: `002-frontend-auth-navigation`
**Created**: 2026-04-01
**Status**: Draft
 **Input**: User description: "The frontend must have a hamburger menu and a sidebar navigation showing all screens. The first screen of the frontend must be a login screen which will auto-fill a default user and password as we have been doing in the backend. In the backend we will have a default user already saved in the database that was saved using the migration seed. The login screen must hit the BFF and BFF must do the things it needs to do to provide an access_token, expires_in, refresh_token."

 ## Clarifications

 ### Session 2026-04-01

 - Q: How should authentication tokens be stored in the browser (session storage, HTTP-only cookies, or hybrid)? → A: HTTP-only cookies with SameSite=Strict for access token; HTTP-only cookie for refresh token. This prevents XSS token theft and is the industry security standard.
 - Q: When should access-token refresh occur relative to `expires_in`? → A: Refresh at 75% of token lifetime to provide a conservative buffer while reducing late-refresh race conditions.
- Q: Where should default login credentials come from? → A: Frontend environment variables in non-production environments only; no hardcoded credentials in source.
- Q: What should happen to in-progress UI state when token refresh fails? → A: Preserve draft state locally with a short TTL and restore once after successful re-login.
- Q: What loading feedback should be shown during authentication and protected-page data fetches? → A: Show skeleton loading placeholders to indicate background processing is in progress.
- Q: What login brute-force protection should be required? → A: Temporary lockout after 5 failed attempts in 15 minutes.

 ## User Scenarios & Testing *(mandatory)*

### User Story 1 - Default User Auto-Filled Login (Priority: P0)

On first load or after session expiry, the user is presented with a login screen that displays a pre-filled default username and password. This allows bootstrapping the application for local development and initial demonstration without manual credential entry. The user can override these fields if desired, but the happy path is to accept the defaults and proceed.

**Why this priority**: This is the foundational UX entry point. Without a working login flow, no authenticated user can access any feature. This must work before any other frontend functionality becomes available. It is the critical path for initial development and testing.

**Independent Test**: Can be fully tested by launching the frontend, verifying the login screen displays with default credentials pre-filled, clicking "Login", and confirming the request reaches the BFF endpoint. Delivers immediate value as the authentication gateway.

**Acceptance Scenarios**:

1. **Given** the user accesses the application for the first time or after session expiry, **When** the page loads, **Then** the login screen is displayed as the primary view with username and password input fields.
2. **Given** the login screen is displayed, **When** the page renders, **Then** the default username and default password are automatically filled into their respective input fields.
3. **Given** the default credentials are pre-filled, **When** the user clicks the "Login" button without modification, **Then** an HTTP POST request is sent to the BFF login endpoint with the provided credentials.
4. **Given** the user wants to use different credentials, **When** they modify the username or password fields, **Then** the custom values are sent to the BFF instead of the defaults.
5. **Given** the user is on the login screen, **When** they click "Login", **Then** a loading state is displayed with skeleton placeholders to provide visual feedback that the request is in-flight.
6. **Given** valid credentials are sent to the BFF, **When** the BFF responds successfully, **Then** the frontend receives `access_token`, `expires_in`, and `refresh_token` in the response.
7. **Given** the frontend receives a successful authentication response, **When** the login response is processed, **Then** authentication cookies are securely established and the user is navigated to the main application dashboard.
8. **Given** invalid credentials are sent to the BFF, **When** the BFF responds with an authentication failure, **Then** the frontend displays a clear error message (e.g., "Invalid username or password") without clearing pre-filled defaults.
9. **Given** 5 failed login attempts occur within 15 minutes for the same identity, **When** another login is attempted during the lockout window, **Then** authentication is blocked and the frontend displays a temporary lockout message with remaining lockout time.

---

### User Story 2 - Sidebar Navigation with Screen Routes (Priority: P1)

After successful login, the user is presented with a persistent sidebar on the left that displays all available screens/pages in the application. The sidebar remains visible across all subsequent views, allowing the user to navigate between features without a page reload. A hamburger menu button is available on mobile or compact viewports to show/hide the sidebar.

**Why this priority**: Navigation is critical UX infrastructure. Without a consistent way to explore features, users cannot access any functionality beyond the first screen they land on. This must be implemented early to support rapid feature development and testing.

**Independent Test**: Can be fully tested by logging in successfully, verifying the sidebar appears, clicking on each navigation item, and confirming the correct screen/page content loads without a full page reload.

**Acceptance Scenarios**:

1. **Given** the user has logged in successfully, **When** the dashboard or first authenticated screen loads, **Then** a sidebar navigation panel is displayed on the left side of the screen, persisting across all navigation actions.
2. **Given** the sidebar is displayed, **When** the screen renders, **Then** all available screens/features are listed as selectable menu items (e.g., Dashboard, Documents, Payments, Settings, etc.).
3. **Given** the user is viewing the sidebar menu, **When** they click on a menu item, **Then** the corresponding screen or page loads without a full page refresh, and the URL is updated to reflect the new view.
4. **Given** the user is on a particular screen, **When** the sidebar renders, **Then** the current screen's menu item is visually highlighted to indicate the active route.
5. **Given** the viewport is wider than a tablet breakpoint, **When** the page loads, **Then** the sidebar is displayed by default and the hamburger menu button is hidden or non-functional.
6. **Given** the viewport is smaller than a tablet breakpoint (mobile), **When** the page loads, **Then** the sidebar is initially hidden and a hamburger menu button is visible at the top-left of the screen.
7. **Given** the user is on a mobile viewport and the sidebar is hidden, **When** they click the hamburger menu button, **Then** the sidebar slides in or appears as an overlay without covering the entire content area.
8. **Given** the user is viewing the sidebar on a mobile viewport and selects a menu item, **When** the navigation completes, **Then** the sidebar is automatically hidden to maximize content space.

---

### User Story 3 - Token Refresh and Session Persistence (Priority: P1)

The application stores the access token securely and automatically refreshes it before expiry using the `refresh_token`. If a request is made with an expired access token, the application attempts to refresh the token silently without interrupting the user experience. If refresh fails, the user is redirected to the login screen.

**Why this priority**: Token lifecycle management ensures the user stays authenticated across their session without repeated login prompts. This is essential for a smooth, uninterrupted user experience. It must work reliably in the background before implementing feature-specific authenticated endpoints.

**Independent Test**: Can be fully tested by logging in, waiting for the access token to approach expiry, making an authenticated API call, verifying the token is refreshed in the background, and confirming the API call succeeds without the user needing to log in again.

**Acceptance Scenarios**:

1. **Given** the user has logged in and received an access token with an `expires_in` value, **When** the frontend processes the login response, **Then** a timer is set to refresh the token at 75% of the token lifetime before expiry.
2. **Given** the refresh token is available and not yet expired, **When** the refresh timer triggers, **Then** an HTTP POST request is sent to the BFF refresh endpoint with the `refresh_token`.
3. **Given** the BFF refresh endpoint responds with a new access token and updated `expires_in`, **When** the response is received, **Then** the new token is stored, replacing the old one, and the refresh timer is reset.
4. **Given** the user makes an authenticated API request with an expired access token, **When** the BFF responds with a 401 Unauthorized status, **Then** the frontend intercepts the response and attempts a token refresh.
5. **Given** the token refresh succeeds after a failed 401 request, **When** the new token is obtained, **Then** the original failed request is automatically retried with the new token.
6. **Given** the token refresh fails (invalid or expired refresh token), **When** the refresh endpoint returns an error, **Then** the user is redirected to the login screen with a message indicating session expiry, while in-progress draft state is preserved locally for one-time restore after re-login.
7. **Given** the user has a valid session and closes the browser, **When** they return and the application reloads, **Then** the application checks for stored tokens and, if valid, continues the session without requiring re-login.

---

## Functional Requirements *(mandatory)*

**Login & Authentication Flow**

- **FR-001**: The login screen MUST display two input fields (username and password) with pre-filled default values sourced from frontend environment configuration for non-production environments (local/dev) and MUST NOT hardcode credentials in source code.
- **FR-002**: The login screen MUST include a "Login" button that sends an HTTP POST request to the BFF login endpoint (`POST /api/auth/login`) with the provided credentials.
- **FR-003**: The login request MUST include the username and password in the request body using standard form or JSON encoding as defined in the BFF contract.
- **FR-004**: The BFF login endpoint MUST return a JSON response containing at least `access_token` (string), `expires_in` (number in seconds), and `refresh_token` (string) on successful authentication.
 - **FR-005**: The frontend MUST securely store the `access_token` in an HTTP-only cookie with `SameSite=Strict` attribute to prevent XSS token theft and CSRF attacks. The cookie MUST NOT be accessible to JavaScript.
 - **FR-006**: The frontend MUST securely store the `refresh_token` in an HTTP-only cookie with `SameSite=Strict` attribute, isolated from the access token and only transmitted on refresh endpoint requests.
- **FR-007**: The frontend MUST implement a token refresh mechanism that automatically refreshes the access token at 75% of token lifetime using the `refresh_token`.
- **FR-008**: The frontend MUST send authenticated API requests with credentials enabled so HTTP-only authentication cookies are transmitted on each request according to cookie policy.
- **FR-009**: The frontend MUST handle 401 Unauthorized responses by automatically attempting a token refresh and retrying the original request, provided the refresh succeeds.
- **FR-010**: If token refresh fails, the frontend MUST clear authentication tokens, redirect the user to the login screen with a session-expiry message, and preserve in-progress draft UI state in local client storage with a TTL of 5 minutes for one-time restore after successful re-login (draft is discarded if not restored within the window).
- **FR-033**: The authentication flow MUST enforce temporary brute-force protection by blocking login after 5 failed attempts within a 15-minute window, and MUST return lockout metadata so the frontend can show remaining lockout time.

**Navigation & Layout**

- **FR-011**: After successful login, the application MUST render a persistent sidebar on the left side of all authenticated pages.
- **FR-012**: The sidebar MUST display a complete list of all available screens/features as clickable navigation items (e.g., Dashboard, Documents, Bills, Payments, Analytics, Settings).
- **FR-013**: Each sidebar navigation item MUST correspond to a distinct route or screen in the application, and clicking the item MUST navigate to that route using client-side routing (no full page reload).
- **FR-014**: The sidebar MUST visually highlight the navigation item corresponding to the currently active route (e.g., bold text, background color, or icon change).
- **FR-015**: The sidebar MUST update the browser URL when a navigation item is clicked, enabling browser back/forward navigation and shareable links for each screen.
- **FR-016**: A hamburger menu icon MUST be displayed in the top-left corner of the application on viewports narrower than 1024px (tablet and mobile).
- **FR-017**: On mobile viewports (narrower than 768px), the sidebar MUST be hidden by default and MUST be displayed as a slide-in overlay or drawer when the hamburger menu is clicked.
- **FR-018**: On mobile viewports, clicking a navigation item in the sidebar MUST close the sidebar after navigation completes.
- **FR-019**: On tablet viewports (768px to 1023px), the sidebar MAY collapse to show only icons or may be hidden by default, with a hamburger menu to toggle its visibility.
- **FR-020**: The main content area MUST occupy the remaining horizontal space after the sidebar (on desktop) or the full viewport width (on mobile when sidebar is hidden).

**UI/UX and Responsiveness**

- **FR-021**: The login screen MUST be fully responsive and centered on the viewport, displaying correctly on mobile, tablet, and desktop screens.
- **FR-022**: The login screen MUST display clear, accessible input labels for username and password fields.
- **FR-023**: The password input field MUST mask the entered text for security (display dots or asterisks instead of plaintext).
- **FR-024**: The "Login" button MUST be disabled and display a loading indicator while an authentication request is in-flight.
- **FR-032**: During authentication and protected-page data fetches, the frontend MUST display skeleton placeholders until content is ready so users perceive active background processing.
- **FR-025**: On login failure, the frontend MUST display a clear, user-friendly error message without clearing the default pre-filled credentials (allowing the user to retry).
- **FR-026**: The application MUST apply the design token system (light/dark theme tokens) from `frontend/src/styles/tokens.ts` consistently across the login screen, navigation, and all authenticated pages. *(Prerequisite: Design tokens must be created or already exist before Phase 1 begins.)*
- **FR-027**: The application MUST support theme switching (light/dark mode) that persists across sessions using local storage or a user preference setting.
- **FR-028**: All navigation items MUST be accessible via keyboard navigation (Tab key) and screen readers for compliance with accessibility standards.

**Authentication State Management**

- **FR-029**: The frontend application MUST maintain a centralized authentication state that tracks whether the user is logged in, the current access token, the refresh token, and the token expiry time.
- **FR-030**: All authenticated API clients or request interceptors MUST automatically send requests with credentials enabled to include HTTP-only authentication cookies for every authenticated request.
- **FR-031**: The frontend MUST use a single, canonical source of truth for the active project context (if multi-tenancy is active) and include it in authenticated requests as required by the BFF.

## Key Entities

- **AuthenticationContext**: Contains the logged-in user's identity, access token, refresh token, token expiry timestamp, and active project context.
- **User**: Represents a logged-in user with at minimum a user ID, username, and any optional attributes (e.g., full name, preferred theme).
- **NavigationItem**: A single entry in the sidebar menu representing a screen or feature, with properties like label, route path, icon, and active state.
- **Session**: The browser-local or server-side session state tracking the user's login status and authentication tokens.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001** [CI]: A user can navigate from the login screen to the main dashboard in under 5 seconds on a simulated 4G mobile connection (throttled to 4 Mbps down / 1.5 Mbps up / 40ms latency, per WebPageTest profile) after entering or accepting default credentials. *Verified by*: US1 integration test (T001) measuring time from login button click to successful dashboard render using consistent network throttle profile.
- **SC-002** [CI]: The sidebar navigation correctly highlights the active route 100% of the time when navigating between screens. *Verified by*: US2 integration test (T010) clicking each sidebar item and confirming the corresponding page content loads and the item is highlighted.
- **SC-003** [CI]: Token refresh succeeds silently in the background before an access token expires, with zero interruption to the user experience. *Verified by*: US3 integration test (T020) waiting for token expiry and making an authenticated API call, confirming the call succeeds without re-login.
- **SC-004** [CI]: Invalid credentials result in a clear error message being displayed to the user within 3 seconds of login attempt submission. *Verified by*: US1 integration test (T003) submitting invalid credentials and confirming error message appears.
- **SC-005** [CI]: Sidebar navigation is fully hidden on mobile viewports narrower than 768px and displays correctly as an overlay when the hamburger menu is clicked. *Verified by*: US2 integration test (T012) simulating mobile viewport resize and confirming sidebar visibility toggles correctly.
- **SC-006** [CI]: The application remains in an authenticated state after a browser refresh when a valid refresh token is present, eliminating the need for re-login on page reload. *Verified by*: US3 integration test (T025) refreshing the page mid-session and confirming the user remains logged in.
- **SC-007** [CI]: The login screen is accessible and usable by keyboard-only navigation (Tab key to navigate fields and buttons). *Verified by*: US1 accessibility test (T005) using keyboard navigation only to log in.
- **SC-008** [CI]: All sidebar navigation links correspond to valid, implemented routes and load the correct screen content without errors. *Verified by*: US2 integration test (T011) clicking every navigation item and confirming HTTP 200 response and expected page content.
- **SC-009** [CI]: During login and protected-page loading, skeleton placeholders are displayed within 300ms (measured as DOM mount time from page load start) and replaced by final content when data is available. *Verified by*: UI integration test (T030) asserting skeleton DOM elements exist within 300ms and are replaced with final content when async data completes.
- **SC-010** [CI]: Brute-force protection blocks authentication after 5 failed login attempts in 15 minutes and allows authentication again after lockout expires. *Verified by*: auth integration test (T031) asserting lockout trigger, lockout message payload, and post-window recovery.

## Assumptions

- **A1**: A default user with username and password is already seeded in the backend database via migration, as stated in the feature description.
- **A2**: The BFF login endpoint is already implemented and accessible at `POST /api/auth/login` and accepts credentials in a standard format (username/password JSON or form-encoded).
- **A3**: The BFF refresh endpoint is already implemented at `POST /api/auth/refresh` and accepts a refresh_token parameter.
- **A4**: The frontend will use React with `react-router-dom` for routing and client-side navigation, as per project conventions.
 - **A5**: Authentication tokens will be stored in HTTP-only cookies with SameSite=Strict attribute (industry security standard); no custom cryptographic storage is required. The BFF must set these cookies in response headers with appropriate expiry times.
- **A6**: The design token system and theme switching infrastructure (light/dark mode) MUST be implemented in `frontend/src/styles/tokens.ts` before Phase 1 begins. If not already present, create a blocking setup task to generate semantic tokens for light/dark themes (e.g., colorPrimary, colorSurface, colorTextPrimary, colorDanger, etc.) and Tailwind CSS variable bindings. This is a hard prerequisite for FR-026 and FR-027.
- **A7**: The application has a single active project context for now; multi-project switching is out of scope for this phase.
- **A8**: Default login auto-fill is enabled only in non-production environments; production builds must not expose default credentials.

## Notes

- The login flow establishes the foundation for feature delivery. All subsequent features depend on authenticated access and proper authorization context.
- Sidebar navigation provides the UX scaffolding for rapid feature discovery and testing during development.
- Token refresh in the background ensures a frictionless experience and reduces the need for manual re-authentication.
- The pre-filled default credentials are acceptable for development but should be configurable or disabled in production deployments.
