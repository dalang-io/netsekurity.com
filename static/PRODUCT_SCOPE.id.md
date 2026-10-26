# Cakupan Produk — Asesmen Keamanan Web Eksternal Netsekurity

Cakupan untuk klien dari **asesmen keamanan web eksternal otomatis** yang diberikan per
domain terverifikasi / kredit. 1 kredit = 1 asesmen eksternal · 1 domain.

> **Dinyatakan terus terang:** ini adalah **asesmen otomatis, eksternal, blackbox, tanpa autentikasi,
> dan read-only** dengan temuan yang dipilah oleh agen — **bukan uji penetrasi penuh**. Asesmen ini tidak
> menguji logika aplikasi yang terautentikasi, otorisasi (IDOR), atau logika bisnis. Untuk itu,
> lihat tingkat **Whitebox** di bawah.

---

## Persiapan dan pemeriksaan ulang

Konfirmasi kendali hostname melalui DNS TXT, file verifikasi HTTPS, atau meta tag
halaman utama. Biarkan bukti terpasang: bukti diperiksa sebelum setiap asesmen
mandiri. Konfirmasi izin dan cakupan secara terpisah sebelum mulai. Verifikasi
yang gagal tidak memakai kredit. Setiap asesmen baru atau ulang membutuhkan
1 kredit per hostname. Subdomain lain memerlukan verifikasi dan kredit sendiri.
Lihat [panduan memulai](/id/panduan-memulai).

## Metodologi (blackbox, eksternal)

1. **Pengintaian dan pemetaan permukaan serangan** — resolusi DNS dan daftar subdomain pasif dari
   certificate transparency dan sumber publik. Subdomain didaftar sebagai informasi Anda;
   hanya domain yang Anda verifikasi yang dinilai.
2. **Sidik jari web** — header server, pengungkapan versi server, dan deteksi web application
   firewall, dibaca dari respons HTTP/S.
3. **Header keamanan** — set OWASP lengkap: HSTS, X-Content-Type-Options, X-Frame-Options,
   Content-Security-Policy, Referrer-Policy, Permissions-Policy; pemeriksaan clickjacking.
4. **TLS / PKI** — masa berlaku sertifikat, jendela kedaluwarsa, penerbit, cakupan SAN.
5. **Pengungkapan informasi** — 13 probe: `.git`, `.env`, cadangan SQL/zip yang terbuka,
   cadangan `wp-config`, `server-status`, `phpinfo`, metadata SVN, path admin/login, `robots.txt`.
6. **Open redirect** — 9 parameter redirect yang umum (`redirect`, `next`, `url`, `return`, …)
   diuji apakah menghasilkan 302 ke alamat eksternal.
7. **Pemilahan oleh agen** — agen analisis AI menghapus duplikat temuan, mengurutkannya menurut
   tingkat keparahan dan dampak bisnis, lalu menulis laporan. Tidak ada manusia yang membaca laporan
   standar sebelum dikirim; pengujian ulang oleh engineer keamanan adalah **tambahan tinjauan ahli**
   yang terpisah (di bawah).

Temuan **dipetakan ke kategori OWASP Top 10** dan diberi skor **CVSS v3.1**, masing-masing dengan
pemetaan CWE, langkah reproduksi, dan panduan perbaikan.

---

## Dalam Cakupan (termasuk di setiap pemindaian)

| Area | Pemeriksaan yang dilakukan |
|------|----------------------------|
| **Pengintaian dan pemetaan permukaan serangan** | Resolusi DNS, daftar subdomain pasif (subdomain didaftar, tidak dinilai) |
| **Sidik jari aplikasi web** | Header server, pengungkapan versi server, deteksi WAF |
| **Header keamanan** | HSTS, X-Content-Type-Options, X-Frame-Options, CSP, Referrer-Policy, Permissions-Policy; clickjacking |
| **TLS / PKI** | Masa berlaku sertifikat, kedaluwarsa, penerbit, cakupan SAN |
| **Pengungkapan informasi** | `.git`, `.env`, cadangan, server-status, phpinfo, robots.txt, metadata SVN, path admin/login |
| **Open redirect** | 9 parameter redirect yang umum diuji apakah menghasilkan 302 ke alamat eksternal |

## Catatan cakupan (agar tepat)

- **"Dipetakan ke OWASP"** berarti temuan *dikategorikan* menurut OWASP Top 10. Itu **bukan** klaim
  bahwa setiap kategori OWASP diuji tuntas — sebagian (misalnya broken access control)
  memerlukan autentikasi dan beberapa identitas yang tidak terjangkau pemindaian eksternal.
- **Tidak diklaim:** SSRF, GraphQL, WebSocket, request smuggling, web cache poisoning, dan kelas lain
  yang tidak tercantum di atas. Ketiadaannya di laporan berarti "tidak dipindai", bukan "terbukti aman".
- **"Read-only"** aman untuk produksi: setiap permintaan adalah GET biasa; tidak ada yang ditulis,
  diunggah, atau dieksploitasi.

## Di Luar Cakupan (tidak termasuk — dinyatakan dengan jelas)

1. **Pengujian terautentikasi.** Pemindaian tanpa autentikasi; logika di balik login/sesi tidak diuji.
2. **Otorisasi / IDOR / eskalasi hak akses.** Memerlukan beberapa identitas; tidak diuji.
3. **Pengujian payload injeksi.** Pemindaian standar tidak mengirim payload SQLi, XSS, LFI/RFI atau
   path traversal. Pemindaian menemukan paparan dan salah konfigurasi, bukan bug input yang bisa dieksploitasi.
4. **Pengujian logika bisnis, alur kerja, pembayaran, atau race condition.** Tidak termasuk.
5. **Pemindaian port jaringan.** Asesmen standar tidak memindai port.
6. **Rantai eksploit yang dalam / manual.** Kami mendeteksi kerentanan; kami tidak menyusun rantai eksploit bertahap.
7. **Analisis kode sumber.** Hanya blackbox.
8. **Tindakan destruktif.** Semua pengujian read-only dan tidak destruktif.
9. **Pekerjaan manual yang diperpanjang.** Bukan pekerjaan berhari-hari ala APT.

## Deliverable

- **Laporan (PDF, bahasa Inggris)** per domain, dapat diunduh dari dasbor:
  - Ringkasan eksekutif dan skor postur keamanan keseluruhan
  - Register risiko (kemungkinan × dampak)
  - Temuan rinci: tingkat keparahan, CVSS v3.1, pemetaan CWE, bukti, langkah reproduksi, panduan perbaikan
  - Referensi CWE / CVE dan bukti konsep (PoC) untuk setiap temuan
- Peta jalan perbaikan (30/60/90) per temuan.

## Whitebox (keyakinan lebih dalam — layanan terpisah)

Untuk pengujian terautentikasi, analisis otorisasi/logika bisnis, dan validasi jalur eksploit:

| Tingkat | Yang termasuk | Harga |
|---------|---------------|-------|
| **Blackbox (otomatis)** | Asesmen eksternal tanpa autentikasi — produk ini | **Termasuk dalam kredit** (harga ada di halaman harga) |
| **Tinjauan ahli (tambahan)** | Seorang engineer keamanan menguji ulang setiap temuan dalam satu laporan yang sudah selesai, menghapus false positive, dan menulis rencana perbaikan untuk stack Anda | **Harga sesuai permintaan** |
| **Whitebox (agen + ahli)** | Tinjauan kode sumber, pengujian terautentikasi, analisis logika bisnis dan rantai eksploit, laporan yang ditulis agen | **$10,000 USD per aplikasi / per domain** |

> Layanan whitebox ditentukan lingkupnya per aplikasi/domain dan mencakup seorang engineer
> keamanan khusus ditambah analisis mendalam dengan bantuan agen. Hubungi kami untuk menjadwalkan.

## Kasus penggunaan yang paling cocok

- **Pemeriksaan postur keamanan eksternal** yang cepat dan berulang atas properti web publik.
- **Asesmen smoke pra-peluncuran** sebelum rilis produksi.
- **Kebersihan kepatuhan** (header, TLS, data yang terbuka) sebagai bukti untuk audit.
- Lapisan dasar yang bisa **dieskalasi ke uji penetrasi manual/whitebox** untuk temuan kritis.

## Pendamping yang disarankan (opsional)

Untuk **aplikasi transaksional produksi** (fintech, e-commerce, layanan yang berhadapan dengan pemerintah),
padukan pemindaian eksternal otomatis dengan tingkat **whitebox** atau **uji penetrasi manual tambahan**
untuk mencakup logika bisnis terautentikasi dan verifikasi jalur eksploit yang lebih dalam.

---

## Pengungkapan kerentanan (sistem kami sendiri)

Menemukan sesuatu di netsekurity.com atau dalang.io? Email **security@dalang.io**. Kami
mengakui laporan dalam 2 hari kerja, tidak menuntut peneliti yang beritikad baik, dan mencantumkan kredit
atas perbaikan bila diminta. Mohon jangan menguji domain pelanggan lain.

---

*Cakupan per domain. 1 kredit = 1 asesmen eksternal · 1 domain. Contoh yang dikerjakan ada di
https://netsekurity.com/id/contoh-laporan.*
