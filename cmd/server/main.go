package main

import (
	 "cerberus-procure/internal/logic"
	 "cerberus-procure/internal/models"
	 "cerberus-procure/internal/repository/sqlite"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sync"
	"time"
)

type session struct {
	userID  int
	expires time.Time
}

var sessions = make(map[string]*session)
var sessionsMu sync.RWMutex

const sessionCookieName = "session_id"
const sessionDuration = 24 * time.Hour

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func cleanupExpiredSessions() {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	now := time.Now()
	for id, s := range sessions {
		if now.After(s.expires) {
			delete(sessions, id)
		}
	}
}

func sessionCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
}

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

	// Create session
	sessionId, err := generateSessionID()
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	sessionsMu.Lock()
	sessions[sessionId] = &session{userID: user.ID, expires: time.Now().Add(sessionDuration)}
	sessionsMu.Unlock()

	http.SetCookie(w, sessionCookie(sessionId, int(sessionDuration.Seconds())))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getSession(r *http.Request) *session {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil
	}
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	s, ok := sessions[cookie.Value]
	if !ok || time.Now().After(s.expires) {
		return nil
	}
	return s
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	s := getSession(r)
	if s == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := authUC.GetUserByID(s.userID)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		sessionsMu.Lock()
		delete(sessions, cookie.Value)
		sessionsMu.Unlock()
	}

	http.SetCookie(w, sessionCookie("", -1))
	w.WriteHeader(http.StatusOK)
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
	statusCode  int
	body        []byte
	handlerName string
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

func getFunctionName(i interface{}) string {
	if i == nil {
		return "unknown"
	}
	pc := reflect.ValueOf(i).Pointer()
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	defaultHandlerName := getFunctionName(next)
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			handlerName:    defaultHandlerName,
		}
		
		lrw.Header().Set("Access-Control-Allow-Origin", "*")
		lrw.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		lrw.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == http.MethodOptions {
			lrw.WriteHeader(http.StatusOK)
		} else {
			next(lrw, r)
		}
		
		if apiLogger != nil {
			handlerName := lrw.handlerName
			if handlerName == "" {
				handlerName = "unknown"
			}
			if lrw.statusCode >= 400 {
				// Clean up the error message for single-line logging
				errMsg := ""
				if len(lrw.body) > 0 {
					errMsg = string(lrw.body)
					if len(errMsg) > 0 && errMsg[len(errMsg)-1] == '\n' {
						errMsg = errMsg[:len(errMsg)-1]
					}
				}
				apiLogger.Printf("[%s] %s %s [%s] - %d %s - %v - Error: %s", r.RemoteAddr, r.Method, r.URL.Path, handlerName, lrw.statusCode, http.StatusText(lrw.statusCode), time.Since(start), errMsg)
			} else {
				apiLogger.Printf("[%s] %s %s [%s] - %d %s - %v", r.RemoteAddr, r.Method, r.URL.Path, handlerName, lrw.statusCode, http.StatusText(lrw.statusCode), time.Since(start))
			}
		}
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	targetHandlerName := getFunctionName(next)
	return func(w http.ResponseWriter, r *http.Request) {
		if lrw, ok := w.(*loggingResponseWriter); ok {
			lrw.handlerName = targetHandlerName
		}

		if getSession(r) == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func seedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		procureUC.SeedData()
		w.WriteHeader(http.StatusOK)
	}
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
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
}

func vendorsHandler(w http.ResponseWriter, r *http.Request) {
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
}

func posHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		list, _ := procureUC.GetPurchaseOrders()
		json.NewEncoder(w).Encode(list)
	} else if r.Method == http.MethodPost {
		var i models.PurchaseOrder
		if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
			if serverLogger != nil { serverLogger.Printf("PO Decode Error: %v", err) }
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := procureUC.SavePurchaseOrder(&i); err != nil {
			if serverLogger != nil { serverLogger.Printf("PO Save Error: %v", err) }
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func poItemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		poIDStr := r.URL.Query().Get("poId")
		var poID int
		fmt.Sscanf(poIDStr, "%d", &poID)
		list, _ := procureUC.GetPOItemsByPOID(poID)
		json.NewEncoder(w).Encode(list)
	} else if r.Method == http.MethodPost {
		var i models.POItem
		if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := procureUC.SavePOItem(&i); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else if r.Method == http.MethodDelete {
		idStr := r.URL.Query().Get("id")
		var id int
		fmt.Sscanf(idStr, "%d", &id)
		if err := procureUC.DeletePOItem(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func invoicesHandler(w http.ResponseWriter, r *http.Request) {
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
}

func invoiceItemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		ciIDStr := r.URL.Query().Get("ciId")
		var ciID int
		fmt.Sscanf(ciIDStr, "%d", &ciID)
		list, _ := procureUC.GetCIAggregatedItems(ciID)
		json.NewEncoder(w).Encode(list)
	}
}

func apsHandler(w http.ResponseWriter, r *http.Request) {
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
}

func containersHandler(w http.ResponseWriter, r *http.Request) {
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
}

func containerItemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		cIDStr := r.URL.Query().Get("containerId")
		var cID int
		fmt.Sscanf(cIDStr, "%d", &cID)
		list, _ := procureUC.GetContainerItemsByContainerID(cID)
		json.NewEncoder(w).Encode(list)
	} else if r.Method == http.MethodPost {
		var i models.ContainerItem
		if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
			if serverLogger != nil { serverLogger.Printf("ContainerItem Decode Error: %v", err) }
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := procureUC.SaveContainerItem(&i); err != nil {
			if serverLogger != nil { serverLogger.Printf("ContainerItem Save Error: %v", err) }
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func containersByBLHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		blIDStr := r.URL.Query().Get("blId")
		var blID int
		fmt.Sscanf(blIDStr, "%d", &blID)
		list, _ := procureUC.GetContainersByBLID(blID)
		json.NewEncoder(w).Encode(list)
	}
}

func blsHandler(w http.ResponseWriter, r *http.Request) {
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
}

func grsHandler(w http.ResponseWriter, r *http.Request) {
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
}

func lotsHandler(w http.ResponseWriter, r *http.Request) {
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
}

func lotsByGRHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		grIDStr := r.URL.Query().Get("grId")
		var grID int
		fmt.Sscanf(grIDStr, "%d", &grID)
		list, _ := procureUC.GetInventoryLotsByGRID(grID)
		json.NewEncoder(w).Encode(list)
	}
}

func allocationsHandler(w http.ResponseWriter, r *http.Request) {
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
}

func bookingsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		list, err := procureUC.GetBookings()
		if err != nil {
			fmt.Println("GetBookings Error:", err)
		}
		json.NewEncoder(w).Encode(list)
	}
}

func bookingTemplateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		data, err := procureUC.GetBookingTemplateData()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	}
}

func bookingBulkImportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		var rows []models.BulkImportRow
		if err := json.NewDecoder(r.Body).Decode(&rows); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		created, err := procureUC.BulkCreateContainerItems(rows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"created": created,
			"total":   len(rows),
		})
	}
}

func checkReferencesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		tableName := r.URL.Query().Get("table")
		idStr := r.URL.Query().Get("id")
		var id int
		fmt.Sscanf(idStr, "%d", &id)

		result, err := procureUC.CheckReferences(tableName, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(result)
	}
}

func deleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodDelete {
		tableName := r.URL.Query().Get("table")
		idStr := r.URL.Query().Get("id")
		var id int
		fmt.Sscanf(idStr, "%d", &id)

		if err := procureUC.DeleteRecord(tableName, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func apTargetGroupsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		list, err := procureUC.GetAPTargetGroups()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(list)
	} else if r.Method == http.MethodPost {
		var g models.AP_TargetGroup
		if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := procureUC.SaveAPTargetGroup(&g); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(g)
	}
}

func apTargetGroupItemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		groupIDStr := r.URL.Query().Get("groupId")
		var groupID int
		fmt.Sscanf(groupIDStr, "%d", &groupID)
		list, err := procureUC.GetAPTargetGroupByID(groupID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(list)
	} else if r.Method == http.MethodPost {
		var item models.AP_TargetGroupItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := procureUC.SaveAPTargetGroupItem(&item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(item)
	} else if r.Method == http.MethodDelete {
		idStr := r.URL.Query().Get("id")
		var id int
		fmt.Sscanf(idStr, "%d", &id)
		if err := procureUC.DeleteAPTargetGroupItem(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func apTargetGroupReferencesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		list, err := procureUC.GetAllReferenceTargets()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(list)
	}
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		getTodosHandler(w, r)
	case http.MethodPost:
		addTodoHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	mux.HandleFunc("/api/logout", corsMiddleware(logoutHandler))
	mux.HandleFunc("/api/me", corsMiddleware(authMiddleware(meHandler)))
	
	mux.HandleFunc("/api/seed", corsMiddleware(authMiddleware(seedHandler)))
	mux.HandleFunc("/api/items", corsMiddleware(authMiddleware(itemsHandler)))
	mux.HandleFunc("/api/vendors", corsMiddleware(authMiddleware(vendorsHandler)))
	mux.HandleFunc("/api/pos", corsMiddleware(authMiddleware(posHandler)))
	mux.HandleFunc("/api/pos/items", corsMiddleware(authMiddleware(poItemsHandler)))
	mux.HandleFunc("/api/invoices", corsMiddleware(authMiddleware(invoicesHandler)))
	mux.HandleFunc("/api/invoices/items", corsMiddleware(authMiddleware(invoiceItemsHandler)))
	mux.HandleFunc("/api/aps", corsMiddleware(authMiddleware(apsHandler)))
	mux.HandleFunc("/api/containers", corsMiddleware(authMiddleware(containersHandler)))
	mux.HandleFunc("/api/containers/items", corsMiddleware(authMiddleware(containerItemsHandler)))
	mux.HandleFunc("/api/containers/bl", corsMiddleware(authMiddleware(containersByBLHandler)))
	mux.HandleFunc("/api/bls", corsMiddleware(authMiddleware(blsHandler)))
	mux.HandleFunc("/api/grs", corsMiddleware(authMiddleware(grsHandler)))
	mux.HandleFunc("/api/lots", corsMiddleware(authMiddleware(lotsHandler)))
	mux.HandleFunc("/api/lots/gr", corsMiddleware(authMiddleware(lotsByGRHandler)))
	mux.HandleFunc("/api/allocations", corsMiddleware(authMiddleware(allocationsHandler)))
	mux.HandleFunc("/api/bookings", corsMiddleware(authMiddleware(bookingsHandler)))
	mux.HandleFunc("/api/bookings/template", corsMiddleware(authMiddleware(bookingTemplateHandler)))
	mux.HandleFunc("/api/bookings/bulk-import", corsMiddleware(authMiddleware(bookingBulkImportHandler)))
	mux.HandleFunc("/api/references", corsMiddleware(authMiddleware(checkReferencesHandler)))
	mux.HandleFunc("/api/delete", corsMiddleware(authMiddleware(deleteRecordHandler)))
	mux.HandleFunc("/api/ap-target-groups", corsMiddleware(authMiddleware(apTargetGroupsHandler)))
	mux.HandleFunc("/api/ap-target-groups/items", corsMiddleware(authMiddleware(apTargetGroupItemsHandler)))
	mux.HandleFunc("/api/ap-target-groups/references", corsMiddleware(authMiddleware(apTargetGroupReferencesHandler)))
	mux.HandleFunc("/api/todos", corsMiddleware(authMiddleware(todosHandler)))
	mux.HandleFunc("/api/todos/toggle", corsMiddleware(authMiddleware(toggleTodoHandler)))
	mux.HandleFunc("/api/todos/delete", corsMiddleware(authMiddleware(deleteTodoHandler)))

	// 프론트엔드 정적 파일 서빙
	distFS, _ := fs.Sub(frontendAssets, "dist")
	mux.Handle("/", http.FileServer(http.FS(distFS)))

	// Clean up expired sessions every hour
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			cleanupExpiredSessions()
		}
	}()

	fmt.Println("Server starting on :8080...")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Server Error:", err)
	}
}
