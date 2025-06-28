-- Drop Inventory Transactions Table
DROP TABLE IF EXISTS inventory_transactions;

-- Drop Warehouses Table
DROP TABLE IF EXISTS warehouses;

-- Drop Items Table
DROP TABLE IF EXISTS items;

-- Triggers are dropped automatically when tables are dropped if they were table-specific.
-- If the trigger function `trigger_set_timestamp()` was only for these tables
-- and not shared with tables from migration 0001 (which it is),
-- then you would not drop the function here.
-- It's generally better to manage trigger functions separately if they are truly shared
-- or ensure they are dropped only when the last table using them is dropped.
-- For now, we assume the function `trigger_set_timestamp` is still needed by accounting tables.

COMMENT ON TABLE inventory_transactions IS 'Dropped table inventory_transactions';
COMMENT ON TABLE warehouses IS 'Dropped table warehouses';
COMMENT ON TABLE items IS 'Dropped table items';
