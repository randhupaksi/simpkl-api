package status

import (
	"strings"
)

var labels = map[string]string{
	"active":               "Aktif",
	"approved":             "Disetujui",
	"archived":             "Diarsipkan",
	"attention":            "Perlu perhatian",
	"awaiting_documents":   "Menunggu dokumen",
	"blocked":              "Diblokir",
	"cancelled":            "Dibatalkan",
	"completed":            "Selesai",
	"draft":                "Belum diajukan",
	"expired":              "Kedaluwarsa",
	"incomplete":           "Belum lengkap",
	"inactive":             "Tidak aktif",
	"locked":               "Terkunci",
	"not_participating":    "Tidak mengikuti PKL",
	"pending":              "Menunggu proses",
	"pending_verification": "Menunggu verifikasi",
	"placement_process":    "Dalam proses penempatan",
	"placed":               "Sudah ditempatkan",
	"preparation":          "Dalam persiapan",
	"ready":                "Siap mulai PKL",
	"rejected":             "Ditolak",
	"revision":             "Perlu diperbaiki",
	"revision_required":    "Perlu diperbaiki",
	"started":              "Sedang berjalan",
	"superseded":           "Digantikan",
	"transferred":          "Dipindahkan",
	"unplaced":             "Belum ditempatkan",
	"uploaded":             "Sudah diunggah",
	"valid":                "Terverifikasi",
	"verified":             "Terverifikasi",
	"verifying":            "Sedang diverifikasi",
}

func Label(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if label, ok := labels[normalized]; ok {
		return label
	}
	return strings.Title(strings.ReplaceAll(normalized, "_", " "))
}
