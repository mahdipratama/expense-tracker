# Expense Tracker

## 📌 Project Objective

Expense Tracker adalah aplikasi web berbasis Go (Golang) untuk mengelola dan melacak pengeluaran harian serta saldo dompet (wallet).
Proyek ini dikembangkan sebagai praktik pembelajaran (Chapter 4), di mana **Database merupakan _source of truth_** dan UI/HTML belum menjadi fokus utama (lebih menitikberatkan pada pengembangan API/Backend).

- **Go Version:** 1.26.6
- **Module:** `github.com/mahdipratama/expense-tracker`
- **Progress:** Milestone 2 ✅ | Milestone 3 ✅ | Milestone 4 ✅ | Milestone 5 ✅ | Milestone 6 ✅ | Milestone 7 ✅

---

## 🚀 Cara Menjalankan Aplikasi

### 1. Instalasi

1. Pastikan Anda telah menginstal [Go](https://golang.org/) (versi 1.26.6 atau terbaru) dan MySQL.
2. Clone repositori ini dan masuk ke root folder.
3. Jalankan perintah berikut untuk mengunduh dependencies:
   ```bash
   go mod download
   ```

### 2. Setup MySQL

Buat database di MySQL lokal Anda:

```sql
CREATE DATABASE expense_tracker CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

_(Opsional tetapi disarankan: Buat user MySQL khusus aplikasi selain `root` untuk alasan keamanan)._

### 3. Configuration

DSN (Data Source Name) dikonfigurasi melalui environment variable. Anda dapat mengaturnya di terminal sebelum menjalankan aplikasi:

```bash
export DB_DSN="username:password@tcp(localhost:3306)/expense_tracker?parseTime=true"
go run ./cmd/web
```

_Catatan: Jika variabel ini tidak diset, aplikasi akan menggunakan nilai default (jika ada) atau mengembalikan pesan error._

---

## 🔌 API Endpoints

Aplikasi ini berjalan di port default (biasanya `4000` atau `3000`). Anda dapat menguji endpoint berikut menggunakan `curl` atau Postman.

#### Wallet

- `GET /wallets` — List semua wallet
- `GET /wallet?id=N` — Detail wallet tertentu
- `POST /wallet/create` — Buat wallet baru

#### Expense

- `GET /expenses` — List semua expense
- `GET /expenses?wallet=N` — List expense dari wallet tertentu
- `POST /expense/create` — Buat expense baru

#### Transfer

- `POST /transfer` — Transfer saldo antar-wallet

#### Health

- `GET /healthz` — Check database connection (Health check)

---

## 🗄️ Database

### Schema

Terdapat dua entitas utama dalam sistem ini: `wallets` dan `expenses`.

**Tabel `wallets`**

```sql
CREATE TABLE wallets (
    id INT NOT NULL AUTO_INCREMENT ,
    name VARCHAR(100) NOT NULL,
    balance INT NOT NULL DEFAULT ,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_balance_positive CHECK (balance >= 0)
);
```

**Tabel `expenses`**

```sql
CREATE TABLE expenses (
    id INT NOT NULL AUTO_INCREMENT,
    wallet_id INT NOT NULL,
    amount INT NOT NULL,
    category VARCHAR(100) NOT NULL,
    spent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_wallet_id` (`wallet_id`),
    CONSTRAINT `expense_ibfk_1` FOREIGN KEY (`wallet_id`) REFERENCES `wallets` (`id`),
    CONSTRAINT `expense_chk_1` CHECK ((`amount` > 0))
);
```

### Index Explanation

Selain _Primary Key_, terdapat indeks tambahan pada tabel `expenses`:

```sql
CREATE INDEX idx_wallet_id ON expenses(wallet_id);
```

**Alasan:** Pencarian dan kalkulasi riwayat pengeluaran hampir selalu difilter berdasarkan `wallet_id`. Indeks ini mencegah _full table scan_ dan mempercepat proses _query_. Penggunaan query dinamis (_dynamic query building_) untuk klausa `WHERE` yang opsional juga diimplementasikan agar indeks tidak di-_bypass_.

### Seed Data

Database dilengkapi dengan data awal (_seed data_) untuk testing:

- **2 Wallets:** "Tunai" (500,000) dan "Bank" (1,500,000).
- **5 Expenses** dari berbagai kategori.
- Data ini memungkinkan endpoint `GET` dapat langsung diuji (misalnya `GET /wallets` atau `GET /expenses`) tanpa perlu membuat data manual secara berulang.

---

## 🏗️ Architecture

Aplikasi ini menggunakan pola arsitektur modular yang memisahkan dependensi (_dependency injection_) melalui _application struct_.

**Alur Aplikasi:**
`main` ➔ `application` ➔ `handlers` ➔ `models` ➔ `database/sql` ➔ `MySQL`

1. **main & application**: Titik masuk (_entry point_) aplikasi yang mengatur konfigurasi, menginisiasi koneksi database, dan menyuntikkan dependensi (_dependency injection_) ke dalam struct `application`.
2. **handlers**: Bertanggung jawab menerima HTTP request, memvalidasi input, memanggil fungsi yang sesuai di _Model_, dan merender HTTP response.
3. **models**: Lapisan _business logic_ dan _database abstraction_. Lapisan ini menggunakan metode seperti `Query()`, `rows.Close()`, dan `rows.Err()` untuk mengambil _multiple rows_ serta mendukung parameter opsional dengan _pointer_.
4. **database/sql & MySQL**: _Driver_ bawaan Go yang menangani komunikasi level bawah dengan MySQL secara aman.

---

## 🧠 Design Decisions & Key Learnings

Dokumentasi ini mencatat keputusan teknis utama (_engineering decisions_) dan pola yang diterapkan dalam proyek:

1. **Keamanan POST Endpoints (Request Body)**
   _Decision:_ Semua POST endpoint menerima data melalui request body (`application/x-www-form-urlencoded`), bukan melalui query string di URL.
   _Rationale:_ Mencegah data sensitif terekspos di log server atau _browser history_, serta mengikuti standar semantik HTTP yang benar untuk mutasi data.

2. **Error Handling Pattern (Type Assertion)**
   _Decision:_ Memetakan error MySQL menjadi _Domain Error_ yang mudah dibaca.
   _Rationale:_ Error spesifik dari MySQL ditangkap menggunakan _Type Assertion_ di layer `models`, lalu diterjemahkan menjadi error domain (seperti `ErrDuplicateWallet`, `ErrInsufficientBalance`). Handler kemudian menggunakan error domain ini untuk menentukan HTTP Status Code dan response yang tepat kepada klien.

3. **Strategi Transaksi (Transaction Atomicity)**
   _Decision:_ Menggunakan pola standar `Begin()` ➔ `defer tx.Rollback()` ➔ _operations_ ➔ `Commit()` untuk proses pencatatan pengeluaran.
   _Rationale:_ Menjamin sifat _atomic_ (berhasil semua atau dibatalkan semua). Jika ada salah satu query yang gagal di tengah proses, `Rollback` akan tereksekusi secara otomatis, menjaga konsistensi database (terbukti pada skenario rollback di mana saldo tidak berubah saat transaksi gagal). Pengecekan `rowsAffected == 0` juga digunakan untuk mendeteksi _wallet_ yang tidak ditemukan saat proses transfer.

4. **Strategi Negative Balance & Row Locking (`FOR UPDATE`)**
   _Decision:_ Menggunakan kueri `SELECT balance ... FOR UPDATE` saat akan memvalidasi dan memotong saldo.
   _Rationale:_ Mencegah terjadinya _race condition_ jika ada beberapa _request_ pengeluaran yang datang di saat bersamaan. _Row locking_ ini mengunci baris spesifik pada tabel `wallets` sementara transaksi berlangsung.
   _(Catatan untuk Transfer:_ Pada implementasi transfer antardompet, aplikasi akan mengunci kedua dompet dengan **urutan ID terkecil lebih dulu** untuk mencegah terjadinya _deadlock_ pada database).

5. **Penggunaan `LastInsertId()`**
   _Decision:_ Memanfaatkan fungsi `LastInsertId()` setelah operasi `INSERT`.
   _Rationale:_ Memungkinkan aplikasi untuk langsung mendapatkan _auto-increment ID_ dari _expense_ atau _wallet_ yang baru saja dibuat, yang sangat penting untuk pencatatan respons JSON/HTTP ke _client_.

## 🔒 Security Testing (SQL Injection Prevention)

Aplikasi ini aman dari serangan _SQL Injection_ karena menggunakan **Prepared Statements** melalui _driver_ `database/sql`. Saat query dikirim, database driver secara otomatis memisahkan antara struktur perintah SQL dan data (parameter).

Berikut adalah hasil pengujian yang membuktikan bahwa input berbahaya diperlakukan murni sebagai nilai teks (_literal string_), bukan instruksi SQL yang dapat dieksekusi.
curl command yang di jalankan:

```bash
curl -i -X POST http://localhost:3000/expense/create \
  -d "wallet=2" \
  -d "amount=1000" \
  -d "category=lain" \
  --data-urlencode "title=x');DROP TABLE expenses;--"
```

Response dari server:

```bash
curl -i -X POST http://localhost:3000/expense/create \
  -d "wallet=2" \
  -d "amount=1000" \
  -d "category=lain" \
  --data-urlencode "title=x');DROP TABLE expenses;--"
HTTP/1.1 201 Created
Content-Type: application/json
Date: Wed, 21 Oct 2026 03:48:55 GMT
Content-Length: 138

{"ID":18,"WalletID":2,"Title":"x');DROP TABLE expenses;--","Amount":1000,"Category":"lain","SpentAt":"2026-10-21T10:48:54.9492498+07:00"}
```

Query ke DB untuk verifikasi:

```bash
SELECT * FROM expenses WHERE title LIKE '%DROP%';
```

Result:

```bash
mysql> SELECT * FROM expense WHERE title LIKE '%DROP%';
+----+-----------+----------------------------+--------+----------+---------------------+
| id | wallet_id | title                      | amount | category | spent_at            |
+----+-----------+----------------------------+--------+----------+---------------------+
| 18 |         2 | x');DROP TABLE expenses;-- |   1000 | lain     | 2026-10-21 03:48:55 |
+----+-----------+----------------------------+--------+----------+---------------------+
1 row in set (0.00 sec)
```

Penjelasan singkat:
Input tidak dieksekusi sebagai perintah SQL karena aplikasi menggunakan mekanisme Prepared Statements (menggunakan placeholder ? pada query). Saat query dikirim, database driver secara otomatis memisahkan antara struktur perintah SQL dan data (parameter). Input berbahaya tersebut diperlakukan murni sebagai nilai teks (literal string) untuk disimpan, bukan sebagai instruksi SQL yang dapat dieksekusi.
Jika aplikasi menggunakan string concatenation tanpa placeholder (misalnya `fmt.Sprintf("INSERT INTO expenses (title) VALUES ('%s')", title))`, maka input `x');DROP TABLE expenses;--` akan menjadi bagian dari SQL statement dan dieksekusi oleh database.
