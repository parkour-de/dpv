package router

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func setupServer(t *testing.T, port string) *http.Server {
	os.Setenv("PORT", port)

	tempDir := t.TempDir()

	// Load original config
	data, err := os.ReadFile("../../config.yml")
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	// Override document_path
	if _, ok := config["storage"]; !ok {
		config["storage"] = make(map[string]interface{})
	}
	storage := config["storage"].(map[string]interface{})
	storage["document_path"] = tempDir

	// Marshal back
	newData, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	configPath := filepath.Join(tempDir, "config.yml")
	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Copy strings_de.ini so translations still work
	translations, err := os.ReadFile("../../strings_de.ini")
	if err == nil {
		_ = os.WriteFile(filepath.Join(tempDir, "strings_de.ini"), translations, 0644)
	}

	server := NewServer(configPath, true)
	go func() {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			t.Error(err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	return server
}

func TestServer(t *testing.T) {
	server := setupServer(t, "8081")
	defer server.Close()

	resp, err := http.Get("http://localhost:8081/dpv/users")
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	expectedContentType := "application/json"
	if resp.Header.Get("Content-Type") != expectedContentType {
		t.Errorf("handler returned unexpected content-type: got %v want %v",
			resp.Header.Get("Content-Type"), expectedContentType)
	}

	if err != nil {
		t.Error(err)
	}
}

func TestRegisterAndGetMe(t *testing.T) {
	server := setupServer(t, "8082")
	defer server.Close()

	// Register user
	reqBody := `{"email":"test@example.com","password":"TestPassword123!","firstname":"John","lastname":"Doe"}`
	resp, err := http.Post("http://localhost:8082/dpv/users", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		// print error body for debugging
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(b))
	}
	resp.Body.Close()

	// Get /dpv/users/me with basic auth
	req, err := http.NewRequest("GET", "http://localhost:8082/dpv/users/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("test@example.com", "TestPassword123!")
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"firstname":"John"`) || !strings.Contains(string(body), `"lastname":"Doe"`) {
		t.Errorf("response body does not contain correct firstname and lastname: %s", string(body))
	}
}

func TestVersion(t *testing.T) {
	server := setupServer(t, "8083")
	defer server.Close()

	resp, err := http.Get("http://localhost:8083/dpv/version")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("version body is empty")
	}
}

func TestDocumentUpload(t *testing.T) {
	server := setupServer(t, "8084")
	defer server.Close()

	client := &http.Client{}

	// Register
	regBody := `{"email":"uploader@example.com","password":"UploaderPass123!","firstname":"U","lastname":"L"}`
	http.Post("http://localhost:8084/dpv/users", "application/json", strings.NewReader(regBody))

	// Create Club
	clubBody := `{"name":"Upload Club","legal_form":"e.V."}`
	req, _ := http.NewRequest("POST", "http://localhost:8084/dpv/clubs", strings.NewReader(clubBody))
	req.SetBasicAuth("uploader@example.com", "UploaderPass123!")
	resp, _ := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Club creation failed: %d", resp.StatusCode)
	}
	// Read the club key from the response
	b, _ := io.ReadAll(resp.Body)
	keyStart := strings.Index(string(b), `"_key":"`) + 8
	keyEnd := strings.Index(string(b)[keyStart:], `"`) + keyStart
	clubKey := string(b[keyStart:keyEnd])

	// Upload Document
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("document", "test.pdf")
	part.Write([]byte("mock pdf content"))
	writer.Close()

	uploadURL := fmt.Sprintf("http://localhost:8084/dpv/clubs/%s/documents", clubKey)
	req, _ = http.NewRequest("POST", uploadURL, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Language", "de")
	req.SetBasicAuth("uploader@example.com", "UploaderPass123!")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(b))
	}

	resBody, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(resBody), "erfolgreich hochgeladen") {
		t.Errorf("unexpected response: %s", string(resBody))
	}
}
