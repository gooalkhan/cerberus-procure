package main

import (
	 "cerberus-procure/internal/logic"
	 "cerberus-procure/internal/models"
	 "cerberus-procure/internal/repository/sqlite"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

var apiLogger *log.Logger
var serverLogger *log.Logger

func initLogger() {
	if err := os.MkdirAll("log", 0755); err != nil {
		fmt.Println("Error creating log directory:", err)
		return
	}
	file1, err := os.OpenFile("log/api.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		apiLogger = log.New(file1, "API ", log.Ldate|log.Ltime)
	}
	file2, err := os.OpenFile("log/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		serverLogger = log.New(file2, "SERVER ", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

//go:embed dist/*
var frontendAssets embed.FS

var todoUC *logic.TodoUseCase
var authUC *logic.AuthUseCase
var procureUC *logic.ProcurementUseCase

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := authUC.Login(input.Username, input.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getTodosHandler(w http.ResponseWriter, r *http.Request) {
	todos, err := todoUC.GetTodos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func addTodoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	todo, err := todoUC.AddTodo(input.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func toggleTodoHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := todoUC.ToggleTodo(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func deleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := todoUC.DeleteTodo(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.statusCode == 0 {
		lrw.statusCode = http.StatusOK
	}
	if lrw.statusCode >= 400 && len(lrw.body) < 200 {
		lrw.body = append(lrw.body, b...)
	}
	return lrw.ResponseWriter.Write(b)
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		lrw.Header().Set("Access-Control-Allow-Origin", "*")
		lrw.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		lrw.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == http.MethodOptions {
			lrw.WriteHeader(http.StatusOK)
		} else {
			next(lrw, r)
		}
		
		if apiLogger != nil {
			if lrw.statusCode >= 400 {
				// Clean up the error message for single-line logging
				errMsg := ""
				if len(lrw.body) > 0 {
					errMsg = string(lrw.body)
					if len(errMsg) > 0 && errMsg[len(errMsg)-1] == '\n' {
						errMsg = errMsg[:len(errMsg)-1]
					}
				}
				apiLogger.Printf("[%s] %s %s - %d %s - %v - Error: %s", r.RemoteAddr, r.Method, r.URL.Path, lrw.statusCode, http.StatusText(lrw.statusCode), time.Since(start), errMsg)
			} else {
				apiLogger.Printf("[%s] %s %s - %d %s - %v", r.RemoteAddr, r.Method, r.URL.Path, lrw.statusCode, http.StatusText(lrw.statusCode), time.Since(start))
			}
		}
	}
}

func main() {
	initLogger()
	todoRepo, err := sqlite.NewSQLiteTodoRepository("todos.db")
	if err != nil {
		panic(err)
	}
	userRepo, err := sqlite.NewSQLiteUserRepository(todoRepo.DB())
	if err != nil {
		panic(err)
	}
	procureRepo, err := sqlite.NewSQLiteProcurementRepository(todoRepo.DB())
	if err != nil {
		panic(err)
	}

	todoUC = logic.NewTodoUseCase(todoRepo)
	authUC = logic.NewAuthUseCase(userRepo)
	procureUC = logic.NewProcurementUseCase(procureRepo)

	// Seed admin user if not exists
	authUC.Register("admin", "1234", "Administrator")

	mux := http.NewServeMux()

	// API 핸들러
	mux.HandleFunc("/api/login", corsMiddleware(loginHandler))
	
	mux.HandleFunc("/api/seed", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			procureUC.SeedData()
			w.WriteHeader(http.StatusOK)
		}
	}))
	// Items API
	mux.HandleFunc("/api/items", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			items, err := procureUC.GetItems()
			if err != nil {
				fmt.Println("GetItems Error:", err)
			}
			json.NewEncoder(w).Encode(items)
		} else if r.Method == http.MethodPost {
			var i models.ItemMaster
			err := json.NewDecoder(r.Body).Decode(&i)
			if err != nil {
				fmt.Println("Decode Error (Items):", err)
			}
			err = procureUC.SaveItem(&i)
			if err != nil {
				fmt.Println("Save Error (Items):", err)
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	// Vendors API
	mux.HandleFunc("/api/vendors", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, _ := procureUC.GetVendors()
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.VendorMaster
			json.NewDecoder(r.Body).Decode(&i)
			procureUC.SaveVendor(&i)
			w.WriteHeader(http.StatusOK)
		}
	}))

	// PO API
	mux.HandleFunc("/api/pos", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, _ := procureUC.GetPurchaseOrders()
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.PurchaseOrder
			json.NewDecoder(r.Body).Decode(&i)
			procureUC.SavePurchaseOrder(&i)
			w.WriteHeader(http.StatusOK)
		}
	}))

	// PO Items API
	mux.HandleFunc("/api/pos/items", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			poIDStr := r.URL.Query().Get("poId")
			var poID int
			fmt.Sscanf(poIDStr, "%d", &poID)
			list, _ := procureUC.GetPOItemsByPOID(poID)
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.POItem
			json.NewDecoder(r.Body).Decode(&i)
			procureUC.SavePOItem(&i)
			w.WriteHeader(http.StatusOK)
		}
	}))

	// Invoices API
	mux.HandleFunc("/api/invoices", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, _ := procureUC.GetCommercialInvoices()
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.CommercialInvoice
			json.NewDecoder(r.Body).Decode(&i)
			procureUC.SaveCommercialInvoice(&i)
			w.WriteHeader(http.StatusOK)
		}
	}))

	mux.HandleFunc("/api/invoices/items", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			ciIDStr := r.URL.Query().Get("ciId")
			var ciID int
			fmt.Sscanf(ciIDStr, "%d", &ciID)
			list, _ := procureUC.GetCIAggregatedItems(ciID)
			json.NewEncoder(w).Encode(list)
		}
	}))

	// AP API
	mux.HandleFunc("/api/aps", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, _ := procureUC.GetAccountPayables()
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.AccountPayable
			if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
				if serverLogger != nil { serverLogger.Printf("AP Decode Error: %v", err) }
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := procureUC.SaveAccountPayable(&i); err != nil {
				if serverLogger != nil { serverLogger.Printf("AP Save Error: %v", err) }
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	// Containers API
	mux.HandleFunc("/api/containers", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, _ := procureUC.GetContainers()
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.Container
			json.NewDecoder(r.Body).Decode(&i)
			procureUC.SaveContainer(&i)
			w.WriteHeader(http.StatusOK)
		}
	}))

	mux.HandleFunc("/api/containers/items", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			cIDStr := r.URL.Query().Get("containerId")
			var cID int
			fmt.Sscanf(cIDStr, "%d", &cID)
			list, _ := procureUC.GetContainerItemsByContainerID(cID)
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.ContainerItem
			json.NewDecoder(r.Body).Decode(&i)
			procureUC.SaveContainerItem(&i)
			w.WriteHeader(http.StatusOK)
		}
	}))

	mux.HandleFunc("/api/containers/bl", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			blIDStr := r.URL.Query().Get("blId")
			var blID int
			fmt.Sscanf(blIDStr, "%d", &blID)
			list, _ := procureUC.GetContainersByBLID(blID)
			json.NewEncoder(w).Encode(list)
		}
	}))

	// BL API
	mux.HandleFunc("/api/bls", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, _ := procureUC.GetBLs()
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.BL
			if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
				if serverLogger != nil { serverLogger.Printf("BL Decode Error: %v", err) }
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := procureUC.SaveBL(&i); err != nil {
				if serverLogger != nil { serverLogger.Printf("BL Save Error: %v", err) }
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	// GR API
	mux.HandleFunc("/api/grs", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, _ := procureUC.GetGoodsReceipts()
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.GoodsReceipt
			if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
				if serverLogger != nil { serverLogger.Printf("GR Decode Error: %v", err) }
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := procureUC.SaveGoodsReceipt(&i); err != nil {
				if serverLogger != nil { serverLogger.Printf("GR Save Error: %v", err) }
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	// Lots API
	mux.HandleFunc("/api/lots", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, _ := procureUC.GetInventoryLots()
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.InventoryLot
			if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
				if serverLogger != nil { serverLogger.Printf("Lot Decode Error: %v", err) }
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := procureUC.SaveInventoryLot(&i); err != nil {
				if serverLogger != nil { serverLogger.Printf("Lot Save Error: %v", err) }
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	mux.HandleFunc("/api/lots/gr", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			grIDStr := r.URL.Query().Get("grId")
			var grID int
			fmt.Sscanf(grIDStr, "%d", &grID)
			list, _ := procureUC.GetInventoryLotsByGRID(grID)
			json.NewEncoder(w).Encode(list)
		}
	}))

	// Allocations API
	mux.HandleFunc("/api/allocations", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, _ := procureUC.GetCostAllocations()
			json.NewEncoder(w).Encode(list)
		} else if r.Method == http.MethodPost {
			var i models.CostAllocation
			if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
				if serverLogger != nil { serverLogger.Printf("Cost Allocation Decode Error: %v", err) }
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := procureUC.SaveCostAllocation(&i); err != nil {
				if serverLogger != nil { serverLogger.Printf("Cost Allocation Save Error: %v", err) }
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	// Bookings API
	mux.HandleFunc("/api/bookings", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			list, err := procureUC.GetBookings()
			if err != nil {
				fmt.Println("GetBookings Error:", err)
			}
			json.NewEncoder(w).Encode(list)
		}
	}))

	mux.HandleFunc("/api/todos", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			getTodosHandler(w, r)
		case http.MethodPost:
			addTodoHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/todos/toggle", corsMiddleware(toggleTodoHandler))
	mux.HandleFunc("/api/todos/delete", corsMiddleware(deleteTodoHandler))

	// 프론트엔드 정적 파일 서빙
	distFS, _ := fs.Sub(frontendAssets, "dist")
	mux.Handle("/", http.FileServer(http.FS(distFS)))

	fmt.Println("Server starting on :8080...")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Server Error:", err)
	}
}
