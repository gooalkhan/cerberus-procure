package main

import (
	"cerberus-procure/internal/logic"
	"cerberus-procure/internal/models"
	"cerberus-procure/internal/repository/memory"
	"encoding/json"
	"syscall/js"
)

var todoUC *logic.TodoUseCase
var authUC *logic.AuthUseCase
var procureUC *logic.ProcurementUseCase

func login(this js.Value, args []js.Value) interface{} {
	username := args[0].String()
	password := args[1].String()
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			user, err := authUC.Login(username, password)
			if err != nil {
				reject.Invoke(err.Error())
				return
			}
			b, _ := json.Marshal(user)
			js.Global().Get("localStorage").Call("setItem", "session_user", string(b))
			resolve.Invoke(string(b))
		}()
		return nil
	})

	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func getSession(this js.Value, args []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			userStr := js.Global().Get("localStorage").Call("getItem", "session_user")
			if userStr.IsNull() || userStr.IsUndefined() || userStr.String() == "" {
				reject.Invoke("No session")
				return
			}
			resolve.Invoke(userStr.String())
		}()
		return nil
	})

	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func getTodos(this js.Value, args []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		// reject := promiseArgs[1]

		go func() {
			todos, _ := todoUC.GetTodos()
			b, _ := json.Marshal(todos)
			resolve.Invoke(string(b))
		}()
		return nil
	})

	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func addTodo(this js.Value, args []js.Value) interface{} {
	title := args[0].String()
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		go func() {
			todo, _ := todoUC.AddTodo(title)
			b, _ := json.Marshal(todo)
			resolve.Invoke(string(b))
		}()
		return nil
	})
	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func toggleTodo(this js.Value, args []js.Value) interface{} {
	id := args[0].String()
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		go func() {
			todoUC.ToggleTodo(id)
			resolve.Invoke()
		}()
		return nil
	})
	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func deleteTodo(this js.Value, args []js.Value) interface{} {
	id := args[0].String()
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		go func() {
			todoUC.DeleteTodo(id)
			resolve.Invoke()
		}()
		return nil
	})
	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func main() {
	todoRepo := memory.NewMemoryTodoRepository()
	userRepo := memory.NewMemoryUserRepository()
	procureRepo := memory.NewMemoryProcurementRepository()

	todoUC = logic.NewTodoUseCase(todoRepo)
	authUC = logic.NewAuthUseCase(userRepo)
	procureUC = logic.NewProcurementUseCase(procureRepo)

	// Seed admin user
	authUC.Register("admin", "1234", "Administrator")

	js.Global().Set("login", js.FuncOf(login))
	js.Global().Set("getSession", js.FuncOf(getSession))
	js.Global().Set("getTodos", js.FuncOf(getTodos))
	js.Global().Set("addTodo", js.FuncOf(addTodo))
	js.Global().Set("toggleTodo", js.FuncOf(toggleTodo))
	js.Global().Set("deleteTodo", js.FuncOf(deleteTodo))

	// Procurement WASM Bridge
	procureObj := js.Global().Get("Object").New()

	procureObj.Set("seedData", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				procureUC.SeedData()
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	// Items
	procureObj.Set("getItems", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				items, _ := procureUC.GetItems()
				b, _ := json.Marshal(items)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("saveItem", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.ItemMaster
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SaveItem(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	// Vendors
	procureObj.Set("getVendors", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetVendors()
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("saveVendor", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.VendorMaster
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SaveVendor(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	// Invoices
	procureObj.Set("getCommercialInvoices", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetCommercialInvoices()
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("getCIAggregatedItems", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ciId := args[0].Int()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetCIAggregatedItems(ciId)
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("saveCommercialInvoice", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.CommercialInvoice
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SaveCommercialInvoice(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	// AP
	procureObj.Set("getAccountPayables", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetAccountPayables()
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("saveAccountPayable", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.AccountPayable
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SaveAccountPayable(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	// PO
	procureObj.Set("getPurchaseOrders", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetPurchaseOrders()
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("getPOItems", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		poId := args[0].Int()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetPOItemsByPOID(poId)
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("savePOItem", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.POItem
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SavePOItem(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))
	procureObj.Set("savePurchaseOrder", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.PurchaseOrder
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SavePurchaseOrder(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	// Containers
	procureObj.Set("getContainers", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetContainers()
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("saveContainer", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.Container
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SaveContainer(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	procureObj.Set("getContainersByBLID", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		id := args[0].Int()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetContainersByBLID(id)
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))

	procureObj.Set("saveContainerItem", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.ContainerItem
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SaveContainerItem(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	// BL
	procureObj.Set("getBLs", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetBLs()
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("saveBL", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.BL
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SaveBL(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	procureObj.Set("getContainerItemsByContainerID", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		id := args[0].Int()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetContainerItemsByContainerID(id)
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))

	// Goods Receipt
	procureObj.Set("getGoodsReceipts", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetGoodsReceipts()
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("getInventoryLotsByGRID", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		grId := args[0].Int()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetInventoryLotsByGRID(grId)
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("saveGoodsReceipt", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.GoodsReceipt
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SaveGoodsReceipt(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	// Inventory Lot
	procureObj.Set("getInventoryLots", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetInventoryLots()
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("saveInventoryLot", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.InventoryLot
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SaveInventoryLot(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	// Cost Allocation
	procureObj.Set("getCostAllocations", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetCostAllocations()
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))
	procureObj.Set("saveCostAllocation", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonStr := args[0].String()
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				var i models.CostAllocation
				json.Unmarshal([]byte(jsonStr), &i)
				procureUC.SaveCostAllocation(&i)
				resolve.Invoke()
			}()
			return nil
		}))
	}))

	procureObj.Set("getBookings", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) interface{} {
			resolve := pArgs[0]
			go func() {
				list, _ := procureUC.GetBookings()
				b, _ := json.Marshal(list)
				resolve.Invoke(string(b))
			}()
			return nil
		}))
	}))

	js.Global().Set("procureApi", procureObj)

	select {}
}
