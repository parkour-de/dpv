package audit

import (
	"bufio"
	"dpv/dpv/src/api"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/t"
	"encoding/json"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

type AuditHandler struct {
	DB *graph.Db
}

func NewHandler(db *graph.Db) *AuditHandler {
	return &AuditHandler{DB: db}
}

func (h *AuditHandler) Get(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := api.RequireGlobalAdmin(r, h.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	filePath := h.DB.Audit.FilePath
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			api.SuccessJson(w, r, []interface{}{})
			return
		}
		api.Error(w, r, t.Errorf("could not open audit log: %w", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var logs []interface{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			logs = append(logs, entry)
		}
	}

	// Reverse to show latest first
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	// Limit to last 100 for now
	if len(logs) > 100 {
		logs = logs[:100]
	}

	api.SuccessJson(w, r, logs)
}
