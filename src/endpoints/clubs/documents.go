package clubs

import (
	"archive/zip"
	"dpv/dpv/src/api"
	"dpv/dpv/src/repository/t"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

// UploadDocument handles file uploads for a club.
func (h *ClubHandler) UploadDocument(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	if authorized, err := h.Service.IsAuthorized(r.Context(), user, key); err != nil || !authorized {
		if err != nil {
			api.Error(w, r, err, http.StatusInternalServerError)
		} else {
			api.Error(w, r, t.Errorf("unauthorized to upload documents for this club"), http.StatusForbidden)
		}
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(50 << 20); err != nil { // 50MB limit
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
		"message":  t.T(t.Errorf("document uploaded successfully"), api.DetectLanguage(r)),
		"filename": filename,
	})
}

// ListDocuments lists documents for a club.
func (h *ClubHandler) ListDocuments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	if authorized, err := h.Service.IsAuthorized(r.Context(), user, key); err != nil || !authorized {
		if err != nil {
			api.Error(w, r, err, http.StatusInternalServerError)
		} else {
			api.Error(w, r, t.Errorf("unauthorized to view documents for this club"), http.StatusForbidden)
		}
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
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	filename := ps.ByName("filename")

	if authorized, err := h.Service.IsAuthorized(r.Context(), user, key); err != nil || !authorized {
		if err != nil {
			api.Error(w, r, err, http.StatusInternalServerError)
		} else {
			api.Error(w, r, t.Errorf("unauthorized to view documents for this club"), http.StatusForbidden)
		}
		return
	}

	path, err := h.Service.Storage.GetDocumentPath("clubs", key, filename)
	if err != nil {
		api.Error(w, r, t.Errorf("document not found"), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, path)
}

// DownloadAllDocuments streams a zip of all documents
func (h *ClubHandler) DownloadAllDocuments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	if authorized, err := h.Service.IsAuthorized(r.Context(), user, key); err != nil || !authorized {
		if err != nil {
			api.Error(w, r, err, http.StatusInternalServerError)
		} else {
			api.Error(w, r, t.Errorf("unauthorized to view documents for this club"), http.StatusForbidden)
		}
		return
	}

	// List files
	filesEntry, err := h.Service.Storage.ListDocuments("clubs", key)
	if err != nil {
		api.Error(w, r, t.Errorf("list documents failed: %w", err), http.StatusInternalServerError)
		return
	}

	if len(filesEntry) == 0 {
		api.Error(w, r, t.Errorf("no documents found"), http.StatusNotFound)
		return
	}

	club, _ := h.Service.GetClub(r.Context(), key, user)
	sanitizedClubName := "documents"
	if club != nil {
		sanitizedClubName = api.SanitizeFilename(club.Name)
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-documents.zip\"", sanitizedClubName))

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, doc := range filesEntry {
		path, err := h.Service.Storage.GetDocumentPath("clubs", key, doc.Name)
		if err != nil {
			continue // skip
		}

		f, err := os.Open(path)
		if err != nil {
			continue
		}

		// Create zip entry
		w, err := zipWriter.Create(doc.Name)
		if err != nil {
			f.Close()
			continue
		}
		if _, err := io.Copy(w, f); err != nil {
			f.Close()
			continue
		}
		f.Close()
	}
}
