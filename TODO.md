# TODO — netsekurity.com (per 2026-10-26)

Prioritas: **P0** = membuat klaim di situs benar / ada risiko; **P1** = minggu ini; **P2** = jadwalkan;
**P3** = rapi-rapi. Pemilik: *Ops* = yang punya akses ke host; *Owner* = butuh keputusan/fakta dari pemilik;
*Dev* = bisa dikerjakan di repo. Konteks lengkap: [`HANDOVER.md`](HANDOVER.md).

## P0

- [ ] **T-01 · Ops · Tarik rilis ke agen Hermes.** Situs menyatakan User-Agent `NetSekurity Scanner`
  (halaman trust + enterprise); itu baru benar untuk pelanggan setelah agen di-`git pull`.
  ```bash
  cd <checkout netsekurity.com milik agen>        # dugaan: /home/dalang/netsekurity.com di 163.128.54.5; belum dikonfirmasi
  git status --short; git log --oneline -3; git remote -v | head -1
  git pull --rebase origin main
  grep -c '^UA = "NetSekurity Scanner"' nsec_scan.py     # harus 1
  grep -ci 'mozilla\|chrome/' nsec_scan.py               # harus 0
  ```
  Verifikasi nyata: pindai domain milik sendiri, cari `NetSekurity Scanner` di access log domain itu.
  *Selesai bila:* kedua `grep` benar dan access log menunjukkan UA baru.
- [ ] **T-02 · Ops → Dev · Pastikan agen meng-commit pekerjaannya ke repo.** Lihat output `git status --short`
  dari T-01. Jika ada perubahan yang belum di-commit (tambalan panas):
  1. `git diff` dan baca; hapus yang tidak sengaja; jangan commit rahasia (`.env`, kunci, token).
  2. `git add <file>` (bukan `-A`), commit atas identitas agen, `git pull --rebase origin main`, `git push origin main`.
  3. Bila yang berubah file platform (`*.go`, templat), kembali ke laptop: `./deploy.sh "..."`; bila hanya `nsec_*`
     atau dokumen, tidak perlu deploy.
  4. Beri tag rilis berikutnya pada commit yang sudah digabung: `v2026.10.26.3` (atau `.N` bebas berikutnya),
     `git push origin <tag>`.
  Aturan ke depan: **tidak ada tambalan panas di host**; semua lewat `main`. Cabang `wip/netsekurity-live-state-20261016`
  adalah contoh akibat tambalan yang tidak di-commit.
- [ ] **T-03 · Owner → Dev · Penyedia model AI dan wilayahnya.** Situs mengatakan "AI analysis agent" dan belum
  menyebut penyedia (FAQ "Does Netsekurity use AI…", baris "AI model provider" di `/trust#third-parties`, baris
  "AI analysis" di tabel data). Pembeli enterprise akan menanyakannya. Saat fakta ada: ubah `trustEN/ID`
  (`Subs`, `Data`), dua FAQ, `CLAIMS.md`; hapus "tanyakan ke security@dalang.io" bila tidak lagi benar; deploy + tag.
  Jangan menulis nama penyedia sebelum pemilik memastikannya (`TestTrustPageListsThirdPartiesAndTheAIQuestion`
  sengaja melarang nama penyedia).

## P1

- [ ] **T-04 · Ops · Cloudflare SSL: dari *flexible* ke Full (strict).** README menyebut *flexible* (origin HTTP :80), belum
  diverifikasi. Full (strict) butuh TLS di origin: sertifikat Cloudflare Origin CA di Pingora `:443`, route baru untuk
  `netsekurity.com`, lalu ubah mode di dashboard Cloudflare. HSTS (`max-age=31536000`) sudah dikirim, jadi **jangan**
  mematikan HTTPS di tengah jalan. Uji dengan `curl -sI https://netsekurity.com/` dan cek tidak ada redirect loop.
- [ ] **T-05 · Owner · Cloudflare "Managed robots.txt" untuk crawler AI.** Cloudflare menambahkan aturan yang memblokir
  crawler pelatihan (GPTBot, ClaudeBot, CCBot, Google-Extended, Applebot-Extended, …) di depan `robots.txt` kita, bertentangan
  dengan aturan *allow* dan tujuan GEO. Putuskan: matikan fitur itu di dashboard Cloudflare (bagian Bots / AI crawlers), atau
  terima dan tulis ulang `static/robots.txt` + `llms.txt` agar jujur. Cek: `curl -s https://netsekurity.com/robots.txt | head -40`.
- [ ] **T-06 · Ops · Cloudflare Web Analytics (RUM)** disuntikkan CDN dan menyala sebelum persetujuan cookie (README,
  "Residual"). Matikan di dashboard bila kebijakan "tidak ada pelacak sebelum setuju" harus berlaku penuh.
- [ ] **T-07 · Owner · Tinjauan penutur asli untuk bahasa Indonesia**: `site_copy_id.go` (seluruhnya),
  `static/PRODUCT_SCOPE.id.md`, string banner di `consent.go`. Ditulis oleh model; istilah dikunci tes tapi nuansa belum
  ditinjau manusia. Perubahan ⇒ sentuh hanya string; tes menjaga bentuk.
- [ ] **T-08 · Dev · Uji jalur *Accept* banner di staging** dengan ID pixel uji (bukan pixel produksi): pixel dimuat dengan
  nonce yang benar, tidak ada pelanggaran CSP, `ViewContent` terkirim via `nskTrack`. Sesi ini sengaja tidak menekan
  Accept di mesin lokal agar tidak mengirim event ke pixel asli; hanya tes unit yang mencakup jalur ini.

## P2

- [ ] **T-09 · Dev · Pindahkan dashboard, checkout, `/su` ke CSP bernonce.** Mereka masih memakai `onclick="…"` dan memakai
  kebijakan baseline (`script-src 'unsafe-inline'`). Ganti handler dengan listener delegasi, lalu pakai `strictCSP`
  (petunjuk di komentar atas `security.go`); perbarui `TestAppPagesGetTheBaselinePolicy` bila namanya berubah.
- [ ] **T-10 · Dev · Penjaga PDF contoh hanya membaca teks.** Ubah templat `siteSampleHTML` atau aturan cetak di
  `assets/input.css` ⇒ jalankan `tools/pdf/render.sh` sendiri dan commit PDF + `static/sample-report.sources`.
  Opsional: masukkan hash templat/CSS cetak ke `sampleSources()`.
- [ ] **T-11 · Dev · API hasil, tahap berikutnya:** `GET /api/v1/assessments` (daftar), unduh PDF memakai token, webhook selesai.
  Saat ini hanya status + jumlah temuan; PDF tetap di dashboard. Setiap perubahan: perbarui `docsEN/ID`, `apiExample`,
  `CLAIMS.md` dan `api_v1_test.go`.
- [ ] **T-12 · Owner · Celah yang akan ditanyakan kuesioner vendor** (lihat tabel "tidak diklaim" di `CLAIMS.md`):
  periode retensi/penghapusan data, DPA, MSA, SLA/uptime khusus Netsekurity, perlakuan PPN/faktur pajak, alamat IP sumber
  pemindai. Tiap jawaban ⇒ sumber ⇒ `CLAIMS.md` ⇒ halaman enterprise/trust (EN + ID).
- [ ] **T-13 · Owner · Bahasa Mandarin / Melayu?** Rekomendasi: belum. Putuskan dari data Umami (pengunjung dengan
  `Accept-Language` zh/ms dan konversi), lalu tambahkan satu entri `siteLangs` + satu set salinan; tes memaksa kelengkapan.
- [ ] **T-18 · Dev/Owner · `report_auto.py` tidak ada di repo.** Referensi CVE, teks PoC, dan pemetaan OWASP dihasilkan di
  sana; klaim laporan bersandar pada pernyataan pemilik. Masukkan ke repo (atau dokumentasikan lokasi + versinya) agar
  klaim itu bisa diverifikasi dan diaudit.

## P3

- [ ] **T-14 · Dev · Isi ulang jumlah temuan untuk scan lama** (kartu masih "rincian ada di laporan"): simpan `assessment_id`
  pada baris `pentests` saat unggah (catatan di `HERMES_UPDATE.md`, "Old scans cannot be backfilled").
- [ ] **T-15 · Dev · `npx update-browserslist-db@latest`** (peringatan caniuse-lite basi saat `npm run build:css`).
- [ ] **T-16 · Ops · `static/security.txt` kedaluwarsa `2027-08-19`**: perbarui sebelum tanggal itu
  (`Expires:` + commit + deploy). Pasang pengingat.
- [ ] **T-17 · Dev · Tes lintas-browser** untuk `print` (fact sheet pengadaan) di Safari/Firefox; PDF contoh hanya dibuat
  dari Chrome headless.

## Selesai di rilis .2 (referensi)

Header + CSP bernonce; font/htmx self-host; aset ber-hash isi (tanpa purge); halaman bukti dwibahasa; ilustrasi isometrik;
tabel mobile; harga konsisten; istilah ID dikunci; API `assessments` + hasil; skrip gerbang CI teruji; cetak + PDF contoh;
daftar pihak ketiga + alur data di `/trust`; `CLAIMS.md` dan `README.md` diperbarui.
