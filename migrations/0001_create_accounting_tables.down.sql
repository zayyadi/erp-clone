-- Drop Journal Lines Table
DROP TABLE IF EXISTS journal_lines;

-- Drop Journal Entries Table
DROP TABLE IF EXISTS journal_entries;

-- Drop Chart of Accounts Table
DROP TABLE IF EXISTS chart_of_accounts;

-- Drop the trigger function if it's no longer needed by other tables
-- (Be cautious if this function is shared)
DROP FUNCTION IF EXISTS trigger_set_timestamp();

-- Drop the UUID extension if no other table uses it
-- (Be cautious, this is a system-wide extension)
-- DROP EXTENSION IF EXISTS "uuid-ossp";

COMMENT ON TABLE journal_lines IS 'Dropped table journal_lines';
COMMENT ON TABLE journal_entries IS 'Dropped table journal_entries';
COMMENT ON TABLE chart_of_accounts IS 'Dropped table chart_of_accounts';
