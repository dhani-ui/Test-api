Persiapan dan Instalasi

### 1. Setup Database
Buat database baru di PostgreSQL Anda dengan nama `apitest`.
```sql
CREATE DATABASE apitest;

```
Unduh Dependensi
​Buka terminal di dalam folder proyek, lalu jalankan:

```
go mod tidy
```
Cara Menjalankan Aplikasi
Pastikan server PostgreSQL dan Redis sudah berjalan.
Jalankan perintah berikut di terminal:
```
go run main.go
```



