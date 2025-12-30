package clubs

import (
	"context"
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
	if !h.isAuthorized(r.Context(), user, key) {
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

// ListDocuments lists documents for a club.
func (h *ClubHandler) ListDocuments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, ok := r.Context().Value("user").(*entities.User)
	if !ok || user == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	if !h.isAuthorized(r.Context(), user, key) {
		api.Error(w, r, t.Errorf("unauthorized to view documents for this club"), http.StatusForbidden)
		return
	}

	files, err := h.Service.Storage.ListDocuments("clubs", key)
	if err != nil {
		api.Error(w, r, t.Errorf("list documents failed: %w", err), http.StatusInternalServerError)
		return
	}

	api.SuccessJson(w, r, files)
}

// GetDocument serves a document for a club.
func (h *ClubHandler) GetDocument(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, ok := r.Context().Value("user").(*entities.User)
	if !ok || user == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	filename := ps.ByName("filename")

	if !h.isAuthorized(r.Context(), user, key) {
		api.Error(w, r, t.Errorf("unauthorized to view documents for this club"), http.StatusForbidden)
		return
	}

	path, err := h.Service.Storage.GetDocumentPath("clubs", key, filename)
	if err != nil {
		api.Error(w, r, t.Errorf("document not found"), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, path)
}

// Helper to check authorization
func (h *ClubHandler) isAuthorized(ctx context.Context, user *entities.User, clubKey string) bool {
	if api.IsAdmin(*user) {
		return true
	}
	administered, err := h.Service.ListClubs(ctx, user.Key)
	if err != nil {
		return false
	}
	for _, c := range administered {
		if c.GetKey() == clubKey {
			return true
		}
	}
	return false
}
