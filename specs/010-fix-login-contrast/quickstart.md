# Quickstart: Dark Mode Login Button Contrast

## Goal

Verify that the login page and any other shared solid primary actions remain readable in dark mode on first render and throughout interaction states.

## Prerequisites

- Node.js environment available for the `frontend/` workspace
- Frontend dependencies installed
- Existing theme toggle or dark-mode class wiring available in the application

## Local Run

```bash
cd /home/ralvescosta/Desktop/Insync/Rafael/costa-financial-assistant/frontend
npm install
npm run dev
```

Open the app in the browser and navigate to the login page.

## Manual Validation Flow

Use this checklist while validating the fix locally:

- [ ] Enable **dark mode** before or while opening the login page.
- [ ] Confirm the `Sign in` button text is readable **immediately on first render**.
- [ ] Hover over the button and confirm the label stays readable.
- [ ] Use the `Tab` key to focus the button and confirm the focus treatment remains visible.
- [ ] Trigger a loading or disabled state and verify the label remains legible while the control still looks inactive or busy.
- [ ] Review the other shared solid-action surfaces that use the same theme contract:
  - `frontend/src/app/ErrorBoundary.tsx`
  - `frontend/src/components/DraftRestoreModal.tsx`
  - `frontend/src/components/ProjectSwitcher.tsx`
- [ ] Switch back to **light mode** and confirm the primary action still looks correct there as well.

## Validation Notes

- 2026-04-04: `npm run test` passed in `frontend/` (`24` files, `108` tests).
- 2026-04-04: `npm run typecheck` passed in `frontend/`.
- 2026-04-04: `npm run lint` completed successfully in `frontend/` after restoring the missing ESLint config.

## Recommended Validation Commands

```bash
cd /home/ralvescosta/Desktop/Insync/Rafael/costa-financial-assistant/frontend
npm run test
npm run lint
npm run typecheck
```

## Expected Result

- No low-contrast `Sign in` button on initial dark-mode load
- No regression in hover, focus, loading, or disabled states
- No light-mode regression on the same primary-action surfaces
- Shared dark-mode primary button styling remains consistent across affected components