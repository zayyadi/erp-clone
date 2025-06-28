package handlers

import (
	"encoding/json"
	"erp-system/internal/inventory/models"
	"erp-system/internal/inventory/service"
	inv_dto "erp-system/internal/inventory/service/dto" // Alias for inventory specific DTOs
	"erp-system/pkg/errors" // Keep this for type assertion if needed
	"erp-system/pkg/logger" // Keep this
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// InventoryHandlers wraps the inventory service to provide HTTP handlers.
type InventoryHandlers struct {
	service service.InventoryService
}

// NewInventoryHandlers creates a new InventoryHandlers instance.
func NewInventoryHandlers(serv service.InventoryService) *InventoryHandlers {
	return &InventoryHandlers{service: serv}
}

// RegisterInventoryRoutes registers all inventory module routes with the provided router.
func (h *InventoryHandlers) RegisterInventoryRoutes(r *mux.Router) {
	// Item Routes
	itemRouter := r.PathPrefix("/api/v1/inventory/items").Subrouter()
	itemRouter.HandleFunc("", h.CreateItem).Methods("POST")
	itemRouter.HandleFunc("", h.ListItems).Methods("GET")
	itemRouter.HandleFunc("/{id}", h.GetItemByID).Methods("GET")
	itemRouter.HandleFunc("/sku/{sku}", h.GetItemBySKU).Methods("GET") // Custom route for by SKU
	itemRouter.HandleFunc("/{id}", h.UpdateItem).Methods("PUT")
	itemRouter.HandleFunc("/{id}", h.DeleteItem).Methods("DELETE")

	// Warehouse Routes
	warehouseRouter := r.PathPrefix("/api/v1/inventory/warehouses").Subrouter()
	warehouseRouter.HandleFunc("", h.CreateWarehouse).Methods("POST")
	warehouseRouter.HandleFunc("", h.ListWarehouses).Methods("GET")
	warehouseRouter.HandleFunc("/{id}", h.GetWarehouseByID).Methods("GET")
	warehouseRouter.HandleFunc("/code/{code}", h.GetWarehouseByCode).Methods("GET")
	warehouseRouter.HandleFunc("/{id}", h.UpdateWarehouse).Methods("PUT")
	warehouseRouter.HandleFunc("/{id}", h.DeleteWarehouse).Methods("DELETE")

	// Inventory Transaction/Adjustment Routes
	adjustmentRouter := r.PathPrefix("/api/v1/inventory/adjustments").Subrouter()
	adjustmentRouter.HandleFunc("", h.CreateInventoryAdjustment).Methods("POST")

	// Inventory Level Routes
	levelRouter := r.PathPrefix("/api/v1/inventory/levels").Subrouter()
	levelRouter.HandleFunc("", h.GetInventoryLevels).Methods("GET")
    levelRouter.HandleFunc("/item/{itemId}/warehouse/{warehouseId}", h.GetSpecificItemStockLevel).Methods("GET")
}


// --- Item Handlers ---

func (h *InventoryHandlers) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req inv_dto.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, errors.NewValidationError("Invalid request payload", err.Error()))
		return
	}
	defer r.Body.Close()

	item, err := h.service.CreateItem(r.Context(), req)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, item)
}

func (h *InventoryHandlers) GetItemByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, errors.NewValidationError("Missing item ID in path", "id"))
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid item ID format", "id"))
		return
	}

	item, err := h.service.GetItemByID(r.Context(), id)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, item)
}

func (h *InventoryHandlers) GetItemBySKU(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sku, ok := vars["sku"]
	if !ok || sku == "" {
		respondWithError(w, errors.NewValidationError("Missing item SKU in path", "sku"))
		return
	}

	item, err := h.service.GetItemBySKU(r.Context(), sku)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, item)
}

func (h *InventoryHandlers) UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, errors.NewValidationError("Missing item ID in path", "id"))
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid item ID format", "id"))
		return
	}

	var req inv_dto.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, errors.NewValidationError("Invalid request payload", err.Error()))
		return
	}
	defer r.Body.Close()

	item, err := h.service.UpdateItem(r.Context(), id, req)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, item)
}

func (h *InventoryHandlers) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, errors.NewValidationError("Missing item ID in path", "id"))
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, errors.NewValidationError("Invalid item ID format", "id"))
		return
	}

	if err := h.service.DeleteItem(r.Context(), id); err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Item deleted successfully"})
}

func (h *InventoryHandlers) ListItems(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	listReq := inv_dto.ListItemRequest{
		Page:     1,
		Limit:    20,
		Name:     queryParams.Get("name"),
		SKU:      queryParams.Get("sku"),
		ItemType: models.ItemType(queryParams.Get("item_type")),
	}

	if pageStr := queryParams.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 { listReq.Page = page }
	}
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 { listReq.Limit = limit }
	}
	if isActiveStr := queryParams.Get("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil { listReq.IsActive = &isActive } else {
			respondWithError(w, errors.NewValidationError("Invalid boolean value for 'is_active'", "is_active"))
			return
		}
	}

	items, total, err := h.service.ListItems(r.Context(), listReq)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, PaginatedResponse{Data: items, Page: listReq.Page, Limit: listReq.Limit, Total: total})
}


// --- Warehouse Handlers ---

func (h *InventoryHandlers) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var req inv_dto.CreateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, errors.NewValidationError("Invalid request payload", err.Error())); return
	}
	defer r.Body.Close()
	warehouse, err := h.service.CreateWarehouse(r.Context(), req)
	if err != nil { respondWithError(w, err); return }
	respondWithJSON(w, http.StatusCreated, warehouse)
}

func (h *InventoryHandlers) GetWarehouseByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r); idStr, _ := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil { respondWithError(w, errors.NewValidationError("Invalid warehouse ID", "id")); return }
	warehouse, err := h.service.GetWarehouseByID(r.Context(), id)
	if err != nil { respondWithError(w, err); return }
	respondWithJSON(w, http.StatusOK, warehouse)
}

func (h *InventoryHandlers) GetWarehouseByCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r); code, _ := vars["code"]
	if code == "" { respondWithError(w, errors.NewValidationError("Missing warehouse code", "code")); return }
	warehouse, err := h.service.GetWarehouseByCode(r.Context(), code)
	if err != nil { respondWithError(w, err); return }
	respondWithJSON(w, http.StatusOK, warehouse)
}

func (h *InventoryHandlers) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r); idStr, _ := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil { respondWithError(w, errors.NewValidationError("Invalid warehouse ID", "id")); return }
	var req inv_dto.UpdateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, errors.NewValidationError("Invalid request payload", err.Error())); return
	}
	defer r.Body.Close()
	warehouse, err := h.service.UpdateWarehouse(r.Context(), id, req)
	if err != nil { respondWithError(w, err); return }
	respondWithJSON(w, http.StatusOK, warehouse)
}

func (h *InventoryHandlers) DeleteWarehouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r); idStr, _ := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil { respondWithError(w, errors.NewValidationError("Invalid warehouse ID", "id")); return }
	if err := h.service.DeleteWarehouse(r.Context(), id); err != nil {
		respondWithError(w, err); return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Warehouse deleted successfully"})
}

func (h *InventoryHandlers) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	listReq := inv_dto.ListWarehouseRequest{ Page: 1, Limit: 20 }
	if pageStr := queryParams.Get("page"); pageStr != "" { if page, err := strconv.Atoi(pageStr); err == nil && page > 0 { listReq.Page = page } }
	if limitStr := queryParams.Get("limit"); limitStr != "" { if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 { listReq.Limit = limit } }
	listReq.Name = queryParams.Get("name")
	listReq.Code = queryParams.Get("code")
	listReq.Location = queryParams.Get("location")
	if isActiveStr := queryParams.Get("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil { listReq.IsActive = &isActive } else {
			respondWithError(w, errors.NewValidationError("Invalid boolean value for 'is_active'", "is_active")); return
		}
	}
	warehouses, total, err := h.service.ListWarehouses(r.Context(), listReq)
	if err != nil { respondWithError(w, err); return }
	respondWithJSON(w, http.StatusOK, PaginatedResponse{Data: warehouses, Page: listReq.Page, Limit: listReq.Limit, Total: total})
}


// --- Inventory Adjustment Handlers ---

func (h *InventoryHandlers) CreateInventoryAdjustment(w http.ResponseWriter, r *http.Request) {
	var req inv_dto.CreateInventoryAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, errors.NewValidationError("Invalid request payload", err.Error())); return
	}
	defer r.Body.Close()

	transaction, err := h.service.CreateInventoryAdjustment(r.Context(), req)
	if err != nil { respondWithError(w, err); return }
	respondWithJSON(w, http.StatusCreated, transaction)
}


// --- Inventory Level Handlers ---

func (h *InventoryHandlers) GetInventoryLevels(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	req := inv_dto.InventoryLevelRequest{}

	if itemIDStr := queryParams.Get("item_id"); itemIDStr != "" {
		id, err := uuid.Parse(itemIDStr)
		if err != nil { respondWithError(w, errors.NewValidationError("Invalid item_id format", "item_id")); return }
		req.ItemID = &id
	}
	if warehouseIDStr := queryParams.Get("warehouse_id"); warehouseIDStr != "" {
		id, err := uuid.Parse(warehouseIDStr)
		if err != nil { respondWithError(w, errors.NewValidationError("Invalid warehouse_id format", "warehouse_id")); return }
		req.WarehouseID = &id
	}
	if asOfDateStr := queryParams.Get("as_of_date"); asOfDateStr != "" {
		t, err := time.Parse("2006-01-02", asOfDateStr)
		if err != nil { respondWithError(w, errors.NewValidationError("Invalid as_of_date format, use YYYY-MM-DD", "as_of_date")); return }
		req.AsOfDate = &t
	}

	levels, err := h.service.GetInventoryLevels(r.Context(), req)
	if err != nil { respondWithError(w, err); return }
	respondWithJSON(w, http.StatusOK, levels)
}

func (h *InventoryHandlers) GetSpecificItemStockLevel(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    itemIDStr, okItem := vars["itemId"]
    warehouseIDStr, okWarehouse := vars["warehouseId"]

    if !okItem || !okWarehouse {
        respondWithError(w, errors.NewValidationError("Missing item ID or warehouse ID in path", ""))
        return
    }

    itemID, err := uuid.Parse(itemIDStr)
    if err != nil { respondWithError(w, errors.NewValidationError("Invalid item ID format", "itemId")); return }

    warehouseID, err := uuid.Parse(warehouseIDStr)
    if err != nil { respondWithError(w, errors.NewValidationError("Invalid warehouse ID format", "warehouseId")); return }

    asOfDateStr := r.URL.Query().Get("as_of_date")
    asOfDate := time.Now()
    if asOfDateStr != "" {
        parsedDate, err := time.Parse("2006-01-02", asOfDateStr)
        if err != nil { respondWithError(w, errors.NewValidationError("Invalid as_of_date format, use YYYY-MM-DD", "as_of_date")); return }
        asOfDate = parsedDate
    }

    quantity, err := h.service.GetItemStockLevelInWarehouse(r.Context(), itemID, warehouseID, asOfDate)
    if err != nil {
        respondWithError(w, err)
        return
    }

    item, itemErr := h.service.GetItemByID(r.Context(), itemID)
    warehouse, whErr := h.service.GetWarehouseByID(r.Context(), warehouseID)

    if itemErr != nil || whErr != nil {
        logger.ErrorLogger.Printf("Error fetching item/warehouse details for stock level response: itemErr=%v, whErr=%v", itemErr, whErr)
         respondWithJSON(w, http.StatusOK, map[string]interface{}{"item_id": itemID, "warehouse_id": warehouseID, "quantity": quantity, "as_of_date": asOfDate})
        return
    }

    responsePayload := inv_dto.ItemStockLevelInfo{
        ItemID: item.ID, ItemSKU: item.SKU, ItemName: item.Name,
        WarehouseID: warehouse.ID, WarehouseCode: warehouse.Code, WarehouseName: warehouse.Name,
        Quantity: quantity, AsOfDate: asOfDate,
    }
    respondWithJSON(w, http.StatusOK, responsePayload)
}
// Ensure there are no characters or comments after this final closing brace.
