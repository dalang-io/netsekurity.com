# HANDOVER — netsekurity.com (per 2026-10-26)

Catatan serah-terima untuk engineer / agen berikutnya. Daftar kerja yang belum selesai ada di
[`TODO.md`](TODO.md); isi README.md tidak diulang di sini. Urutan baca: `CLAUDE.md` →
file ini → `TODO.md` → `README.md` → `CLAIMS.md` → `AGENT_WORKER.md` / `HERMES_UPDATE.md`.

## Perubahan lokal sesudah rilis .2 — redesign konten dan verifikasi

Belum commit/push/deploy. Lihat [rencana dan audit implementasi](PLAN-content-redesign.md)
untuk cakupan serta bukti validasi. Homepage EN/ID sekarang mengutamakan pemilik
website, pemeriksaan berulang, contoh hasil, harga, dan pentest menyeluruh.
Panduan baru: `/getting-started` dan `/id/panduan-memulai`.

- `verification.go`: DNS TXT, file HTTPS, meta tag; TLS valid, tanpa redirect,
  pembatasan ukuran dan IP publik. `domains.go` menyediakan unduh bukti milik akun.
- `assessment_start.go`: recheck bukti, debit/ledger/antrean atomik; worker memeriksa
  lagi sebelum claim dan refund sekali bila bukti hilang. Kontrak partner terpisah.
- **Migrasi integrasi:** start API (termasuk alias `/pentests`) kini wajib memakai
  `authorised=true` atau boolean JSON `"authorised": true`. Perbarui klien sebelum
  rilis; contoh repo sudah diperbarui. Migrasi DB menambah `pentests.authorized_at`.
- `result_summary.go`: langkah berikutnya EN/ID dari counts aktual; PDF tetap Inggris.
- Dependency baru: `golang.org/x/net v0.55.0` untuk parser HTML, Go tetap 1.25.
- Suite Go, race detector, vet, build, CSS, dan diff check lulus. Audit browser:
  63 kombinasi halaman/lebar layar tanpa overflow, 11 asersi interaksi lulus.
- Scope scanner, kemampuan harness/CVE, dan status produksi tidak berubah oleh
  perubahan lokal ini. Empat dependency frontend lama dengan temuan kerentanan tercatat dari
  instalasi npm; pembaruan dependency tersebut di luar cakupan redesign.

## 1. Status

Produksi menjalankan **`v2026.10.26.2`** (commit `b6f31ab`, `main`). Deploy lewat `./deploy.sh`
(PID `1117940 → 1707740`, landing HTTP 200). Dicek langsung di produksi setelah deploy:

- 12 halaman publik (EN + ID, termasuk halaman bukti) HTTP 200; `/api/v1/assessments` tanpa token → 401.
- Header keamanan lengkap + CSP nonce di `/`; aset memakai `?v=<hash isi file>` dan `cache-control: immutable`.

| Tag | Commit | Isi |
|---|---|---|
| `v2026.10.26` | `acb075d` | Dashboard light-mode, sidebar, bahasa awam |
| `v2026.10.26.1` | `b9ed2ea` | Revamp situs publik: enterprise, trust, SEO/GEO, EN + ID |
| `v2026.10.26.2` | `b6f31ab` | Perbaikan kritik: header/CSP, mobile, halaman bukti dwibahasa, ilustrasi, API hasil, aset ber-hash |

**Yang BELUM selesai dan membuat sebagian klaim di situs belum tentu benar di agen:**
agen Hermes belum di-`git pull` (User-Agent `NetSekurity Scanner`), dan isi checkout agen belum
pernah dilihat (apakah ada perubahan yang belum di-commit). Lihat `TODO.md` T-01 dan T-02.
Sesi ini tidak bisa ssh ke host produksi (pembacaan produksi ditolak oleh classifier), jadi
langkah itu harus dijalankan manusia atau setelah izin diberikan.

## 2. Peta kode (yang berubah di rilis .2)

| Area | File | Catatan |
|---|---|---|
| Header + CSP | `security.go` | Dua kebijakan: ketat (nonce, halaman publik) dan baseline (dashboard/checkout/`/su`) |
| Aset | `assets.go` | `assetURL()` / template func `asset`: `?v=<sha256[:12]>`; immutable hanya untuk file embed yang nyata |
| Mesin halaman publik | `site.go`, `site_html.go`, `site_evidence_html.go` | Satu registry (`sitePages`), satu template per halaman, EN + ID |
| Teks | `site_copy.go` (tipe), `site_copy_en.go`, `site_copy_id.go` | `TestSiteCopyIsComplete` menjaga keduanya sama bentuk |
| Ilustrasi | `tools/svg/iso.py` → `site_svg.go` (GENERATED) | Jangan edit Go; jalankan `python3 tools/svg/iso.py` |
| PDF contoh | `tools/pdf/render.sh` → `static/netsekurity-sample-report-*.pdf` + `static/sample-report.sources` | Tes gagal bila teks contoh berubah tanpa render ulang |
| Kartu OG | `tools/og/render.sh` → `static/og*.png` | Hash isi dipakai di `?v=` |
| API CI/CD | `apitokens.go` | `POST /api/v1/assessments` (alias `/pentests`), `GET /api/v1/assessments/{id}`, 409 bila sudah ada yang berjalan |
| Bukti → siapa yang menjaga | `CLAIMS.md` | Setiap klaim publik + sumbernya; yang sengaja TIDAK diklaim |

## 3. Cara rilis

```bash
go test ./... -count=1                      # atau biarkan deploy.sh yang menjalankannya
./deploy.sh "feat: ..."                     # test → css → build linux → commit+push main → scp → stop/start → cek PID
git tag -a vYYYY.MM.DD.N -m "..." && git push origin vYYYY.MM.DD.N     # setelah deploy sukses, pada commit yang ter-deploy
```

- Tag hanya untuk rilis yang di-deploy (`vYYYY.MM.DD[.N]`). Commit dokumen saja tidak di-tag.
- **Tidak perlu purge Cloudflare** untuk CSS, favicon, kartu OG, PDF contoh, font, htmx: URL-nya berubah
  bersama isinya. Halaman HTML `no-store`. Jika `tools/og/render.sh` atau `tools/pdf/render.sh` dijalankan,
  commit hasilnya; hash di URL mengurus sisanya.
- Setelah mengubah template atau kelas apa pun: `npm run build:css` dan commit `static/css/styles.css`
  (di-embed ke binary; `deploy.sh` membangunnya lagi).
- Rollback: `netsekurity.prev` ada di direktori aplikasi; atau deploy ulang commit sebelumnya.

## 4. Menjalankan lokal (aman)

`.env` di repo berisi rahasia asli dan **menang atas environment variable** (`env.go`). Jalankan dari
direktori kosong dengan DB sekali-pakai:

```bash
go build -o /tmp/nsk-dev . && mkdir -p /tmp/nsk-run && cd /tmp/nsk-run \
 && DATABASE_PATH=/tmp/nsk-run/dev.db JWT_SECRET=dev-only PORT=8199 SERVER_PORT=8199 /tmp/nsk-dev
```

Jangan menekan **Accept** pada banner cookie di mesin yang memuat pixel Meta asli: itu mengirim event ke
pixel produksi (lihat T-08).

Menguji lebar ponsel: situs mengirim `X-Frame-Options: DENY`, jadi iframe langsung tidak bisa dipakai.
Cara yang dipakai sesi ini: proxy kecil satu-origin yang membuang header frame/CSP, halaman harness dengan
iframe selebar 320/390, lalu axe-core 4.9.1 (`npm pack axe-core@4.9.1`) disuntikkan ke iframe. Hasil rilis .2:
0 pelanggaran pada 14 halaman di 1280 dan 390 px; tanpa overflow horizontal di 320 dan 390 px.

## 5. Jebakan yang sudah dipelajari

- **Nonce ⇒ tanpa ETag.** Setiap respons halaman publik membawa nonce sendiri; ETag/304 akan memasangkan
  body lama dengan nonce baru dan memblokir semua script. Jangan tambahkan ETag atau cache bersama untuk halaman.
- **Tidak ada inline handler / `style="..."` di markup publik** (CSP ketat; `security_test.go` menegakkannya).
  Perilaku dipasang oleh satu script ber-nonce di layout. Dashboard/checkout/`/su` masih inline (T-09).
- **`rateLimit` tidak membatasi GET** (sengaja, agar polling fragmen aman). Endpoint polling API memakai
  `rateLimitAll` (300/menit per alamat klien).
- **Font dan htmx tidak bisa pakai `?v=`**: preload harus sama dengan URL `@font-face`. Mereka diberi versi
  lewat nama file dan sha256-nya dikunci tes. Ubah byte ⇒ ganti nama.
- **Tes saling memengaruhi lewat DB bersama**: tes lain meng-claim "scan antre berikutnya". Tes baru yang
  membuat baris `pentests` harus membersihkannya (`apiFixture` sudah melakukannya).
- **gopls di editor kadang menampilkan diagnostik basi** (field/fungsi "undefined") setelah file diganti
  besar-besaran. Kebenaran: `go build ./... && go vet ./... && go test ./...`.
- **Skrip CI di dokumentasi benar-benar dijalankan** oleh tes (`sh`, `curl`, `jq`; dilewati bila tidak ada).
  Tes itu menemukan bug di draf pertama (pipe menelan kegagalan `curl`).
- Dua format waktu di DB (RFC3339 dari Go, `YYYY-MM-DD HH:MM:SS` dari SQLite): `parseStoredTime` menerima keduanya.

## 6. Keputusan yang diambil (dan alasannya)

| Keputusan | Alasan |
|---|---|
| Penyedia model AI **tidak disebut** di situs; pembaca diarahkan ke security@dalang.io | Tidak ada sumber di repo; menebak lebih berbahaya daripada mengaku belum dipublikasikan (T-03) |
| Subdomain **tidak** dimasukkan ke HSTS, tanpa `preload` | Sulit ditarik kembali; naikkan nanti bila sudah yakin |
| `POST /api/v1/pentests` dipertahankan sebagai alias | Pipeline lama tidak boleh rusak; nama publik `/assessments` (kata "pentest" bertentangan dengan "bukan uji penetrasi") |
| Start saat sudah ada yang berjalan: **409** (dulu 200 + body error) | Skrip yang memakai `curl -f` harus tahu tidak ada yang dimulai; dokumentasi lama menyebut 400 |
| `findings` = `null` sampai laporan dikirim dengan rincian; tidak pernah angka nol | Nol terbaca "bersih" |
| PDF contoh tanpa harga, dibuat dari halaman | Tidak basi saat harga berubah; hanya basi bila teks/templat contoh berubah |
| Tabel lebar melipat kolom (bukan scroll) di ponsel; tabel-menjadi-kartu mempertahankan ARIA role | Kolom Buy dulu keluar layar di 390 px |
| Ilustrasi tanpa kata (hanya angka 1-5), `aria-hidden`, ide yang sama ada di teks | Kata di dalam gambar tidak bisa diterjemahkan/dibaca |
| Bahasa Inggris British; istilah Indonesia dikunci tes (Direksi, engineering lead, read-only, false positive) | Konsistensi; operator di Singapura |
| Tidak ada klaim sertifikasi/pelanggan/fasilitas/uptime | Milik fasilitas IDC atau entitas lain, bukan Netsekurity (lihat `CLAIMS.md`) |

## 7. Yang tidak bisa diverifikasi dari repo

- Isi checkout agen Hermes dan apakah ada perubahan yang belum di-commit.
- Mode SSL Cloudflare (README: *flexible*, origin HTTP :80). Beralih ke Full (strict) memerlukan TLS di origin (T-04).
- `report_auto.py` (pembuat referensi CVE, teks PoC, pemetaan OWASP) tidak ada di repo ini; klaim laporan
  bergantung pada pernyataan pemilik (T-18).
- Jalur Accept pada banner (pixel Meta) belum dijalankan di browser pada rilis ini; hanya tes unit (T-08).
- Durasi pemindaian "10 menit sampai 24 jam" adalah pernyataan pemilik, tidak diukur oleh apa pun di repo.
