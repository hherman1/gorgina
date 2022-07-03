-- Add support for the hidden column
ALTER TABLE catalog ADD COLUMN IF NOT EXISTS hidden boolean NOT NULL DEFAULT false;