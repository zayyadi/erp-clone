-- Create Items Table
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    unit_of_measure VARCHAR(20) NOT NULL,
    item_type VARCHAR(20) NOT NULL, -- RAW_MATERIAL, FINISHED_GOOD, WIP, NON_INVENTORY
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_items_item_type ON items(item_type);
CREATE INDEX IF NOT EXISTS idx_items_is_active ON items(is_active);
CREATE INDEX IF NOT EXISTS idx_items_deleted_at ON items(deleted_at);
COMMENT ON COLUMN items.item_type IS 'Valid types: RAW_MATERIAL, FINISHED_GOOD, WIP, NON_INVENTORY';

-- Create Warehouses Table
CREATE TABLE IF NOT EXISTS warehouses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_warehouses_is_active ON warehouses(is_active);
CREATE INDEX IF NOT EXISTS idx_warehouses_deleted_at ON warehouses(deleted_at);

-- Create Inventory Transactions Table
CREATE TABLE IF NOT EXISTS inventory_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    item_id UUID NOT NULL,
    warehouse_id UUID NOT NULL,
    quantity NUMERIC(10, 3) NOT NULL, -- Quantity of the transaction (always positive)
    transaction_type VARCHAR(30) NOT NULL, -- e.g., RECEIVE_STOCK, ISSUE_STOCK, ADJUST_STOCK_IN, ADJUST_STOCK_OUT
    reference_id UUID, -- Optional: links to PO, SO, Adjustment ID, Transfer ID etc.
    transaction_date TIMESTAMPTZ NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- No soft delete for transactions usually; they are reversed by counter-transactions.

    CONSTRAINT fk_inv_item
        FOREIGN KEY(item_id)
        REFERENCES items(id)
        ON DELETE RESTRICT, -- Prevent deleting an item if it has transactions
    CONSTRAINT fk_inv_warehouse
        FOREIGN KEY(warehouse_id)
        REFERENCES warehouses(id)
        ON DELETE RESTRICT -- Prevent deleting a warehouse if it has transactions
);

CREATE INDEX IF NOT EXISTS idx_inv_transactions_item_id ON inventory_transactions(item_id);
CREATE INDEX IF NOT EXISTS idx_inv_transactions_warehouse_id ON inventory_transactions(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_inv_transactions_transaction_type ON inventory_transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_inv_transactions_transaction_date ON inventory_transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_inv_transactions_reference_id ON inventory_transactions(reference_id);
COMMENT ON COLUMN inventory_transactions.quantity IS 'Quantity of the transaction (always positive). Transaction type indicates inflow/outflow.';
COMMENT ON COLUMN inventory_transactions.transaction_type IS 'e.g., RECEIVE_STOCK, ISSUE_STOCK, ADJUST_STOCK_IN, ADJUST_STOCK_OUT, TRANSFER_OUT, TRANSFER_IN, etc.';


-- Apply timestamp update trigger to new tables
-- (Assuming trigger_set_timestamp() function was created in 0001 migration)
CREATE TRIGGER set_timestamp_items
BEFORE UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_warehouses
BEFORE UPDATE ON warehouses
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_inventory_transactions
BEFORE UPDATE ON inventory_transactions
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();
