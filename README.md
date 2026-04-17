Persiapan dan Instalasi

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
-d '{"username": "staf_biasa", "password": "password123", "role": "user"}'
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
-d '
{
   "title": "Task Title",
   "description": "Task Description",
   "status": "pending|completed",
   "due_date": "YYYY-MM-DD"
}'
```














