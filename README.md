Persiapan dan Instalasi

### 1. Setup Database
Buat database baru di PostgreSQL Anda dengan nama `apitest`.
```sql
CREATE DATABASE apitest;

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







