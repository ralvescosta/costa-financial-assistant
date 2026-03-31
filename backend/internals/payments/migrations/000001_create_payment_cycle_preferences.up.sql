-- Migration: create payment_cycle_preferences
-- Stores each project's preferred payment day of month for billing cycle calculations.

CREATE TABLE IF NOT EXISTS payment_cycle_preferences (
    id                     UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id             UUID     NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    preferred_day_of_month SMALLINT NOT NULL CHECK (preferred_day_of_month BETWEEN 1 AND 28),
    updated_by             UUID     NOT NULL,
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_cycle_prefs_project ON payment_cycle_preferences (project_id);
