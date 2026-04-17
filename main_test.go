package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// --- Helper: Setup Test App ---
func setupTestApp() *gin.Engine {
	// Set env untuk testing agar tidak menggunakan Postgres dan Redis
	os.Setenv("ENV", "testing")
	gin.SetMode(gin.TestMode)

	// Gunakan SQLite In-Memory untuk unit test cepat & bersih
	var err error
	DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect test database")
	}
	
	// Migrasi tabel User dan Task untuk pengujian
	DB.AutoMigrate(&User{}, &Task{})

	// Bypass Redis
	RedisClient = nil

	return setupRouter()
}

// --- Helper: Register & Login untuk mendapatkan Token ---
func getTestToken(r *gin.Engine, username, password, role string) string {
	// 1. Register
	regData := map[string]string{
		"username": username,
		"password": password,
		"role":     role,
	}
	body, _ := json.Marshal(regData)
	reqReg, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	wReg := httptest.NewRecorder()
	r.ServeHTTP(wReg, reqReg)

	// 2. Login
	loginData := map[string]string{
		"username": username,
		"password": password,
	}
	bodyLogin, _ := json.Marshal(loginData)
	reqLogin, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(bodyLogin))
	wLogin := httptest.NewRecorder()
	r.ServeHTTP(wLogin, reqLogin)

	// 3. Ekstrak Token
	var res map[string]string
	json.Unmarshal(wLogin.Body.Bytes(), &res)
	return res["token"]
}

// --- Test Cases ---

func TestRegisterAndLogin(t *testing.T) {
	r := setupTestApp()

	// Test Register
	regData := map[string]string{"username": "testuser", "password": "123", "role": "user"}
	body, _ := json.Marshal(regData)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %v", w.Code)
	}

	// Test Login
	loginData := map[string]string{"username": "testuser", "password": "123"}
	bodyLogin, _ := json.Marshal(loginData)
	reqLogin, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(bodyLogin))
	wLogin := httptest.NewRecorder()
	r.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %v", wLogin.Code)
	}
}

func TestCreateTask_AsAdmin(t *testing.T) {
	r := setupTestApp()
	
	// Dapatkan token sebagai Admin
	token := getTestToken(r, "admin_tester", "rahasia", "admin")

	taskData := map[string]string{
		"title":       "Task Khusus Admin",
		"description": "Test admin task",
		"status":      "pending",
		"due_date":    "2026-12-31",
	}
	body, _ := json.Marshal(taskData)

	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Harusnya BERHASIL (201) karena rolenya Admin
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 for admin, got %v", w.Code)
	}
}

func TestCreateTask_AsUser(t *testing.T) {
	r := setupTestApp()
	
	// Dapatkan token sebagai User biasa
	token := getTestToken(r, "user_tester", "rahasia", "user")

	taskData := map[string]string{
		"title":       "Task User",
		"status":      "pending",
		"due_date":    "2026-12-31",
	}
	body, _ := json.Marshal(taskData)

	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Harusnya DITOLAK (403) karena rolenya bukan Admin
	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected status 403 Forbidden for regular user, got %v", w.Code)
	}
}

func TestGetTasks(t *testing.T) {
	r := setupTestApp()
	
	// Setup data awal langsung ke DB (Bypass API untuk kecepatan)
	DB.Create(&Task{Title: "Task Selesai", Status: "completed", DueDate: "2026-10-10"})
	DB.Create(&Task{Title: "Task Belum", Status: "pending", DueDate: "2026-10-11"})

	// Dapatkan token (Siapapun yang login bisa melakukan GET)
	token := getTestToken(r, "viewer", "123", "user")

	// Lakukan request dengan filter
	req, _ := http.NewRequest("GET", "/tasks?status=completed", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %v", w.Code)
	}

	var response struct {
		Tasks []Task `json:"tasks"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	// Pastikan filter berjalan dengan benar (Hanya mendapat 1 task)
	if len(response.Tasks) != 1 || response.Tasks[0].Title != "Task Selesai" {
		t.Errorf("Filter GET failed, expected length 1 and title 'Task Selesai'")
	}
}
