## Hasil Pengujian SQL Injection

curl command yang di jalankan:

```
curl -i -X POST http://localhost:3000/expense/create \
  -d "wallet=2" \
  -d "amount=1000" \
  -d "category=lain" \
  --data-urlencode "title=x');DROP TABLE expenses;--"
```

Response dari server:

```
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

```
SELECT * FROM expenses WHERE title LIKE '%DROP%';
```

Result:

```
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
