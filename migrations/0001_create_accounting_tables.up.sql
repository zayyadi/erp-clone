-- Create Chart of Accounts Table
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS chart_of_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_code VARCHAR(20) NOT NULL UNIQUE,
    account_name VARCHAR(100) NOT NULL,
    account_type VARCHAR(50) NOT NULL, -- e.g., ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
    parent_account_id UUID,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_parent_account
        FOREIGN KEY(parent_account_id)
        REFERENCES chart_of_accounts(id)
        ON DELETE SET NULL -- Or RESTRICT, depending on desired behavior
);

CREATE INDEX IF NOT EXISTS idx_coa_parent_account_id ON chart_of_accounts(parent_account_id);
CREATE INDEX IF NOT EXISTS idx_coa_account_type ON chart_of_accounts(account_type);
CREATE INDEX IF NOT EXISTS idx_coa_deleted_at ON chart_of_accounts(deleted_at);


-- Create Journal Entries Table
CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entry_date TIMESTAMPTZ NOT NULL,
    description VARCHAR(255),
    reference VARCHAR(100), -- e.g., Invoice number, PO number
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT', -- e.g., DRAFT, POSTED, VOIDED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_je_entry_date ON journal_entries(entry_date);
CREATE INDEX IF NOT EXISTS idx_je_status ON journal_entries(status);
CREATE INDEX IF NOT EXISTS idx_je_deleted_at ON journal_entries(deleted_at);


-- Create Journal Lines Table
CREATE TABLE IF NOT EXISTS journal_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    journal_id UUID NOT NULL,
    account_id UUID NOT NULL,
    amount NUMERIC(15, 2) NOT NULL, -- Amount stored as positive, is_debit determines effect
    currency VARCHAR(3) DEFAULT 'USD',
    is_debit BOOLEAN NOT NULL, -- TRUE for Debit, FALSE for Credit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- No soft delete for journal lines usually, they are deleted if entry is deleted (hard or soft)
    CONSTRAINT fk_journal_entry
        FOREIGN KEY(journal_id)
        REFERENCES journal_entries(id)
        ON DELETE CASCADE, -- If journal entry is deleted, its lines are also deleted
    CONSTRAINT fk_chart_of_account
        FOREIGN KEY(account_id)
        REFERENCES chart_of_accounts(id)
        ON DELETE RESTRICT -- Prevent deleting an account if it's used in journal lines
                           -- This might be too restrictive for soft deletes of accounts.
                           -- Consider application-level checks or SET NULL if appropriate.
);

CREATE INDEX IF NOT EXISTS idx_jl_journal_id ON journal_lines(journal_id);
CREATE INDEX IF NOT EXISTS idx_jl_account_id ON journal_lines(account_id);

-- Trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to chart_of_accounts
CREATE TRIGGER set_timestamp_chart_of_accounts
BEFORE UPDATE ON chart_of_accounts
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- Apply trigger to journal_entries
CREATE TRIGGER set_timestamp_journal_entries
BEFORE UPDATE ON journal_entries
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- Apply trigger to journal_lines
CREATE TRIGGER set_timestamp_journal_lines
BEFORE UPDATE ON journal_lines
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

COMMENT ON COLUMN chart_of_accounts.account_type IS 'Valid types: ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE';
COMMENT ON COLUMN journal_entries.status IS 'Valid statuses: DRAFT, POSTED, VOIDED';
COMMENT ON COLUMN journal_lines.amount IS 'Amount stored as positive value; is_debit flag determines debit/credit nature.';

-- Note on ON DELETE RESTRICT for journal_lines.account_id:
-- If an account is soft-deleted (deleted_at is set), it can still be referenced by journal_lines.
-- If an account is hard-deleted, this constraint would prevent it if lines exist.
-- For a system with soft deletes, application logic should prevent posting to/using soft-deleted accounts.
-- The ON DELETE RESTRICT here primarily protects against hard deletion of an account that has history.
-- If an account needs to be "removed" but has history, it should be deactivated (is_active = false)
-- and potentially merged or archived through application processes.
-- For ON DELETE SET NULL on chart_of_accounts.parent_account_id: if a parent is deleted, children become top-level.

-- Seed some basic accounts (optional, can be done via application logic/UI)
-- Example:
-- INSERT INTO chart_of_accounts (account_code, account_name, account_type, is_active) VALUES
-- ('1000', 'Cash', 'ASSET', TRUE),
-- ('1100', 'Accounts Receivable', 'ASSET', TRUE),
-- ('2000', 'Accounts Payable', 'LIABILITY', TRUE),
-- ('3000', 'Common Stock', 'EQUITY', TRUE),
-- ('4000', 'Sales Revenue', 'REVENUE', TRUE),
-- ('5000', 'Office Supplies Expense', 'EXPENSE', TRUE);

-- Make sure to create a corresponding .down.sql file to revert these changes.
