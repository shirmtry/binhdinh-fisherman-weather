// Smoke test httptest cho các route chính.
// HTK – 0396241674 – cập nhật 09/11/2026
package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func goi(method, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, target, nil)
	Handler(w, r)
	return w
}

func TestTrangChinh(t *testing.T) {
	for _, tc := range []struct {
		path   string
		status int
		chua   string
	}{
		{"/", 200, "0396241674"},
		{"/about", 200, "Cập nhật: 09/11/2026"},
		{"/style.css", 200, "footer"},
		{"/app.js", 200, "api/zones"},
		{"/khong-ton-tai", 404, "0396241674"},
	} {
		w := goi("GET", tc.path)
		if w.Code != tc.status {
			t.Fatalf("%s: muốn %d, được %d", tc.path, tc.status, w.Code)
		}
		if !strings.Contains(w.Body.String(), tc.chua) {
			t.Fatalf("%s: thiếu chuỗi %q", tc.path, tc.chua)
		}
	}
}

func TestAPIZones(t *testing.T) {
	w := goi("GET", "/api/zones")
	if w.Code != 200 {
		t.Fatalf("muốn 200, được %d", w.Code)
	}
	var kq struct {
		OK    bool   `json:"ok"`
		Note  string `json:"note"`
		Total int    `json:"total"`
		Zones []struct {
			ID      string    `json:"id"`
			HienTai DiemDuBao `json:"hienTai"`
		} `json:"zones"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &kq); err != nil {
		t.Fatal(err)
	}
	if !kq.OK || kq.Total != 6 || len(kq.Zones) != 6 {
		t.Fatalf("zones không hợp lệ: %+v", kq)
	}
	if kq.Note != dataNote {
		t.Fatalf("thiếu nhãn mock: %q", kq.Note)
	}
	for _, z := range kq.Zones {
		if z.ID == "" || z.HienTai.GioCap < 2 || z.HienTai.Song <= 0 {
			t.Fatalf("zone lỗi: %+v", z)
		}
	}
}

func TestAPIForecast(t *testing.T) {
	w := goi("GET", "/api/forecast?zone=cu-lao-xanh")
	if w.Code != 200 {
		t.Fatalf("muốn 200, được %d", w.Code)
	}
	var kq struct {
		OK       bool        `json:"ok"`
		Hours    int         `json:"hours"`
		Forecast []DiemDuBao `json:"forecast"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &kq); err != nil {
		t.Fatal(err)
	}
	if !kq.OK || kq.Hours != 72 || len(kq.Forecast) != 25 {
		t.Fatalf("forecast không hợp lệ: hours=%d len=%d", kq.Hours, len(kq.Forecast))
	}
	coDo := false
	for _, d := range kq.Forecast {
		if d.MucCanhBao == "do" {
			coDo = true
		}
		if d.MucCanhBao != "do" && d.MucCanhBao != "vang" && d.MucCanhBao != "xanh" {
			t.Fatalf("mức cảnh báo lạ: %q", d.MucCanhBao)
		}
	}
	if !coDo {
		t.Fatal("Cù Lao Xanh phải có cảnh báo đỏ trong đợt bão mô phỏng")
	}

	if w := goi("GET", "/api/forecast?zone=vinh-quy-nhon&hours=24"); w.Code != 200 {
		t.Fatalf("hours=24 lỗi: %d", w.Code)
	}
	if w := goi("GET", "/api/forecast?zone=sai"); w.Code != 400 {
		t.Fatalf("zone sai phải trả 400, được %d", w.Code)
	}
	if w := goi("GET", "/api/forecast"); w.Code != 400 {
		t.Fatalf("thiếu zone phải trả 400, được %d", w.Code)
	}
}

func TestRewriteVercel(t *testing.T) {
	w := goi("GET", "/api/index/api/zones")
	if w.Code != 200 {
		t.Fatalf("rewrite /api/index/api/zones: muốn 200, được %d", w.Code)
	}
	w = goi("GET", "/api/index")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "0396241674") {
		t.Fatalf("rewrite /api/index: %d", w.Code)
	}
}