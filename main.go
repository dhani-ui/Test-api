package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// --- Konfigurasi Global ---
var (
	DB          *gorm.DB
	RedisClient *redis.Client
	ctx         = context.Background()
	jwtSecret   = []byte("super-secret-key") // Gunakan env variable di production
)

// --- Model Data ---
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique" json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"` // Contoh: "admin" atau "user"
}

type Task struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"required,oneof=pending completed"`
	DueDate     string `json:"due_date" binding:"required,datetime=2006-01-02"` // Validasi format YYYY-MM-DD
}

type Pagination struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	TotalTasks  int64 `json:"total_tasks"`
}

// --- Setup Logging ---
func initLogger() {
	// Membuka atau membuat file app.log
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Gagal membuka file log: %v", err)
	}

	// Menulis output ke Terminal dan File secara bersamaan
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	// Mengatur format log: Tanggal | Waktu | File & Baris Kode
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// --- Setup Database & Redis ---
func initDB(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	// Migrasi tabel User dan Task
	DB.AutoMigrate(&User{}, &Task{})
	log.Println("Database PostgreSQL terhubung dan termigrasi.")
}

func initRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Peringatan: Gagal terhubung ke Redis: %v", err)
	} else {
		log.Println("Redis terhubung.")
	}
}

// --- Middlewares ---

// Middleware Autentikasi: Mengecek Token Valid
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			log.Printf("Auth Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Middleware Otorisasi: Mengecek Hak Akses Role
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse token kembali untuk mengambil claims (sudah dipastikan valid oleh AuthMiddleware)
		token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userRole := claims["role"].(string)
			if userRole != requiredRole {
				c.JSON(http.StatusForbidden, gin.H{"error": "Otorisasi ditolak. Anda bukan " + requiredRole})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// --- Concurrent Execution: Cache Invalidation ---
func invalidateCacheConcurrently() {
	if RedisClient == nil {
		return
	}

	go func() {
		iter := RedisClient.Scan(ctx, 0, "tasks*", 0).Iterator()
		for iter.Next(ctx) {
			RedisClient.Del(ctx, iter.Val())
		}
		if err := iter.Err(); err != nil {
			log.Printf("ERROR saat invalidasi cache Redis: %v", err)
		}
	}()
}

// --- API Handlers (Auth) ---

// Register User Baru
func register(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("Validation Error (Register): %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Bcrypt Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}
	user.Password = string(hashedPassword)

	// Default role jika tidak diisi
	if user.Role == "" {
		user.Role = "user"
	}

	if err := DB.Create(&user).Error; err != nil {
		log.Printf("DB Error (Register): %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username mungkin sudah terdaftar"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User berhasil didaftarkan"})
}

// Login & Generate Token
func login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	// Bandingkan password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	// Buat JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("JWT Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// --- API Handlers (Tasks) ---

// Create Task
func createTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		log.Printf("Validation Error (Create): %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if err := DB.Create(&task).Error; err != nil {
		log.Printf("DB Error (Create): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	invalidateCacheConcurrently()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task created successfully",
		"task":    task,
	})
}

// Get All Tasks
func getTasks(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	cacheKey := fmt.Sprintf("tasks:page:%d:limit:%d:status:%s:search:%s", page, limit, status, search)

	if RedisClient != nil {
		cachedData, err := RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var response map[string]interface{}
			json.Unmarshal([]byte(cachedData), &response)
			c.JSON(http.StatusOK, response)
			return
		}
	}

	query := DB.Model(&Task{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var totalTasks int64
	query.Count(&totalTasks)

	offset := (page - 1) * limit
	var tasks []Task
	if err := query.Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		log.Printf("DB Error (GetAll): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	totalPages := int(math.Ceil(float64(totalTasks) / float64(limit)))
	response := gin.H{
		"tasks": tasks,
		"pagination": Pagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalTasks:  totalTasks,
		},
	}

	if RedisClient != nil {
		if jsonData, err := json.Marshal(response); err == nil {
			RedisClient.Set(ctx, cacheKey, jsonData, 5*time.Minute)
		}
	}

	c.JSON(http.StatusOK, response)
}

// Get Task by ID
func getTaskByID(c *gin.Context) {
	id := c.Param("id")
	cacheKey := "task:" + id

	if RedisClient != nil {
		cachedData, err := RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var task Task
			json.Unmarshal([]byte(cachedData), &task)
			c.JSON(http.StatusOK, task)
			return
		}
	}

	var task Task
	if err := DB.First(&task, id).Error; err != nil {
		log.Printf("DB Error (GetByID): %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if RedisClient != nil {
		if jsonData, err := json.Marshal(task); err == nil {
			RedisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute)
		}
	}

	c.JSON(http.StatusOK, task)
}

// Update Task
func updateTask(c *gin.Context) {
	id := c.Param("id")
	var task Task

	if err := DB.First(&task, id).Error; err != nil {
		log.Printf("DB Error (Update - NotFound): %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var input Task
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Validation Error (Update): %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	task.Title = input.Title
	task.Description = input.Description
	task.Status = input.Status
	task.DueDate = input.DueDate

	if err := DB.Save(&task).Error; err != nil {
		log.Printf("DB Error (Update): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	invalidateCacheConcurrently()

	c.JSON(http.StatusOK, gin.H{
		"message": "Task updated successfully",
		"task":    task,
	})
}

// Delete Task
func deleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := DB.Delete(&Task{}, id).Error; err != nil {
		log.Printf("DB Error (Delete): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	invalidateCacheConcurrently()

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}

// --- Setup Router ---
func setupRouter() *gin.Engine {
	r := gin.Default()

	// Rute Publik (Registrasi & Login)
	r.POST("/register", register)
	r.POST("/login", login)

	// Grup Rute yang Membutuhkan Autentikasi (Harus Login)
	api := r.Group("/tasks")
	api.Use(AuthMiddleware())
	{
		// Semua yang login (user biasa & admin) boleh melihat data
		api.GET("", getTasks)
		api.GET("/:id", getTaskByID)

		// Grup Rute yang Membutuhkan Otorisasi Khusus Admin
		adminRoutes := api.Group("")
		adminRoutes.Use(RoleMiddleware("admin"))
		{
			// Hanya user dengan role "admin" yang boleh menambah/mengubah/menghapus
			adminRoutes.POST("", createTask)
			adminRoutes.PUT("/:id", updateTask)
			adminRoutes.DELETE("/:id", deleteTask)
		}
	}
	return r
}

func main() {
	// Inisialisasi Logger paling pertama
	initLogger()

	// DSN default jika tidak dijalankan saat testing
	if os.Getenv("ENV") != "testing" {
		dsn := "host=localhost user=postgres password=postgres dbname=apitest port=5432 sslmode=disable TimeZone=Asia/Jakarta"
		initDB(dsn)
		initRedis()
	}

	r := setupRouter()
	log.Println("Server berjalan di port 8080...")
	r.Run(":8080")
}
