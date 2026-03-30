# Costa Financial Assistant

A multi-tenant financial assistant focused on helping users organize bills and account statements from PDF documents, track payments, and build financial history over time.

This project is being built end-to-end using a Spec-Driven Development workflow with GitHub Spec Kit, with GitHub Copilot acting as the coding agent.

## Project Vision

The application is designed to support a practical monthly financial workflow:

- Upload and classify PDFs (bill or account statement)
- Process documents asynchronously to extract structured data
- Show payment-ready bill details (Pix QR payload and barcode)
- Track bill payment status and overdue items
- Reconcile statement transactions against known bills
- Provide historical dashboards for spending and payment compliance

The domain is optimized for Brazilian financial context (Pix and boleto/barcode workflows).

## Development Approach

This repository follows Spec-Driven Development as the primary delivery method.

Feature development starts in the specs folder and moves to code only after requirements, plan, data model, contracts, and tasks are defined.

## Architecture (Planned)

The platform uses a modular backend plus web frontend:

- Backend: Go microservices + BFF pattern
- BFF: Echo + Huma (OpenAPI-first), MVC separation
- Data and infra: PostgreSQL, Redis, RabbitMQ, S3-compatible object storage
- Observability: OpenTelemetry
- Frontend: React + Vite + Tailwind CSS (mobile-first, tokenized theming)

Current backend module structure is under backend/cmd and backend/internals:

- bff
- files
- bills
- payments
- identity
- onboarding
- migrations

## License

This project is licensed under the MIT License.
See [LICENSE](LICENSE) for details.