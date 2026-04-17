Go Task API (To-Do List with Role-Based Auth)
​API Backend To-Do List yang tangguh dan siap produksi (Production-Ready). Dibangun menggunakan Golang, PostgreSQL (Database Utama), Redis (Caching), dan JWT (Autentikasi & Otorisasi).
​Fitur Utama
​Autentikasi Aman: Registrasi dan Login menggunakan enkripsi password bcrypt dan JSON Web Token (JWT).
​Otorisasi Berbasis Peran (RBAC):
​User: Hanya dapat melihat daftar tugas (GET).
​Admin: Memiliki hak akses penuh untuk membuat, memperbarui, dan menghapus tugas (POST, PUT, DELETE).
​CRUD Operations & Fitur Lanjutan: Mendukung pencarian (search), filter status, dan paginasi.
​Concurrent Caching: Optimasi performa menggunakan Redis. Pembersihan cache dilakukan secara asinkron di background (goroutine) agar tidak menghalangi respons API.
​Logging Terpusat: Semua error dan aktivitas dicatat ke dalam file app.log lengkap dengan informasi waktu dan baris kode.
​Automated Unit Testing: Dilengkapi dengan pengujian otomatis menggunakan SQLite In-Memory.
​Prasyarat (Tech Stack)
​Sebelum menjalankan aplikasi, pastikan sistem Anda memiliki:
​Go (versi 1.18 atau lebih baru)
​PostgreSQL (berjalan di port default 5432)
​Redis (berjalan di port default 6379)
​Persiapan dan Instalasi
​1. Setup Database
​Buat database baru di PostgreSQL Anda dengan nama apitest.
CREATE DATABASE apitest;
​(Catatan: Secara default, aplikasi mencoba terhubung menggunakan username postgres dan password postgres. Anda dapat menyesuaikan kredensial ini pada variabel dsn di dalam fungsi main() di file main.go).
​2. Unduh Dependensi
​Buka terminal di dalam folder proyek, lalu jalankan:
go mod tidy
​Cara Menjalankan Aplikasi
​Pastikan server PostgreSQL dan Redis sudah berjalan.
​Jalankan perintah berikut di terminal:
go run main.go
​Server akan berjalan di http://localhost:8080. File app.log akan otomatis dibuat untuk mencatat riwayat server.
​Cara Menjalankan Unit Test
​Proyek ini menggunakan SQLite In-Memory untuk pengujian sehingga database utama Anda tetap aman. Jalankan perintah berikut:
go test -v ./...
​Dokumentasi API
​1. Autentikasi (Public)
​Metode POST - Endpoint: /register - Deskripsi: Mendaftarkan user baru (Role: admin atau user)
Metode POST - Endpoint: /login - Deskripsi: Login dan mendapatkan token JWT
​Contoh Payload /register & /login:
{
"username": "admin_utama",
"password": "rahasia123",
"role": "admin"
}
(Kosongkan field role saat register jika ingin mendaftar sebagai user biasa).
​2. Manajemen Tugas (Protected via JWT)
​Semua endpoint di bawah ini wajib menyertakan token JWT pada HTTP Header:
Authorization: Bearer <TOKEN_ANDA>
​GET /tasks (Akses: Admin & User) - Mengambil semua tugas (Support Pagination)
​GET /tasks/:id (Akses: Admin & User) - Mengambil detail satu tugas
​POST /tasks (Akses: Admin Saja) - Membuat tugas baru
​PUT /tasks/:id (Akses: Admin Saja) - Memperbarui data tugas
​DELETE /tasks/:id (Akses: Admin Saja) - Menghapus tugas
​Parameter Query untuk GET /tasks:
​status: Filter berdasarkan status (pending atau completed).
​page: Nomor halaman (default: 1).
​limit: Jumlah data per halaman (default: 10).
​search: Mencari kata kunci pada judul atau deskripsi.
​Contoh Payload POST /tasks:
{
"title": "Selesaikan Laporan Bulanan",
"description": "Laporan keuangan bulan April",
"status": "pending",
"due_date": "2026-04-30"
}
