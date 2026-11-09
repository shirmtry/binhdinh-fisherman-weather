// Thời tiết biển ngư dân Bình Định — HTK 0396241674 — cập nhật 09/11/2026
// Dữ liệu mô phỏng – cập nhật 09/11/2026 (gọi API Go Gin: /api/zones, /api/forecast).
"use strict";

const $ = (s) => document.querySelector(s);
const zoneSelect = $("#zone-select");
const hoursSelect = $("#hours-select");
const cardsBox = $("#zone-cards");
const currentBox = $("#current-box");
const tbody = $("#forecast-table tbody");

let chartGioSong = null;
let chartTamNhin = null;
let zonesData = [];

const NHAN_MUC = { do: "ĐỎ", vang: "VÀNG", xanh: "XANH" };
const MAU_MUC = { do: "#d32f2f", vang: "#e6a700", xanh: "#2e7d32" };

function badge(muc) {
  return `<span class="badge ${muc}">${NHAN_MUC[muc] || muc}</span>`;
}

async function taiZones() {
  try {
    const res = await fetch("/api/zones");
    const data = await res.json();
    if (!data.ok) throw new Error("API lỗi");
    zonesData = data.zones;
    zoneSelect.innerHTML = zonesData
      .map((z) => `<option value="${z.id}">${z.ten}</option>`)
      .join("");
    veCards();
    await taiForecast();
  } catch (e) {
    cardsBox.textContent = "Không tải được dữ liệu. Vui lòng thử lại hoặc gọi HTK 0396241674.";
    currentBox.textContent = "Lỗi kết nối API dự báo.";
  }
}

function veCards() {
  const dangChon = zoneSelect.value;
  cardsBox.innerHTML = zonesData
    .map((z) => {
      const ht = z.hienTai;
      return `<div class="zone-card ${z.id === dangChon ? "active" : ""}" data-zone="${z.id}" role="button" tabindex="0">
        <h3>${z.ten} ${badge(ht.mucCanhBao)}</h3>
        <p>Gió hiện tại: <b>cấp ${ht.gioCap}</b> · Sóng: <b>${ht.song.toFixed(1)} m</b></p>
        <p>72h: gió max cấp ${z.tomTat.gioMax}, sóng max ${z.tomTat.songMax.toFixed(1)} m</p>
        <p>Đỏ: ${z.tomTat.soGioDo}h · Vàng: ${z.tomTat.soGioVang}h</p>
      </div>`;
    })
    .join("");
  cardsBox.querySelectorAll(".zone-card").forEach((el) => {
    const chon = () => {
      zoneSelect.value = el.dataset.zone;
      taiForecast();
    };
    el.addEventListener("click", chon);
    el.addEventListener("keydown", (ev) => {
      if (ev.key === "Enter" || ev.key === " ") chon();
    });
  });
}

async function taiForecast() {
  const zone = zoneSelect.value || "cu-lao-xanh";
  const hours = hoursSelect.value || "72";
  veCards();
  currentBox.textContent = "Đang tải dự báo…";
  try {
    const res = await fetch(`/api/forecast?zone=${encodeURIComponent(zone)}&hours=${hours}`);
    const data = await res.json();
    if (!data.ok) throw new Error(data.error || "API lỗi");
    veHienTai(data);
    veBieuDo(data.forecast, data.zone.ten, hours);
    veBang(data.forecast);
  } catch (e) {
    currentBox.textContent = "Không tải được dự báo: " + e.message;
  }
}

function veHienTai(data) {
  const z = data.zone;
  const d = data.forecast[0];
  currentBox.className = "current-box muc-" + d.mucCanhBao;
  currentBox.innerHTML = `
    <h3>${z.ten} ${badge(d.mucCanhBao)} — ${d.tenCanhBao}</h3>
    <p>${z.moTa}</p>
    <div class="metrics">
      <div class="metric">Gió<b>cấp ${d.gioCap}</b></div>
      <div class="metric">Gió giật<b>cấp ${d.gioGiat}</b></div>
      <div class="metric">Tốc độ gió<b>${d.tocDoGio} km/h</b></div>
      <div class="metric">Sóng<b>${d.song.toFixed(1)} m</b></div>
      <div class="metric">Tầm nhìn<b>${d.tamNhin.toFixed(1)} km</b></div>
    </div>
    <p><strong>Khuyến cáo:</strong> ${d.khuyenCao}</p>
    <p class="mock-note">Dữ liệu mô phỏng – cập nhật 09/11/2026 · Dự báo ${data.hours} giờ, bước ${data.buocGio} giờ.</p>`;
}

function veBieuDo(ds, tenVung, hours) {
  const labels = ds.map((d) => d.nhanThoiGian);
  if (chartGioSong) chartGioSong.destroy();
  if (chartTamNhin) chartTamNhin.destroy();

  chartGioSong = new Chart($("#chart-gio-song"), {
    type: "line",
    data: {
      labels,
      datasets: [
        {
          label: "Gió (cấp Beaufort)",
          data: ds.map((d) => d.gioCap),
          borderColor: "#13558c",
          backgroundColor: "rgba(19,85,140,0.15)",
          yAxisID: "yGio",
          tension: 0.3,
          fill: true,
        },
        {
          label: "Sóng (m)",
          data: ds.map((d) => d.song),
          borderColor: "#d32f2f",
          backgroundColor: "rgba(211,47,47,0.10)",
          yAxisID: "ySong",
          tension: 0.3,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        title: { display: true, text: `Gió & sóng ${hours}h — ${tenVung} (mô phỏng 09/11/2026)` },
      },
      scales: {
        yGio: { position: "left", min: 0, max: 10, title: { display: true, text: "Cấp Beaufort" } },
        ySong: { position: "right", min: 0, max: 6, grid: { drawOnChartArea: false }, title: { display: true, text: "Sóng (m)" } },
      },
    },
  });

  chartTamNhin = new Chart($("#chart-tam-nhin"), {
    type: "line",
    data: {
      labels,
      datasets: [
        {
          label: "Tầm nhìn (km)",
          data: ds.map((d) => d.tamNhin),
          borderColor: "#2e7d32",
          backgroundColor: "rgba(46,125,50,0.12)",
          tension: 0.3,
          fill: true,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        title: { display: true, text: `Tầm nhìn ${hours}h — ${tenVung} (mô phỏng 09/11/2026)` },
      },
      scales: { y: { min: 0, max: 13, title: { display: true, text: "km" } } },
    },
  });
}

function veBang(ds) {
  tbody.innerHTML = ds
    .map(
      (d) => `<tr class="row-${d.mucCanhBao}">
        <td>${d.nhanThoiGian}</td>
        <td>${d.gioCap}</td>
        <td>${d.gioGiat}</td>
        <td>${d.tocDoGio}</td>
        <td>${d.song.toFixed(1)}</td>
        <td>${d.tamNhin.toFixed(1)}</td>
        <td>${badge(d.mucCanhBao)}</td>
      </tr>`
    )
    .join("");
}

zoneSelect.addEventListener("change", taiForecast);
hoursSelect.addEventListener("change", taiForecast);
taiZones();