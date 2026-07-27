# Konvensi API

Prefix versi adalah `/api/v1`. Response sukses memakai `success`, `message`, `data`, dan `meta`. Response gagal memakai `success`, `message`, `code`, `errors`, dan `request_id`.

Daftar data mendukung `page`, `per_page`, `search`, serta filter khusus domain. Batas maksimal `per_page` adalah 100.

Access token dikirim melalui `Authorization: Bearer <token>`. Refresh token dirotasi setiap kali digunakan. Perubahan administratif yang membutuhkan alasan dapat mengirim header `X-Change-Reason`.

Status HTTP utama:

- `200` berhasil.
- `201` data dibuat.
- `401` belum terautentikasi.
- `403` tidak memiliki permission.
- `404` data tidak ditemukan.
- `409` konflik aturan bisnis.
- `422` payload atau transisi tidak valid.
- `500` kesalahan internal tanpa detail sensitif.
