package clubs

import (
	"archive/zip"
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/storage"
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
		h.handleUnauthorized(w, r, err)
		return
	}

	filesEntry, err := h.Service.Storage.ListDocuments("clubs", key)
	if err != nil {
		api.Error(w, r, t.Errorf("list documents failed: %w", err), http.StatusInternalServerError)
		return
	}

	if len(filesEntry) == 0 {
		api.Error(w, r, t.Errorf("no documents found"), http.StatusNotFound)
		return
	}

	h.serveZip(w, r, key, user, filesEntry)
}

func (h *ClubHandler) handleUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		api.Error(w, r, err, http.StatusInternalServerError)
	} else {
		api.Error(w, r, t.Errorf("unauthorized to view documents for this club"), http.StatusForbidden)
	}
}

func (h *ClubHandler) serveZip(w http.ResponseWriter, r *http.Request, key string, user *entities.User, filesEntry []storage.Document) {
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
		_ = h.addFileToZip(zipWriter, "clubs", key, doc.Name)
	}
}

func (h *ClubHandler) addFileToZip(zw *zip.Writer, category, key, filename string) error {
	path, err := h.Service.Storage.GetDocumentPath(category, key, filename)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := zw.Create(filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

// DeleteDocument handles document deletion.
func (h *ClubHandler) DeleteDocument(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	filename := ps.ByName("filename")

	club, err := h.Service.GetClub(r.Context(), key, user)
	if err != nil {
		api.Error(w, r, err, http.StatusForbidden)
		return
	}

	isAdmin := api.IsAdmin(*user)
	status := club.Membership.Status

	// Authorization logic:
	// 1. Admins can delete anything.
	// 2. Owners can only delete if status is NOT 'requested' and NOT 'active'.
	if !isAdmin {
		if status == "requested" || status == "active" {
			api.Error(w, r, t.Errorf("cannot delete documents while membership is requested or active"), http.StatusForbidden)
			return
		}
	}

	err = h.Service.Storage.DeleteDocument("clubs", key, filename)
	if err != nil {
		api.Error(w, r, t.Errorf("failed to delete document: %w", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
