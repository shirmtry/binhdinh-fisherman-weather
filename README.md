# ⚓ binhdinh-fisherman-weather — Thời tiết biển ngư dân Bình Định

> HTK · Hotline/Zalo: **0396241674** · Email: htk.contact@gmail.com · Cập nhật: 09/11/2026

## 🌊 Vấn đề (mùa mưa lũ 2026)

Tháng 11/2026, ngư dân Bình Định ra khơi trong lúc bão xuất hiện bất ngờ: bão số 9, 10/2026
đổ bộ Quy Nhơn – Phù Cát – Hoài Nhơn, nhiều tàu cá gặp nạn khi **biển động cấp 7–8**.
Gió giật mạnh, sóng cao 4–5 m, tầm nhìn giảm chỉ trong vài giờ, trong khi bà con thiếu kênh
tra cứu nhanh dự báo gió – sóng theo từng vùng biển trước mỗi chuyến biển.
Dự án này cung cấp dự báo 72 giờ cho **6 vùng biển Bình Định** kèm cảnh báo đỏ/vàng/xanh
và khuyến cáo ra khơi cụ thể.

**Dữ liệu mô phỏng – cập nhật 09/11/2026** (không thay thế bản tin chính thức của Đài KTTV).

## 🗺️ 6 vùng biển

1. Nhơn Lý – Nhơn Hải
2. Cù Lao Xanh
3. Đề Gi – Cát Tiến
4. Mỹ Thành – Nhơn Thành
5. Hoài Nhơn – Tam Quan
6. Vịnh Quy Nhơn

## ✨ Tính năng

- **Go Gin API**:
  - `GET /api/zones` — 6 vùng biển, hiện trạng + tóm tắt 72h (gió max, sóng max, số giờ đỏ/vàng)
  - `GET /api/forecast?zone=<id>&hours=24|48|72` — chuỗi dự báo bước 3 giờ: gió (cấp Beaufort),
    gió giật, tốc độ km/h, sóng (m), tầm nhìn (km), khuyến cáo ra khơi
  - `GET /api/health`
- **Cảnh báo đỏ/vàng/xanh** theo gió + sóng:
  - ĐỎ: gió ≥ cấp 7 hoặc sóng ≥ 4 m → KHÔNG ra khơi
  - VÀNG: gió cấp 5–6 hoặc sóng 2,5–4 m → hạn chế, chỉ gần bờ
  - XANH: còn lại → có thể ra khơi gần bờ, theo dõi bão
- **Frontend Chart.js**: biểu đồ gió/sóng và tầm nhìn 24/48/72 giờ cho vùng chọn,
  thẻ 6 vùng biển, bảng dự báo chi tiết, banner khuyến cáo.
- Trang chủ mô tả vấn đề Bình Định 2026, trang Giới thiệu, trang 404 có hotline HTK.
- Responsive 375px + 1440px, UI tiếng Việt.

## 🛠️ Công nghệ

- Go 1.24+ / **Gin v1.11.0** (`api/index.go` — Handler cho Vercel Go runtime)
- `go:embed` nhúng frontend tĩnh (`public/`) vào binary
- HTML/CSS/JS thuần + Chart.js 4 (CDN)
- Vercel: rewrite `/(.*) → /api/index/$1`, khôi phục path trong Handler

## 🚀 Cách chạy local

```bash
export PATH="$PATH:/c/Program Files/Go/bin"   # Windows Git Bash nếu Go chưa trong PATH
go mod download
go build ./...
go vet ./...
go test ./...
PORT=8080 go run ./cmd/server
# Mở http://localhost:8080
```

API thử nhanh:

```bash
curl "http://localhost:8080/api/zones"
curl "http://localhost:8080/api/forecast?zone=cu-lao-xanh&hours=72"
```

## 🌐 Deploy

- Vercel (production): xem `deploy_url.txt` — https://binhdinh-fisherman-weather.vercel.app
- GitHub: https://github.com/shirmtry/binhdinh-fisherman-weather

## 📞 Liên hệ HTK – 0396241674 – 09/11/2026

- Tác giả: **HTK**
- Hotline/Zalo: **0396241674** — [tel:0396241674](tel:0396241674) — https://zalo.me/0396241674
- Email: [htk.contact@gmail.com](mailto:htk.contact@gmail.com)
- Cập nhật: 09/11/2026 — bối cảnh mùa mưa lũ 2026 Bình Định (bão số 9, 10/2026, lũ sông Kôn,
  sạt lở An Lão – Vĩnh Thạnh – Tây Sơn, xói lở Nhơn Lý – Cát Tiến – Mỹ Thành).

© 2026 HTK · Hotline/Zalo: 0396241674 · Email: htk.contact@gmail.com