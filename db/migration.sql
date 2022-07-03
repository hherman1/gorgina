-- Add support for the hidden column
ALTER TABLE catalog ADD COLUMN IF NOT EXISTS hidden boolean NOT NULL DEFAULT false;

-- Add support for use notes
ALTER TABLE catalog ADD COLUMN IF NOT EXISTS last_note text;
ALTER TABLE activity ADD COLUMN IF NOT EXISTS note text;