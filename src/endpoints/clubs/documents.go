package clubs

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// UploadDocument handles file uploads for a club.
func (h *ClubHandler) UploadDocument(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, ok := r.Context().Value("user").(*entities.User)
	if !ok || user == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	// Check auth: user must be board member or admin
	administered, err := h.Service.ListClubs(r.Context(), user.Key)
	if err != nil {
		api.Error(w, r, err, http.StatusInternalServerError)
		return
	}

	authorized := false
	if api.IsAdmin(*user) {
		authorized = true
	} else {
		for _, c := range administered {
			if c.GetKey() == key {
				authorized = true
				break
			}
		}
	}

	if !authorized {
		api.Error(w, r, t.Errorf("unauthorized to upload documents for this club"), http.StatusForbidden)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		api.Error(w, r, t.Errorf("parse multipart form failed: %w", err), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		api.Error(w, r, t.Errorf("get document from form failed: %w", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save using storage service
	filename, err := h.Service.Storage.SaveDocument("clubs", key, header.Filename, file)
	if err != nil {
		api.Error(w, r, t.Errorf("save document failed: %w", err), http.StatusInternalServerError)
		return
	}

	api.SuccessJson(w, r, map[string]string{
		"message":  t.T("document uploaded successfully"),
		"filename": filename,
	})
}
