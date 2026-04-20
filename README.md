Persiapan dan Instalasi

clone test api
```
git clone https://github.com/dhani-ui/Test-api.git
```

### 1. Setup Database
Buat database baru di PostgreSQL Anda dengan nama `apitest`.
```sql
CREATE ROLE postgres WITH LOGIN SUPERUSER PASSWORD 'postgres';

-- Buat database
CREATE DATABASE apitest;

-- Keluar dari psql
\q

```
Unduh Dependensi
​Buka terminal di dalam folder proyek, lalu jalankan:

```bash
go mod tidy
```
Jalankan Redis
```
redis-server
```
Cara Menjalankan Aplikasi
Pastikan server PostgreSQL dan Redis sudah berjalan.
Jalankan perintah berikut di terminal:
```bash
go run main.go
```
Server akan berjalan di http://localhost:8080. File app.log akan otomatis dibuat untuk mencatat riwayat server.

Cara Menjalankan Unit Test
​Proyek ini menggunakan SQLite In-Memory untuk pengujian sehingga database utama Anda tetap aman. 
Jalankan perintah berikut:
```bash
go test -v ./...
```

membuat akun admin
```
curl -X POST http://localhost:8080/register \
-H "Content-Type: application/json" \
-d '{"username": "juragan", "password": "123", "role": "admin"}'
```
membuat akun user biasa
```

curl -X POST http://localhost:8080/register \
-H "Content-Type: application/json" \
-d '{"username": "budi", "password": "budiaja", "role": "user"}'
```

Cek Log File

```
cat app.log

```

Enpoint - POST /tasks
```
TOKEN=$(curl -s -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{"username": "juragan", "password": "123"}' | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

curl -X POST http://localhost:8080/tasks \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
   "title": "Task Title",
   "description": "Task Description",
   "status": "pending",
   "due_date": "2026-04-20"
}'
```
endpoint GET /tasks
```
# 1. Ambil Token terbaru
TOKEN=$(curl -s -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{"username": "juragan", "password": "123"}' | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

# 2. Panggil endpoint GET
curl -X GET http://localhost:8080/tasks \
-H "Authorization: Bearer $TOKEN"

```
endpoint GET/task/:id 
```
# 1. Ambil Token terbaru
TOKEN=$(curl -s -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{"username": "juragan", "password": "123"}' | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

# 2. Ambil detail tugas dengan ID 1
curl -X GET http://localhost:8080/tasks/1 \
-H "Authorization: Bearer $TOKEN"
```
endpoint PUT /tasks/:id 
```
# 1. Ambil Token terbaru
TOKEN=$(curl -s -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{"username": "juragan", "password": "123"}' | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

# 2. Update tugas ID 1 dengan data VALID
curl -X PUT http://localhost:8080/tasks/1 \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
   "title": "Update Projek Backend",
   "description": "Fitur CRUD sudah hampir selesai",
   "status": "completed",
   "due_date": "2026-04-20"
}'
```
endpoint DELETE/task/:id 
```
# 1. Ambil Token terbaru (Login sebagai Admin)
TOKEN=$(curl -s -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{"username": "juragan", "password": "123"}' | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

# 2. Hapus tugas dengan ID 1
curl -X DELETE http://localhost:8080/tasks/1 \
-H "Authorization: Bearer $TOKEN"
```














