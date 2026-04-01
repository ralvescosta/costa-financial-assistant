# Research: Frontend Authentication & Navigation System

**Feature**: 002-frontend-auth-navigation  
**Created**: 2026-04-01

## Decision 1: Authentication Transport and Storage

**Decision**: Use BFF-issued HTTP-only cookies (`SameSite=Strict`) for auth session transport; frontend sends authenticated requests with `credentials: include`.

**Rationale**:
- Prevents JavaScript token access and reduces XSS credential exfiltration risk.
- Fits BFF-centric session control and cookie rotation strategy.
- Matches clarified requirement set in spec.

**Alternatives considered**:
- Bearer token in browser storage: rejected due to XSS risk and contradiction with clarified constraints.
- Memory-only token model: rejected due to weak refresh/reload UX for this feature scope.

## Decision 2: Refresh Timing Strategy

**Decision**: Trigger refresh at 75% of `expires_in` lifetime.

**Rationale**:
- Provides conservative pre-expiry buffer.
- Reduces late-refresh race conditions and user-visible 401 retries.
- Explicitly aligned with clarification answers.

**Alternatives considered**:
- 90% lifetime: rejected to avoid narrow safety window.
- On-401-only refresh: rejected due to noisy UX and avoidable request failures.

## Decision 3: Default Credential Source

**Decision**: Default login auto-fill values come from frontend environment configuration and are enabled only for non-production environments.

**Rationale**:
- Eliminates hardcoded credentials in source code.
- Keeps local/dev onboarding fast while preserving production safety.

**Alternatives considered**:
- Hardcoded defaults in source: rejected for security and governance reasons.
- Runtime bootstrap endpoint for defaults: deferred; unnecessary complexity for this feature.

## Decision 4: Refresh-Failure Recovery

**Decision**: On refresh failure, clear auth session, redirect to login, and preserve in-progress draft state with short TTL for one-time restore after re-login.

**Rationale**:
- Preserves user work while maintaining secure forced re-authentication.
- Limits stale state risk via TTL + one-time restore semantics.

**Alternatives considered**:
- Discard all in-progress state: rejected due to poor UX.
- Prompt user before redirect: rejected due to interruption and failure-path complexity.

## Decision 5: Loading Feedback Pattern

**Decision**: Use skeleton placeholders for login and protected-page initial loading states.

**Rationale**:
- Improves perceived responsiveness.
- Gives clear feedback that background work is active.
- Supports measurable timing criteria in CI checks.

**Alternatives considered**:
- Spinner-only loading: rejected due to lower perceived progress quality.
- No explicit loading UI: rejected due to poor feedback clarity.

## Decision 6: Brute-Force Protection UX Contract

**Decision**: Enforce temporary lockout after 5 failed login attempts in 15 minutes; frontend displays lockout message with remaining time from server metadata.

**Rationale**:
- Improves baseline authentication security posture.
- Produces deterministic UX and testable behavior.

**Alternatives considered**:
- No lockout policy: rejected for security risk.
- CAPTCHA-first model: deferred as out-of-scope complexity for this phase.

## Decision 7: Navigation Responsiveness and Accessibility

**Decision**: Desktop persistent sidebar, mobile hidden drawer via hamburger, keyboard-accessible nav controls and active-route semantics.

**Rationale**:
- Aligns with feature scope and mobile-first usage.
- Supports accessibility requirements (focus order, labels, active state semantics).

**Alternatives considered**:
- Top navigation only: rejected as it reduces discoverability for multi-screen flows.
- Desktop drawer hidden by default: rejected for unnecessary interaction cost.
