package router

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	port := os.Getenv("PORT")
	os.Setenv("PORT", "8081")
	defer os.Setenv("PORT", port)

	server := NewServer("../../config.yml", true)
	go func() {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			t.Error(err)
		}
	}()
	defer server.Close()

	// Wait 50 milliseconds for server to start listening to requests
	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get("http://localhost:8081/api/users")
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

	/*expectedBody := `"hello world"`
	if string(body) != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v",
			string(body), expectedBody)
	}*/
}

func TestRegisterAndGetMe(t *testing.T) {
	os.Setenv("PORT", "8082")
	server := NewServer("../../config.yml", true)
	go func() {
		_ = server.ListenAndServe()
	}()
	defer server.Close()
	time.Sleep(50 * time.Millisecond)

	// Register user
	reqBody := `{"email":"test@example.com","password":"testpass","vorname":"John","name":"Doe"}`
	resp, err := http.Post("http://localhost:8082/dpv/users", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Get /dpv/users/me with basic auth
	req, err := http.NewRequest("GET", "http://localhost:8082/dpv/users/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("test@example.com", "testpass")
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
	if !strings.Contains(string(body), `"vorname":"John"`) || !strings.Contains(string(body), `"name":"Doe"`) {
		t.Errorf("response body does not contain correct vorname and name: %s", string(body))
	}
}
