# Feature Specification: Financial Bill Organizer

**Feature Branch**: `001-financial-bill-organizer`
**Created**: 2026-03-30
**Status**: Draft
**Input**: User description: "Build an application that helps organize financial life by ingesting bills (credit card, energy, internet, etc.) and bank account balance statements as PDF files, extracting structured data asynchronously, enabling bill payment tracking via Pix QR codes and barcodes, performing cross-referencing with account statements, and presenting a financial history dashboard."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Upload and Classify a PDF Document (Priority: P1)

The user uploads a PDF file (a bill or a bank account balance statement) to the application. The system asks whether the document is a bill or an account balance statement. If it is a bill, the system asks which type of bill it represents (e.g., credit card, energy, internet). If it is an account balance statement, the system asks which registered bank account label it belongs to. The document is stored securely and queued for asynchronous analysis.

**Why this priority**: This is the entry point for all value the application provides. Without ingestion and classification, no analysis, cross-referencing, or payment assistance is possible. It is the foundational P1 MVP slice.

**Independent Test**: Can be fully tested by uploading a PDF, answering the classification questions, and verifying the document appears in the user's document list with the correct label and a "pending analysis" status. Delivers value immediately as a digital filing cabinet.

**Acceptance Scenarios**:

1. **Given** the user is on the upload screen, **When** they select a PDF file and submit it, **Then** the system stores the file and displays a classification dialog asking whether it is a bill or an account balance statement.
2. **Given** the user selects "bill" in the classification dialog, **When** they are asked for the bill type, **Then** the system presents available bill-type labels and requires the user to select or create one before confirming.
3. **Given** the user selects "account balance statement" in the classification dialog, **When** they are asked which bank account it belongs to, **Then** the system presents the user's registered bank account labels and requires selection before confirming.
4. **Given** the classification is confirmed, **When** the system acknowledges, **Then** the document is listed in the user's document library with its label, upload date, and the status "Pending Analysis".
5. **Given** the user attempts to upload a non-PDF file, **When** the system validates the upload, **Then** the system rejects the file and displays a clear error message without storing anything.

---

### User Story 2 - Asynchronous PDF Analysis and Data Extraction (Priority: P2)

After a PDF is uploaded and classified, the system analyses it asynchronously in the background. For a bill document, the system extracts the due date, total amount due, Pix QR code payload, and barcode string. For an account balance statement, the system extracts every transaction line (date, description, amount, credit/debit indicator) and stores them for later cross-referencing.

**Why this priority**: Data extraction transforms raw documents into structured, actionable financial data. Without it, the payment assistant and cross-referencing features cannot function. It is P2 because upload (P1) must work first.

**Independent Test**: Can be fully tested by uploading a known bill PDF, waiting for background processing to complete, and verifying the extracted due date, amount, QR code, and barcode are displayed correctly on the document detail screen.

**Acceptance Scenarios**:

1. **Given** a bill PDF has been uploaded and classified, **When** the asynchronous analysis completes successfully, **Then** the document status changes to "Analysed" and the extracted due date, amount, Pix QR code, and barcode are stored and visible on the document detail screen.
2. **Given** an account balance statement PDF has been uploaded and classified, **When** the asynchronous analysis completes successfully, **Then** the document status changes to "Analysed" and all transaction lines with their dates, descriptions, and amounts are stored and visible.
3. **Given** the PDF analysis fails (corrupt file, unrecognised format, extraction errors), **When** the error is detected, **Then** the document status changes to "Analysis Failed", the user is notified, and the partial data is not persisted as complete.
4. **Given** a bill PDF has been analysed, **When** the user views the document detail, **Then** the Pix QR code is displayed as a scannable image and the barcode string is displayed in human-readable form alongside a copy button.
5. **Given** the system is processing a large PDF, **When** the user checks the document status, **Then** a progress indicator shows "Processing" until completion.

---

### User Story 3 - Bank Account Registration (Priority: P2)

The user registers labelled bank account identifiers so that account balance statement PDFs can be attributed to the correct account. The system stores only non-sensitive labels (e.g., "Nubank Checking", "Bradesco Savings") — no account numbers, passwords, or sensitive financial identifiers are stored.

**Why this priority**: Required before account balance PDFs can be meaningfully classified and cross-referenced. Kept P2 (same tier as extraction) because it is a short setup flow that enables the full pipeline.

**Independent Test**: Can be tested by registering two bank account labels, uploading an account balance PDF, selecting one of the registered labels, and confirming the document is correctly attributed to that label in the document list.

**Acceptance Scenarios**:

1. **Given** the user navigates to account settings, **When** they create a new bank account label with a name, **Then** the label is saved and appears in the list of registered accounts.
2. **Given** at least one account exists, **When** the user uploads an account balance PDF, **Then** the classification dialog presents the registered account labels as selectable options.
3. **Given** the user attempts to register an account with an empty or duplicate name, **When** the system validates the input, **Then** the system rejects the entry and displays an appropriate error.
4. **Given** the user deletes a bank account label that has documents attributed to it, **When** the deletion is confirmed, **Then** the system warns the user that attributed documents will become unlinked and requires explicit confirmation.

---

### User Story 4 - Preferred Payment Day and Bill Payment Dashboard (Priority: P3)

The user specifies their preferred day of the month to review and pay bills. The application presents a payment dashboard listing all outstanding bills for the current cycle, each displaying the due date, amount, Pix QR code (scannable), and barcode. The user can mark individual bills as paid directly from this screen. The dashboard filters to show only unpaid bills by default.

**Why this priority**: This is the primary daily-use value proposition — the reason the user will open the app. It requires P1 (upload) and P2 (extraction) to be working.

**Independent Test**: Can be tested end-to-end by uploading and analysing two bill PDFs, setting a preferred payment day, opening the payment dashboard, verifying both bills appear with correct extracted data, marking one as paid, and confirming it disappears from the outstanding list.

**Acceptance Scenarios**:

1. **Given** the user has set a preferred payment day, **When** they open the payment dashboard, **Then** all bills with a due date in the current monthly cycle that are not yet marked as paid are listed, sorted by due date ascending.
2. **Given** a bill is listed on the payment dashboard, **When** the user views it, **Then** the Pix QR code is displayed as a scannable image and the barcode string is visible with a copy-to-clipboard action.
3. **Given** the user marks a bill as paid on the dashboard, **When** the action is confirmed, **Then** the bill moves out of the outstanding list and is recorded as paid on that date.
4. **Given** the user has not yet set a preferred payment day, **When** they open the payment dashboard, **Then** the system prompts them to set one before proceeding.
5. **Given** a bill's due date has passed and it is still unpaid, **When** the dashboard renders, **Then** the bill is visually flagged as overdue.

---

### User Story 5 - Cross-Reference Account Statement with Bills (Priority: P4)

When the user uploads a bank account balance statement and its analysis completes, the system automatically attempts to cross-reference each transaction line in the statement against the stored bills for the same period. Matched transactions are linked to their corresponding bill, giving the user a reconciliation view showing which bills were paid according to the statement and which bills remain unmatched (potentially forgotten or paid via another channel).

**Why this priority**: This is an advanced analytical feature that requires all prior stories to be working. It provides the financial oversight value but is not needed for the basic bill-payment workflow.

**Independent Test**: Can be tested by uploading a set of bills for a given month, then uploading an account statement for the same month, waiting for cross-reference analysis, and verifying that bills matching statement transactions are shown as "Confirmed Paid" and unmatched bills are shown as "Unconfirmed".

**Acceptance Scenarios**:

1. **Given** an account balance statement has been analysed, **When** cross-referencing completes, **Then** each statement transaction that matches a known bill is linked to that bill and both are shown as reconciled.
2. **Given** cross-referencing completes, **When** the user views the reconciliation summary, **Then** bills with no matching statement transaction are highlighted as "Unconfirmed Payment".
3. **Given** a single statement transaction could match multiple bills (ambiguous match), **When** the cross-reference engine encounters the ambiguity, **Then** the system presents the candidate bills to the user for manual selection rather than auto-linking.
4. **Given** the user manually links a transaction to a bill, **When** the link is saved, **Then** it is recorded as a user-confirmed reconciliation and appears in the history.

---

### User Story 6 - Financial History Dashboard (Priority: P5)

The user can access a dashboard that visualises their complete financial history across all ingested months. The dashboard shows aggregated expenditure over time, bill-category breakdowns, month-over-month comparisons, and payment compliance rate (bills paid on time vs. overdue). All historical months are preserved.

**Why this priority**: Lowest priority because it is a reporting/analytics layer on top of already-stored data. Maximum value is unlocked after several months of usage.

**Independent Test**: Can be tested by seeding multiple months of bill and statement data, then verifying the dashboard renders correct totals, category breakdowns, and trend charts for the seeded period.

**Acceptance Scenarios**:

1. **Given** the user opens the financial history dashboard, **When** data exists for multiple months, **Then** a timeline view shows monthly total expenditure for at least the last 12 months.
2. **Given** the dashboard is rendered, **When** the user selects a specific month, **Then** a category breakdown shows total spend per bill type (credit card, energy, internet, etc.) for that month.
3. **Given** the dashboard is rendered, **When** the user views the payment compliance section, **Then** a percentage is shown for bills paid on or before the due date vs. overdue for each month.
4. **Given** no data exists for a selected period, **When** the dashboard attempts to render, **Then** an informational state is shown prompting the user to upload documents for that period.

---

### Edge Cases

- What happens when a bill PDF contains no extractable Pix QR code or barcode (e.g., older format bills)? The system must store what it can extract and mark the missing fields explicitly as "not found" rather than failing the entire analysis.
- What happens when the same bill PDF is uploaded twice? The system must detect the duplicate (by file hash) and prompt the user rather than creating a duplicate record.
- What happens when a month has zero bills uploaded? The payment dashboard must show an empty state and the history dashboard must handle gaps gracefully.
- What happens when the cross-reference engine cannot find any matching transactions for a bill that the user marked as paid manually? The manual mark takes precedence; the system records it as a user-confirmed payment.
- What happens when an uploaded PDF has no machine-readable content (scanned image-only PDF)? The analysis must flag it as requiring OCR or unsupported and notify the user.
- What happens when a bank account label is referenced by documents but the user tries to delete it? The system must warn and prevent or require re-attribution.
- What happens when the document storage backend is temporarily unavailable during upload? The upload must fail gracefully with a user-friendly error; no orphaned metadata records must be left.

## Requirements *(mandatory)*

### Functional Requirements

**Document Ingestion**
- **FR-001**: System MUST accept PDF file uploads and store them in a configurable file-storage backend (not limited to a single provider; provider is selected via environment configuration).
- **FR-002**: System MUST prompt the user to classify each uploaded PDF as either a "bill" or an "account balance statement" before completing the upload flow.
- **FR-003**: System MUST prompt the user to select a bill-type label (e.g., credit card, energy, internet) when the document is classified as a bill.
- **FR-004**: System MUST prompt the user to select a registered bank account label when the document is classified as an account balance statement.
- **FR-005**: System MUST detect duplicate uploads by comparing file content hashes and warn the user before creating a duplicate record.
- **FR-006**: System MUST reject non-PDF file uploads before storing anything.

**Bank Account Management**
- **FR-007**: System MUST allow the user to create, list, update, and delete bank account labels.
- **FR-008**: System MUST store only non-sensitive label information for bank accounts (name only; no account numbers, routing numbers, or credentials).
- **FR-009**: System MUST warn the user and require explicit confirmation before deleting a bank account label that has attributed documents.

**Asynchronous PDF Analysis**
- **FR-010**: System MUST process uploaded PDFs asynchronously after classification, without blocking the user interface.
- **FR-011**: System MUST extract the following fields from bill PDFs: due date, total amount due, Pix QR code payload, barcode string.
- **FR-012**: System MUST extract all transaction lines from account balance statement PDFs, capturing at minimum: transaction date, description, amount, and credit/debit indicator.
- **FR-013**: System MUST update the document status to reflect the current analysis state: Pending, Processing, Analysed, or Analysis Failed.
- **FR-014**: System MUST notify the user when analysis completes or fails.
- **FR-015**: System MUST record which fields could not be extracted individually (e.g., "Pix QR code: not found") rather than failing the entire document when partial extraction is possible.

**Payment Dashboard**
- **FR-016**: System MUST allow the user to configure a preferred monthly payment day.
- **FR-017**: System MUST present a payment dashboard listing all outstanding (unpaid, analysed) bills for the current monthly cycle, sorted by due date.
- **FR-018**: System MUST display the Pix QR code as a scannable image and the barcode string with a copy action for each bill on the payment dashboard.
- **FR-019**: System MUST allow the user to mark a bill as paid from the payment dashboard, recording the date of the marking.
- **FR-020**: System MUST visually distinguish overdue bills (due date past, still unpaid) from upcoming bills on the dashboard.

**Cross-Referencing**
- **FR-021**: System MUST automatically attempt to cross-reference account statement transaction lines against stored bills for the same period when an account statement analysis completes.
- **FR-022**: System MUST link matched transactions to their corresponding bill and mark both as reconciled.
- **FR-023**: System MUST present ambiguous matches (one transaction matching multiple candidate bills) to the user for manual resolution rather than auto-linking.
- **FR-024**: System MUST allow the user to manually link a statement transaction to a bill.
- **FR-025**: System MUST preserve the full history of all ingested months without any automatic purging.

**Financial History Dashboard**
- **FR-026**: System MUST display a monthly expenditure timeline for all available historical data.
- **FR-027**: System MUST display per-month category breakdowns by bill type.
- **FR-028**: System MUST display a payment compliance rate (on-time vs. overdue) per month.

**Frontend Design System**
- **FR-029**: The frontend MUST implement a centralised design token system as the sole
  source for all colors, with primitive tokens defining the full palette and semantic
  tokens (e.g., `colorPrimary`, `colorSurface`, `colorTextPrimary`, `colorDanger`)
  being the only values referenced in components and style files.
- **FR-030**: All views and components MUST support both a light theme and a dark theme;
  every semantic color token MUST be defined for both themes.
- **FR-031**: The active theme MUST be user-selectable and the preference MUST be
  persisted across sessions. On first load with no stored preference, the application
  MUST honour the OS-level `prefers-color-scheme` setting.
- **FR-032**: Theme switching MUST apply instantly without a page reload.

### Key Entities

- **Document**: A stored PDF file with its classification (bill or account statement), storage reference, upload timestamp, file hash, analysis status, and attributed label.
- **BillRecord**: Extracted data from a bill document — due date, amount due, Pix QR code payload, barcode string, payment status, paid date, and link back to its Document.
- **StatementRecord**: Extracted data from an account balance statement — the set of transaction lines extracted from the PDF, linked to the Document and the bank account label.
- **TransactionLine**: A single line from an account statement — transaction date, description, amount, credit/debit indicator, and optional link to a matched BillRecord.
- **BankAccount**: A user-registered label for a bank account — name only, no sensitive data.
- **BillType**: A user-managed or system-provided label for a category of bill (e.g., "Credit Card – Nubank", "Energy – CEMIG", "Internet – Claro").
- **PaymentCycle**: Tracks which bills belong to a given month/cycle, the user's preferred payment day for that cycle, and the overall reconciliation status.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can upload a bill PDF, classify it, and see it appear in the document list with "Pending Analysis" status in under 10 seconds.
- **SC-002**: PDF analysis (data extraction) completes for a standard single-page bill within 60 seconds of upload.
- **SC-003**: The payment dashboard renders all outstanding bills for the current cycle in under 2 seconds, regardless of the number of historical documents stored.
- **SC-004**: 95% of standard bill PDFs with machine-readable content yield at least the due date and total amount upon extraction (based on common Brazilian bill formats).
- **SC-005**: Cross-reference matching links at least 80% of statement transactions to their corresponding bills when bill descriptions follow standard issuer naming conventions.
- **SC-006**: A user can open the payment dashboard and scan or copy a Pix QR code to complete a payment without leaving the app, reducing bill-payment effort by eliminating the need to locate physical or email copies of bills.
- **SC-007**: The financial history dashboard loads and renders 12 months of data in under 3 seconds.
- **SC-008**: Duplicate PDF uploads are detected 100% of the time before any data is persisted for the duplicate.

## Assumptions

- The initial target user is a single individual managing their own personal finances (multi-user / household sharing is out of scope for v1).
- Authentication and user account management already exist or will be addressed in a separate foundational feature; this specification assumes an authenticated user context.
- The application will be used primarily in Brazil; bill formats (Pix, boleto/barcode) are those common to Brazilian financial institutions.
- The file storage backend default for v1 is S3-compatible object storage; the architecture must be provider-agnostic so the provider can be changed via environment configuration without code changes.
- Image-only (non-machine-readable) scanned PDFs require OCR to extract data; OCR integration is out of scope for v1 — these documents will be flagged as "requires OCR" and the user notified.
- The user is responsible for registering their own bill-type and bank account labels; the system does not automatically create these from analysed content.
- All financial data is for personal tracking and display purposes only; the system does not initiate any payments, banking operations, or connections to financial institution APIs.
- Mobile-responsive web interface is in scope; native mobile app is out of scope for v1.
- The preferred payment day setting is a single global day per user, not per-bill-type (multi-day scheduling is out of scope for v1).
- Data retention: all ingested documents and extracted data are retained indefinitely unless explicitly deleted by the user.

## Design Constraints

### Color Palette & Theming

The application MUST use a design token system with two layers:

**Primitive tokens** (full palette, never referenced in components directly):

| Token | Light value | Dark value | Purpose |
|---|---|---|---|
| `blue50` | `#EFF6FF` | `#EFF6FF` | Lightest blue tint |
| `blue500` | `#3B82F6` | `#3B82F6` | Brand blue |
| `blue600` | `#2563EB` | `#60A5FA` | Hover / active blue |
| `green500` | `#22C55E` | `#22C55E` | Success base |
| `green600` | `#16A34A` | `#4ADE80` | Success emphasis |
| `yellow500` | `#EAB308` | `#EAB308` | Warning base |
| `red500` | `#EF4444` | `#EF4444` | Danger base |
| `red600` | `#DC2626` | `#F87171` | Danger emphasis |
| `sky400` | `#38BDF8` | `#38BDF8` | Info base |
| `neutral50` | `#F9FAFB` | `#111827` | Surface (inverted between themes) |
| `neutral100` | `#F3F4F6` | `#1F2937` | Elevated surface |
| `neutral200` | `#E5E7EB` | `#374151` | Border / divider |
| `neutral400` | `#9CA3AF` | `#6B7280` | Disabled / secondary text |
| `neutral700` | `#374151` | `#D1D5DB` | Secondary text |
| `neutral900` | `#111827` | `#F9FAFB` | Primary text (inverted) |
| `white` | `#FFFFFF` | `#1F2937` | Base background |
| `black` | `#000000` | `#000000` | Absolute black |
| `overlay` | `rgba(0,0,0,0.5)` | `rgba(0,0,0,0.7)` | Modal backdrop |

**Semantic tokens** (the only values components may reference):

| Semantic token | Light → Primitive | Dark → Primitive |
|---|---|---|
| `colorBackground` | `white` | `neutral50` |
| `colorSurface` | `neutral50` | `neutral100` |
| `colorSurfaceElevated` | `neutral100` | `neutral100` |
| `colorBorder` | `neutral200` | `neutral200` |
| `colorBorderFocus` | `blue500` | `blue600` |
| `colorDivider` | `neutral200` | `neutral200` |
| `colorTextPrimary` | `neutral900` | `neutral900` |
| `colorTextSecondary` | `neutral700` | `neutral700` |
| `colorTextDisabled` | `neutral400` | `neutral400` |
| `colorTextInverse` | `white` | `neutral900` |
| `colorPrimary` | `blue500` | `blue500` |
| `colorPrimaryHover` | `blue600` | `blue600` |
| `colorPrimaryActive` | `blue600` | `blue600` |
| `colorSuccess` | `green500` | `green500` |
| `colorSuccessEmphasis` | `green600` | `green600` |
| `colorWarning` | `yellow500` | `yellow500` |
| `colorDanger` | `red500` | `red500` |
| `colorDangerEmphasis` | `red600` | `red600` |
| `colorInfo` | `sky400` | `sky400` |
| `colorOverlay` | `overlay` | `overlay` |

### Theming Rules

- The theme engine MUST make the full semantic token set available globally (e.g., via CSS custom properties or a React context / CSS-in-JS theme provider).
- Components MUST reference semantic tokens exclusively — never primitives, never hardcoded hex/rgb values.
- The bill payment dashboard MUST use `colorDanger` / `colorDangerEmphasis` for overdue bills and `colorSuccess` / `colorSuccessEmphasis` for paid bills to provide immediate visual affordance.
- The financial history charts MUST use the primary palette (`colorPrimary`, `colorSuccess`, `colorWarning`, `colorDanger`) consistently across all chart series.
- Interactive elements (buttons, links, inputs) MUST have hover, active, focus, and disabled states each mapped to an appropriate semantic token.

### Typography Scale

All text rendered in the application MUST reference semantic typography tokens from
the centralised design token file. Hardcoded font-size, font-weight, or line-height
values in component or style files are forbidden.

**Primitive font-size scale** (rem-based; browser root = 16px):

| Token | Value | px equiv | Intended use |
|---|---|---|---|
| `fontSizeXs` | `0.75rem` | 12px | Captions, badges, status text |
| `fontSizeSm` | `0.875rem` | 14px | Secondary labels, metadata, hints |
| `fontSizeBase` | `1rem` | 16px | Primary body text, input values |
| `fontSizeLg` | `1.125rem` | 18px | Card content, emphasised body |
| `fontSizeXl` | `1.25rem` | 20px | Section sub-headings |
| `fontSize2xl` | `1.5rem` | 24px | Page sub-headings |
| `fontSize3xl` | `1.875rem` | 30px | Page primary headings |
| `fontSize4xl` | `2.25rem` | 36px | Hero / dashboard display numbers |

**Semantic typography tokens** (components reference these):

| Semantic token | Primitive | Weight | Line-height |
|---|---|---|---|
| `fontSizeCaption` | `fontSizeXs` | `fontWeightRegular` | `lineHeightSnug` |
| `fontSizeBodySmall` | `fontSizeSm` | `fontWeightRegular` | `lineHeightNormal` |
| `fontSizeBody` | `fontSizeBase` | `fontWeightRegular` | `lineHeightNormal` |
| `fontSizeLabel` | `fontSizeSm` | `fontWeightMedium` | `lineHeightTight` |
| `fontSizeHeading4` | `fontSizeLg` | `fontWeightSemibold` | `lineHeightSnug` |
| `fontSizeHeading3` | `fontSizeXl` | `fontWeightSemibold` | `lineHeightSnug` |
| `fontSizeHeading2` | `fontSize2xl` | `fontWeightBold` | `lineHeightTight` |
| `fontSizeHeading1` | `fontSize3xl` | `fontWeightBold` | `lineHeightTight` |
| `fontSizeDisplay` | `fontSize4xl` | `fontWeightBold` | `lineHeightTight` |

Application-specific typography conventions:
- Dashboard summary totals and balance figures MUST use `fontSizeDisplay` or `fontSizeHeading1`.
- Bill amounts in the payment dashboard MUST use `fontSizeHeading3` or larger.
- Due dates paired with status colours MUST use `fontSizeLabel`.
- Chart axis labels and legends MUST use `fontSizeCaption`.
- Form input values and helper text MUST use `fontSizeBody` and `fontSizeBodySmall` respectively.

Font weight tokens: `fontWeightRegular` (400), `fontWeightMedium` (500),
`fontWeightSemibold` (600), `fontWeightBold` (700).

Line-height tokens: `lineHeightTight` (1.25), `lineHeightSnug` (1.375),
`lineHeightNormal` (1.5), `lineHeightRelaxed` (1.625).

### Layout & Responsiveness

All screens and components MUST follow a **mobile-first** approach: base styles
target the smallest viewport (minimum 320px) and breakpoints are applied using
`min-width` media queries exclusively. `max-width` layout breakpoints are forbidden.

**Breakpoints**:

| Token | Min-width | Target |
|---|---|---|
| `breakpointSm` | `480px` | Large phones |
| `breakpointMd` | `768px` | Tablets |
| `breakpointLg` | `1024px` | Laptops / small desktops |
| `breakpointXl` | `1280px` | Desktops |
| `breakpoint2xl` | `1536px` | Large / wide desktops |

Screen-by-screen responsive layout mandates:
- **PDF Upload & Classification**: single-column card at 320px; two-column grid at `breakpointMd`+.
- **Payment Dashboard**: bill cards stacked vertically on mobile; 2-column grid at `breakpointMd`+; 3-column at `breakpointLg`+; QR code minimum `200×200px` at all breakpoints.
- **Financial History Dashboard**: charts full-width single-column on mobile; multi-panel side-by-side at `breakpointLg`+.
- **Document Library**: list view on mobile; table view at `breakpointMd`+.

Universal responsive rules:
- Every screen MUST be functional without horizontal scroll at 320px viewport width.
- Touch targets MUST be at minimum `44×44 CSS px` (WCAG 2.5.5).
- Body font size MUST be at minimum `fontSizeBase` (16px) on all viewports.
- All images and icons MUST use SVG or provide `1×`/`2×`/`3×` raster variants for standard, Retina, and high-DPI displays.
- No container MUST have a fixed width that produces horizontal overflow narrower than `breakpointSm` (480px).
