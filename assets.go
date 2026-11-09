// Package fishermanweather nhúng toàn bộ frontend tĩnh vào binary Go.
// Dự án: binhdinh-fisherman-weather — Thời tiết biển cho ngư dân Bình Định.
// Tác giả: HTK <htk.contact@gmail.com> (https://zalo.me/0396241674)
// Hotline/Zalo: 0396241674 — Cập nhật: 09/11/2026
package fishermanweather

import "embed"

// PublicFS chứa thư mục public/ (HTML, CSS, JS) để phục vụ cả trên Vercel lẫn local.
//
//go:embed public
var PublicFS embed.FS