// Based on backend models (internal/inventory/models/item.go)
// and DTOs (internal/inventory/service/dto/item_dto.go)

export enum ItemType {
  FinishedGood = "Finished Good",
  RawMaterial = "Raw Material",
  WorkInProgress = "Work In Progress",
  Service = "Service",
  Other = "Other",
}

export enum ValuationMethod {
  FIFO = "FIFO", // First-In, First-Out
  LIFO = "LIFO", // Last-In, First-Out (Note: LIFO is not allowed under IFRS)
  WeightedAverage = "Weighted Average",
  SpecificIdentification = "Specific Identification", // Usually for high-value, unique items
}

export interface Item {
  id: string;
  sku: string; // Stock Keeping Unit
  name: string;
  description?: string;
  item_type: ItemType;
  unit_of_measure: string; // e.g., "pcs", "kg", "ltr"
  purchase_price?: number | null; // Default purchase price
  sales_price?: number | null;   // Default sales price
  valuation_method?: ValuationMethod | null; // How item cost is determined
  is_active: boolean;
  created_at: string;
  updated_at: string;
  // Potential future fields:
  // category_id?: string;
  // brand_id?: string;
  // supplier_id?: string;
  // stock_quantity?: number; // This would typically come from a separate stock level query
}

export interface CreateItemRequest {
  sku: string;
  name: string;
  description?: string;
  item_type: ItemType;
  unit_of_measure: string;
  purchase_price?: number | null;
  sales_price?: number | null;
  valuation_method?: ValuationMethod | null;
  is_active?: boolean; // Defaults to true on backend
}

export interface UpdateItemRequest {
  name?: string;
  description?: string;
  item_type?: ItemType;
  unit_of_measure?: string;
  purchase_price?: number | null;
  sales_price?: number | null;
  valuation_method?: ValuationMethod | null;
  is_active?: boolean;
}

export interface ListItemRequest {
  page?: number;
  limit?: number;
  name?: string;
  sku?: string;
  item_type?: ItemType | string; // string for query param flexibility
  is_active?: boolean | string; // string for query param flexibility
}

export interface PaginatedItemResponse {
  data: Item[];
  page: number;
  limit: number;
  total: number;
}

// For Redux state
export interface ItemState {
  items: Item[];
  selectedItem: Item | null;
  loading: boolean;
  error: string | null | undefined;
  page: number;
  limit: number;
  total: number;
  loadingCreate: boolean;
  loadingUpdate: boolean;
  loadingDelete: boolean;
  loadingFetch: boolean;
}
