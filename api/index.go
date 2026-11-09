// Package handler phục vụ API + frontend Thời tiết biển ngư dân Bình Định.
// Dự án: binhdinh-fisherman-weather — ngư dân ra khơi tháng 11/2026 gặp bão bất ngờ,
// biển động cấp 7–8; cần dự báo gió (Beaufort), sóng (m), tầm nhìn và khuyến cáo ra khơi.
// Tác giả: HTK — Hotline/Zalo: 0396241674 — Email: htk.contact@gmail.com
// Cập nhật: 09/11/2026. Dữ liệu mô phỏng – cập nhật 09/11/2026.
package handler

import (
	"fmt"
	"io/fs"
	"math"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	app "binhdinh-fisherman-weather"

	"github.com/gin-gonic/gin"
)

const dataNote = "Dữ liệu mô phỏng – cập nhật 09/11/2026"

// mocThoiGian: 00:00 ngày 09/11/2026 (giờ Việt Nam, UTC+7) — cố định để mock nhất quán.
var mocThoiGian = time.Date(2026, time.November, 9, 0, 0, 0, 0, time.FixedZone("ICT", 7*3600))

// DiemDuBao là một mốc dự báo 3 giờ cho vùng biển.
type DiemDuBao struct {
	GioThu      int     `json:"gioThu"`      // giờ thứ mấy kể từ 09/11/2026 00:00
	ThoiGian    string  `json:"thoiGian"`    // ISO 8601
	NhanThoiGian string `json:"nhanThoiGian"` // nhãn ngắn cho biểu đồ, vd "09/11 06h"
	GioCap      int     `json:"gioCap"`      // cấp gió Beaufort (2–8)
	TocDoGio    int     `json:"tocDoGio"`    // km/h trung bình
	GioGiat     int     `json:"gioGiat"`     // cấp gió giật
	Song        float64 `json:"song"`        // chiều cao sóng (m)
	TamNhin     float64 `json:"tamNhin"`     // tầm nhìn (km)
	MucCanhBao  string  `json:"mucCanhBao"`  // "do" | "vang" | "xanh"
	TenCanhBao  string  `json:"tenCanhBao"`  // nhãn tiếng Việt
	KhuyenCao   string  `json:"khuyenCao"`   // khuyến cáo ra khơi
}

// ThamSoMock điều khiển đường cong bão mô phỏng cho từng vùng biển.
type ThamSoMock struct {
	GioNen   float64 // gió nền (cấp Beaufort)
	BienDo   float64 // biên độ tăng khi bão số 10/2026 áp sát
	DinhGio  float64 // giờ bão ảnh hưởng mạnh nhất
	DoRong   float64 // độ rộng vùng ảnh hưởng (giờ)
	Pha      float64 // pha dao động nhỏ theo thời điểm trong ngày
}

// VungBien là một trong 6 vùng biển Bình Định.
type VungBien struct {
	ID      string     `json:"id"`
	Ten     string     `json:"ten"`
	MoTa    string     `json:"moTa"`
	HienTai DiemDuBao  `json:"hienTai"`
	TomTat  TomTatVung `json:"tomTat"`
	thamSo  ThamSoMock
}

// TomTatVung tổng kết đợt dự báo 72 giờ.
type TomTatVung struct {
	GioMax     int     `json:"gioMax"`     // cấp gió lớn nhất 72h
	SongMax    float64 `json:"songMax"`    // sóng lớn nhất 72h (m)
	GioDinh    string  `json:"gioDinh"`    // thời điểm gió mạnh nhất
	SoGioDo    int     `json:"soGioDo"`    // số giờ cảnh báo đỏ
	SoGioVang  int     `json:"soGioVang"`  // số giờ cảnh báo vàng
	MucXauNhat string  `json:"mucXauNhat"` // mức cảnh báo xấu nhất 72h
}

// danhSachVung: 6 vùng biển Bình Định theo đặc tả dự án.
var danhSachVung = []VungBien{
	{ID: "nhon-ly-nhon-hai", Ten: "Nhơn Lý – Nhơn Hải", MoTa: "Ngư trường ven bờ phía bắc bán đảo Phương Mai, làng chài tôm hùm Nhơn Hải; bờ biển xói lở sau bão 2026.",
		thamSo: ThamSoMock{GioNen: 4.2, BienDo: 3.4, DinhGio: 30, DoRong: 14, Pha: 0.4}},
	{ID: "cu-lao-xanh", Ten: "Cù Lao Xanh", MoTa: "Đảo tiền tiêu xã Nhơn Châu, vùng khơi trống trải — hứng gió bão số 10/2026 mạnh nhất, tàu nhỏ tuyệt đối tránh.",
		thamSo: ThamSoMock{GioNen: 4.8, BienDo: 3.2, DinhGio: 27, DoRong: 13, Pha: 1.1}},
	{ID: "de-gi-cat-tien", Ten: "Đề Gi – Cát Tiến", MoTa: "Cửa biển Đề Gi (Phù Cát – Phù Mỹ) và dải bờ Cát Tiến; luồng tàu cá đông, sóng vỡ nguy hiểm khi biển động cấp 7–8.",
		thamSo: ThamSoMock{GioNen: 4.5, BienDo: 3.3, DinhGio: 33, DoRong: 15, Pha: 2.0}},
	{ID: "my-thanh-nhon-thanh", Ten: "Mỹ Thành – Nhơn Thành", MoTa: "Dải bờ Mỹ Thành (Phù Cát) – Nhơn Thành, bãi ngang tàu công suất nhỏ; xói lở bờ biển nghiêm trọng mùa mưa lũ 2026.",
		thamSo: ThamSoMock{GioNen: 4.0, BienDo: 2.8, DinhGio: 30, DoRong: 14, Pha: 2.8}},
	{ID: "hoai-nhon-tam-quan", Ten: "Hoài Nhơn – Tam Quan", MoTa: "Ngư trường cửa biển Tam Quan – Hoài Hải phía bắc tỉnh, đội tàu xa bờ lớn; hướng bão đổ bộ Hoài Nhơn.",
		thamSo: ThamSoMock{GioNen: 4.6, BienDo: 3.4, DinhGio: 24, DoRong: 12, Pha: 3.5}},
	{ID: "vinh-quy-nhon", Ten: "Vịnh Quy Nhơn", MoTa: "Vịnh kín từ cảng Quy Nhơn tới Thị Nại, được che chắn một phần; nơi trú tránh cho tàu cá khi biển động.",
		thamSo: ThamSoMock{GioNen: 3.2, BienDo: 2.2, DinhGio: 36, DoRong: 16, Pha: 4.2}},
}

func lamTron1(x float64) float64 { return math.Round(x*10) / 10 }

// phanLoai cảnh báo đỏ/vàng/xanh theo gió Beaufort + sóng.
func phanLoai(gioCap int, song float64) (muc, ten, khuyenCao string) {
	switch {
	case gioCap >= 7 || song >= 4.0:
		return "do", "Cảnh báo ĐỎ – biển động cấp 7–8",
			"KHÔNG ra khơi. Tàu đang hoạt động khẩn trương về bờ hoặc vào nơi trú ẩn; tuyệt đối không vượt cửa biển khi sóng lớn."
	case gioCap >= 5 || song >= 2.5:
		return "vang", "Cảnh báo VÀNG – gió mạnh, sóng cao",
			"Hạn chế ra khơi, chỉ hoạt động gần bờ với tàu công suất lớn; áo phao, bộ đàm đầy đủ, theo dõi bão số 10/2026 liên tục."
	default:
		return "xanh", "Cảnh báo XANH – tương đối an toàn",
			"Có thể ra khơi gần bờ nhưng phải nghe dự báo trước mỗi chuyến; biển chuyển xấu nhanh trong mùa bão tháng 11/2026."
	}
}

// duBaoVung sinh chuỗi dự báo 72 giờ, bước 3 giờ (25 mốc) — hoàn toàn mô phỏng, tất định.
func duBaoVung(v VungBien) []DiemDuBao {
	ts := v.thamSo
	ds := make([]DiemDuBao, 0, 25)
	for h := 0; h <= 72; h += 3 {
		gio := ts.GioNen + ts.BienDo*math.Exp(-math.Pow(float64(h)-ts.DinhGio, 2)/(2*ts.DoRong*ts.DoRong)) + 0.55*math.Sin(float64(h)/5+ts.Pha)
		gio = math.Max(2, math.Min(8, gio))
		gioCap := int(math.Round(gio))
		gioGiat := gioCap + 1
		if gioGiat > 10 {
			gioGiat = 10
		}
		song := lamTron1(math.Max(0.3, 0.25+gio*0.52+0.25*math.Sin(float64(h)/7+ts.Pha)))
		tamNhin := lamTron1(math.Max(0.8, math.Min(12, 12.5-gio*1.25+0.6*math.Cos(float64(h)/6+ts.Pha))))
		tocDo := int(math.Round(3.6 * math.Pow(float64(gioCap)+1, 1.45)))
		muc, ten, kc := phanLoai(gioCap, song)
		t := mocThoiGian.Add(time.Duration(h) * time.Hour)
		ds = append(ds, DiemDuBao{
			GioThu: h, ThoiGian: t.Format(time.RFC3339),
			NhanThoiGian: fmt.Sprintf("%02d/%02d %02dh", t.Day(), int(t.Month()), t.Hour()),
			GioCap: gioCap, TocDoGio: tocDo, GioGiat: gioGiat, Song: song, TamNhin: tamNhin,
			MucCanhBao: muc, TenCanhBao: ten, KhuyenCao: kc,
		})
	}
	return ds
}

func tomTat(ds []DiemDuBao) TomTatVung {
	tt := TomTatVung{MucXauNhat: "xanh"}
	for _, d := range ds {
		if d.GioCap > tt.GioMax {
			tt.GioMax = d.GioCap
			tt.GioDinh = d.ThoiGian
		}
		if d.Song > tt.SongMax {
			tt.SongMax = d.Song
		}
		switch d.MucCanhBao {
		case "do":
			tt.SoGioDo += 3
			tt.MucXauNhat = "do"
		case "vang":
			tt.SoGioVang += 3
			if tt.MucXauNhat != "do" {
				tt.MucXauNhat = "vang"
			}
		}
	}
	return tt
}

func timVung(id string) (VungBien, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, v := range danhSachVung {
		if v.ID == id {
			return v, true
		}
	}
	return VungBien{}, false
}

var (
	once     sync.Once
	ginApp   *gin.Engine
	staticFS fs.FS
)

func buildApp() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	staticFS, _ = fs.Sub(app.PublicFS, "public")

	serveFile := func(c *gin.Context, name string, status int) {
		b, err := fs.ReadFile(staticFS, name)
		if err != nil {
			c.String(500, "Không đọc được tài nguyên.")
			return
		}
		ct := "application/octet-stream"
		switch path.Ext(name) {
		case ".html":
			ct = "text/html; charset=utf-8"
		case ".css":
			ct = "text/css; charset=utf-8"
		case ".js":
			ct = "application/javascript; charset=utf-8"
		}
		c.Data(status, ct, b)
	}

	lienHe := gin.H{"author": "HTK", "hotline": "0396241674", "email": "htk.contact@gmail.com", "zalo": "https://zalo.me/0396241674"}

	api := r.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "project": "binhdinh-fisherman-weather", "note": dataNote, "contact": lienHe})
	})
	api.GET("/zones", func(c *gin.Context) {
		ds := make([]VungBien, 0, len(danhSachVung))
		for _, v := range danhSachVung {
			chuoi := duBaoVung(v)
			v.HienTai = chuoi[0]
			v.TomTat = tomTat(chuoi)
			ds = append(ds, v)
		}
		c.JSON(200, gin.H{
			"ok": true, "note": dataNote, "capNhat": "09/11/2026",
			"boiCanh": "Bão số 10/2026 áp sát Bình Định, biển động cấp 7–8; mùa mưa lũ tháng 11/2026.",
			"total": len(ds), "zones": ds, "contact": lienHe,
		})
	})
	api.GET("/forecast", func(c *gin.Context) {
		zoneID := c.Query("zone")
		v, ok := timVung(zoneID)
		if !ok {
			ids := make([]string, 0, len(danhSachVung))
			for _, z := range danhSachVung {
				ids = append(ids, z.ID)
			}
			c.JSON(400, gin.H{"ok": false, "error": "Thiếu hoặc sai tham số zone. Truyền ?zone=<id>.", "hopLe": ids, "hotline": "0396241674"})
			return
		}
		hours := 72
		if vq, err := strconv.Atoi(c.Query("hours")); err == nil {
			switch vq {
			case 24, 48, 72:
				hours = vq
			}
		}
		chuoi := duBaoVung(v)
		gioiHan := make([]DiemDuBao, 0, len(chuoi))
		for _, d := range chuoi {
			if d.GioThu <= hours {
				gioiHan = append(gioiHan, d)
			}
		}
		v.HienTai = chuoi[0]
		v.TomTat = tomTat(gioiHan)
		c.JSON(200, gin.H{
			"ok": true, "note": dataNote, "capNhat": "09/11/2026",
			"mocThoiGian": mocThoiGian.Format(time.RFC3339), "buocGio": 3, "hours": hours,
			"zone": v, "forecast": gioiHan, "contact": lienHe,
		})
	})

	r.GET("/", func(c *gin.Context) { serveFile(c, "index.html", 200) })
	r.GET("/about", func(c *gin.Context) { serveFile(c, "about.html", 200) })
	r.GET("/404", func(c *gin.Context) { serveFile(c, "404.html", 404) })
	r.GET("/style.css", func(c *gin.Context) { serveFile(c, "style.css", 200) })
	r.GET("/app.js", func(c *gin.Context) { serveFile(c, "app.js", 200) })

	r.NoRoute(func(c *gin.Context) {
		p := strings.TrimPrefix(c.Request.URL.Path, "/")
		if strings.HasPrefix(p, "api/") {
			c.JSON(404, gin.H{"ok": false, "error": "Không tìm thấy API.", "hotline": "0396241674"})
			return
		}
		serveFile(c, "404.html", 404)
	})
	return r
}

// restorePath khôi phục đường dẫn gốc khi Vercel rewrite qua /api/index/...
func restorePath(r *http.Request) {
	p := r.URL.Path
	if orig := r.Header.Get("X-Original-Path"); orig != "" {
		p = orig
	} else {
		for _, prefix := range []string{"/api/index.go", "/api/index"} {
			if p == prefix {
				p = "/"
				break
			}
			if strings.HasPrefix(p, prefix+"/") {
				p = strings.TrimPrefix(p, prefix)
				break
			}
		}
	}
	r.URL.Path = p
	if q := r.URL.RawQuery; q != "" {
		r.RequestURI = p + "?" + q
	} else {
		r.RequestURI = p
	}
}

// Handler là entrypoint cho Vercel Go runtime (api/index.go).
func Handler(w http.ResponseWriter, r *http.Request) {
	restorePath(r)
	once.Do(func() { ginApp = buildApp() })
	ginApp.ServeHTTP(w, r)
}