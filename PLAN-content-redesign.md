# Arah konten dan struktur web NetSekurity

Tanggal: 26 Oktober 2026.

Dokumen acuan redesign konten dengan tema visual yang sebagian besar dipertahankan.
Arah positioning di bawah disetujui pemilik dalam percakapan. Redesign konten,
struktur halaman, panduan, dan pilihan verifikasi telah diimplementasikan dan diuji
di checkout lokal. Belum di-deploy. Cakupan scanner standar tidak diperluas oleh
redesign ini; kemampuan yang belum terbukti tidak dijadikan janji produk.

## Positioning yang disepakati

NetSekurity mempermudah pengujian keamanan yang bisa diulang mengikuti perubahan
aplikasi. Masalah utamanya adalah jarak antara frekuensi rilis aplikasi dan jadwal
pentest menyeluruh yang lebih jarang.

- **Offer:** uji keamanan aplikasi saat dibutuhkan; terima temuan, bukti, dan
  prioritas perbaikan melalui proses pemesanan yang praktis.
- **Persona awal:** founder teknis dan engineering lead tim kecil yang aktif
  merilis aplikasi dengan kapasitas security terbatas. Persona sekunder adalah
  tim yang sudah memiliki pentest berkala dan membutuhkan pengujian di antaranya.
- **Angle utama:** aplikasi terus berubah; pengujian keamanannya perlu bisa mengikuti.
- **Penguat konversi:** mudah dimulai, mudah diulang, biaya jelas, hasil dapat ditindaklanjuti.
- **Pemicu tambahan:** pengujian setelah perbaikan dan saat ada kerentanan relevan
  yang sudah didukung pemeriksaannya.
- **Bukti kualitas:** contoh temuan beserta bukti, dampak, prioritas, dan perbaikannya.
- **Mekanisme:** agent harness security khusus NetSekurity, dijelaskan setelah
  manfaat produk dan hanya sejauh kemampuan yang sudah terverifikasi.
- **Upsell:** pentest menyeluruh untuk tambahan kedalaman dan cakupan pengujian.
  Layanan utama harus tetap berguna tanpa pembelian tambahan.

Rumusan offer awal:

> Uji keamanan aplikasi kapan dibutuhkan. Dapatkan temuan, bukti, dan prioritas
> perbaikan tanpa proses pemesanan pentest yang panjang.

Contoh angle awal untuk audiens engineering: “Deploy terus. Uji keamanannya juga.”
Untuk pembeli nonteknis, gunakan bahasa yang lebih dekat dengan aktivitas mereka:
“Website terus berubah. Periksa keamanannya secara berkala.”

Kemudahan memulai dapat dirumuskan sebagai “Tidak perlu menjadi ahli keamanan
untuk mulai menguji aplikasi Anda.” Penggunaan layanan tidak berarti seluruh
perbaikan dapat dilakukan tanpa developer.

## Penajaman audiens dari percakapan lanjutan

Pemilik menekankan bahwa calon pembeli kemungkinan besar tidak berlatar teknis.
Penajaman yang diterapkan: pemilik bisnis atau aplikasi menjadi pembaca utama halaman
penjualan; developer atau vendor mereka menjadi pelaksana verifikasi, perbaikan,
dan integrasi. Kedua peran perlu mendapat informasi yang sesuai.

Homepage harus membuat pembeli memahami masalah, hasil, harga, dan langkah
berikutnya. Contoh laporan dan dokumentasi menyediakan rincian bagi pelaksana teknis.
Detail cakupan dan kebutuhan akses tetap terlihat sebelum pembelian.

## Navigasi yang diterapkan

Navigasi utama: Cara kerja, Contoh hasil, Harga, Pentest menyeluruh.
Tindakan akun: Masuk dan tombol utama Mulai pemeriksaan.
Cara kerja dan Harga dapat berupa anchor di homepage pada tahap awal.

Halaman pendukung:

- Contoh hasil: ringkasan bahasa awam, contoh temuan, bukti dan detail untuk developer.
- Pentest menyeluruh: kapan diperlukan, cakupan, proses, deliverable dan konsultasi.
- Cakupan pemeriksaan: area yang didukung, batas dan kebutuhan akses.
- Keamanan dan privasi: operator, cara pemeriksaan, akses laporan dan penanganan data.
- Panduan memulai: verifikasi kepemilikan, menjalankan pemeriksaan dan membaca hasil.
- Dokumentasi developer: API dan integrasi CI/CD.
- Kontak dan informasi perusahaan; informasi pengadaan tetap tersedia sebagai pendukung.

URL yang sudah ada perlu dipertahankan atau diarahkan dengan benar saat implementasi.
Halaman dan tautan bahasa Indonesia dan Inggris tetap konsisten.

## Urutan homepage yang diterapkan

| Urutan | Pertanyaan pengunjung | Isi |
|---|---|---|
| 1 | Apa manfaatnya untuk saya? | Hero dengan pengujian praktis dan berulang, CTA mulai, tautan contoh hasil, harga awal yang diambil dari sumber harga |
| 2 | Mengapa perlu diperiksa lagi? | Situasi setelah update, setelah perbaikan dan di antara jadwal pentest; CVE hanya jika kemampuan relevan tersedia |
| 3 | Apa yang saya dapatkan? | Contoh satu temuan: masalah, dampak, tindakan dan bukti; ringkasan untuk pemilik dan rincian untuk developer |
| 4 | Apa yang harus saya lakukan? | Masukkan alamat, verifikasi kepemilikan, jalankan pemeriksaan, baca hasil dan teruskan kepada developer |
| 5 | Apa yang diperiksa? | Cakupan dalam bahasa awam, batas ringkas dan tautan metodologi lengkap; ringkasan cara menjaga data dan akses |
| 6 | Berapa biaya untuk mulai dan mengulang? | Satu pemeriksaan dan paket berulang, unit pembelian, masa berlaku, durasi serta biaya pengujian ulang yang jelas |
| 7 | Bagaimana jika perlu pemeriksaan lebih dalam? | Pentest menyeluruh sebagai layanan lanjutan dengan cakupan tambahan |
| 8 | Apa yang masih membuat saya ragu? | FAQ tentang bantuan teknis, akses, produksi, hasil, perbaikan, privasi dan CTA akhir |

## Bahasa dan penyajian

- Gunakan “pemeriksaan keamanan” untuk memperkenalkan layanan. Istilah “pentest”
  dijelaskan singkat bila relevan dan hanya digunakan untuk cakupan yang mendukungnya.
- Mulai dengan masalah dan dampak bagi pemilik aplikasi. Tampilkan istilah teknis
  beserta bukti pada rincian temuan.
- Contoh ilustratif: “File konfigurasi bisa diakses publik” → “File seperti ini
  dapat berisi pengaturan atau kredensial” → “Minta developer membatasi akses dan
  memeriksa apakah ada rahasia yang terpapar.” Jangan menyatakan rahasia ditemukan
  jika bukti hanya menunjukkan respons HTTP yang masih perlu divalidasi.
- Jelaskan CVE sebagai kerentanan yang diketahui pada perangkat lunak; jelaskan
  relevansinya dengan aplikasi pelanggan sebelum menampilkan nomor referensinya.
- Jangan menaruh CVSS, CWE, OWASP, blackbox, whitebox atau agent harness di hero.
- Pertahankan detail penting sebelum pembelian: kebutuhan akses, cakupan, waktu,
  harga, bahasa laporan dan keterlibatan manusia.
- Hindari klaim “selalu aman”, bebas semua CVE, tanpa false positive, hasil instan
  atau tanpa kebutuhan teknis sama sekali.

## Pengalaman setelah tombol mulai

Bahasa sederhana harus berlanjut sampai hasil pemeriksaan.

1. Minta alamat website dan jelaskan kebutuhan akses serta harga sebelum pembelian.
2. Jelaskan verifikasi sebagai bukti kendali atas website. Konfirmasi kewenangan
   pelanggan untuk mengizinkan pengujian dan cakupannya sebelum mulai.
3. Sediakan instruksi verifikasi yang dapat disalin dan dikirim sendiri oleh pemilik
   kepada pengelola domain; tersedia sebagai tombol salin instruksi.
4. Setelah prasyarat terpenuhi, tampilkan ringkasan cakupan dan biaya untuk memulai.
5. Hasil mendahulukan hal yang perlu diperhatikan dan langkah selanjutnya, lalu
   rincian teknis bagi developer. Pengiriman laporan dilakukan melalui mekanisme
   yang menjaga kerahasiaan, bukan tautan publik.
6. Setelah perbaikan, arahkan pada pemeriksaan ulang dengan biaya yang jelas.

Belum ada temuan tidak boleh ditampilkan sebagai jaminan aplikasi aman. Pengujian
yang gagal atau area yang tidak terjangkau harus dibedakan dari hasil tanpa temuan.

## Opsi verifikasi yang disepakati

Pemilik menyetujui penambahan pilihan verifikasi pada 26 Oktober 2026. DNS tetap
dipertahankan; file verifikasi dan meta tag sudah menjadi opsi tambahan di dashboard,
backend, dan panduan publik, dengan pengujian positif serta negatif.

| Metode | Tindakan pelanggan | Penggunaan |
|---|---|---|
| DNS TXT | Tambahkan token ke pengaturan DNS | Website atau hostname API, bagi pelanggan dengan akses domain |
| File verifikasi melalui HTTPS | Tempatkan file berisi token di `/.well-known/netsekurity-verification.txt` | Pelanggan dengan akses hosting atau bantuan developer; API dapat menyediakan endpoint pada lokasi tersebut |
| Meta tag | Tambahkan `<meta name="netsekurity-verification" content="TOKEN_UNIK">` pada `<head>` halaman utama | CMS atau website builder yang mendukung pengaturan tersebut; tidak untuk API tanpa HTML |

Ketiganya tersedia. Jalur bantuan pengelola hanya menyalin instruksi; aplikasi
tidak mengirim pesan kepada pihak lain.

Antarmuka menggunakan judul “Konfirmasi akses ke website Anda” dengan pilihan
berdasarkan akses yang dimiliki pelanggan:

- “Saya bisa mengubah pengaturan website” → panduan meta tag.
- “Saya bisa mengunggah file ke hosting” → file dan panduan penempatan.
- “Saya bisa mengatur domain” → panduan DNS TXT.
- “Website saya dikelola orang lain” → instruksi yang dapat disalin dan dikirim
  sendiri kepada pengelola website. Ini jalur bantuan, bukan metode verifikasi baru.

Persyaratan implementasi:

- Token unik terikat pada akun dan hostname; token publik bukan kredensial login.
- File atau tag diperiksa pada lokasi yang ditentukan, bukan sembarang halaman,
  komentar atau unggahan pengguna.
- Pemeriksaan HTTPS memvalidasi sertifikat. Redirect tidak boleh memperluas
  target yang diverifikasi; fetcher harus menolak tujuan jaringan internal,
  termasuk setelah resolusi ulang DNS atau redirect.
- Verifikasi satu hostname tidak otomatis memberikan izin menguji hostname lain,
  subdomain, alamat IP bersama atau seluruh infrastruktur hosting.
- Periksa kembali bukti kendali sebelum pengujian. Pelanggan tidak perlu memasang
  ulang token selama bukti masih valid; jika tidak dapat dikonfirmasi, tahan
  pengujian dan sediakan jalur mencoba kembali atau verifikasi ulang.
- Bukti kendali dipisahkan dari konfirmasi kewenangan dan cakupan pengujian.
- Email berdomain sama dan login Google biasa tidak diterima sebagai bukti tunggal.
- Pilihan metode dan panduan publik baru ditampilkan setelah tersedia dan diuji.

## Batas kemampuan yang dipertahankan

- Status harness: sudah berjalan, sedang dikembangkan, atau masih arah produk.
  Pertanyaan ini belum terjawab dalam percakapan.
- Repo saat ini mendeskripsikan asesmen eksternal read-only tanpa autentikasi,
  pengujian injeksi, otorisasi atau logika bisnis. Jangan mengubah fakta ini hanya
  dengan mengganti copy menjadi “pentest mendalam”.
- Dukungan pengujian API, pemeriksaan CVE, pembaruan pengujian dan metode validasi
  harus dibuktikan sebelum dijadikan janji produk.
- Tidak ada penjadwal bawaan yang terdokumentasi. Pengulangan lewat dashboard
  atau pemanggilan API tidak sama dengan pemantauan otomatis sepanjang waktu.
- Laporan pelanggan saat ini didokumentasikan sebagai PDF berbahasa Inggris.
  Ringkasan tindakan berikutnya kini tersedia dalam bahasa Inggris dan Indonesia
  di dashboard, berdasarkan jumlah temuan yang benar-benar dilaporkan.
- Panduan setup dan salin instruksi tersedia. Pengiriman instruksi otomatis serta
  perbandingan hasil otomatis tidak ditawarkan. Pemeriksaan ulang memakai 1 kredit
  dan dinyatakan jelas sebelum memulai.
- Harga, durasi dan klaim publik mengacu pada sumber yang dipelihara di CLAIMS.md.

## Acuan repo

- [Klaim dan sumber](CLAIMS.md)
- [Cakupan produk saat ini](static/PRODUCT_SCOPE.id.md)
- [Konten Indonesia saat ini](site_copy_id.go)
- [Templat halaman saat ini](site_html.go)

Simpan keputusan berikutnya di dokumen ini dengan membedakan keputusan yang
disepakati, usulan, dan kemampuan produk yang sudah terverifikasi.


## Audit implementasi lokal — 26 Oktober 2026

| Persyaratan | Implementasi / bukti pemeriksaan |
|---|---|
| Tema dipertahankan; persona pemilik aplikasi didahulukan | Template dan CSS yang sama; hero EN/ID menggunakan manfaat dan tindakan sederhana. Screenshot desktop dan viewport 390 px diperiksa. |
| Offer, persona, angle konsisten | `site_copy_en.go`, `site_copy_id.go`, `marketing-copy.md`; pemeriksaan setelah pembaruan, perbaikan, dan di antara pentest. |
| Navigasi empat pilihan dan alur keputusan homepage | `site.go`, `site_html.go`; `TestHomeFollowsTheOwnersDecisionPath`. |
| Contoh hasil, bukti, tindakan dan batas klaim | Kartu contoh dengan validasi developer, halaman contoh/PDF yang sudah ada, `CLAIMS.md`; suite kualitas halaman dan kesegaran PDF lolos. |
| Harga, durasi, biaya pengulangan, bantuan developer | Harga dari sumber produk, durasi pemilik 10 menit–24 jam, 1 kredit per pemeriksaan ulang; panduan EN/ID dan FAQ. |
| Upsell pentest menyeluruh | Bagian ringkas homepage dan halaman `/enterprise` EN/ID; detail teknis dapat dibuka setelah penjelasan manfaat. |
| Panduan pemilik dan dokumentasi developer terpisah | `/getting-started`, `/id/panduan-memulai`; contoh API tetap di dokumentasi/otomasi. |
| Semua URL, bahasa, metadata, tautan dan sitemap valid | Registry halaman dipertahankan dan ditambah panduan; suite `site_test.go`, `site_quality_test.go`, dan keamanan halaman lolos. |
| DNS, file, meta tag berfungsi | `verification.go`, `domains.go`; `TestHTTPSVerificationProof`, `TestVerificationMethodsPersistAndDownloadIsAccountBound`. |
| Token terikat akun dan hostname; lokasi bukti tepat | Uji akun lain, token salah, komentar/script/template/body, HTML salah lokasi, respons terlalu besar, status gagal. |
| HTTPS valid, tanpa redirect, menolak jaringan internal | `TestVerificationRejectsSpecialNetworks`, `TestVerificationRejectsUntrustedTLS`, uji redirect dan hostname tidak valid. Koneksi menggunakan IP hasil resolusi yang telah diperiksa. |
| Salin instruksi dan unduh file | Dashboard empat pilihan akses; 11 asersi browser termasuk disclosure, clipboard, unduh file, checkbox izin, error HTMX, dan layout setelah panel dibuka. |
| Izin dan bukti diperiksa sebelum mulai | `authorised=true` wajib, timestamp disimpan; `TestStartRequiresPermissionAndCurrentProofWithoutSpendingOnFailure`. |
| Bukti hilang saat antre tidak diteruskan ke scanner | `TestWorkerDoesNotReceiveTargetAfterProofIsRemoved`; gagal dan refund satu kali. Kontrak partner tetap terpisah. |
| Kredit dan antrean konsisten | Transaksi atomik; `TestConcurrentStartsChargeOnce`, `TestQueueFailureRollsBackCreditAndLedger`. |
| Hasil mudah dibaca dan tidak mengarang jaminan | `result_summary.go`; `TestResultGuidanceDistinguishesMissingAndZero`; status antre/gagal dipisahkan dari hasil tanpa temuan, PDF tetap privat. |
| Aset sesuai konten baru | CSS dibangun ulang; kartu OG EN/ID dirender ulang dengan font lokal. PDF contoh tidak berubah dan pemeriksaan kesegarannya lolos. |

Validasi selesai: `go test ./...`, `go test -race ./...`, `go vet ./...`,
`go build`, `npm run build:css`, dan `git diff --check`. Pemeriksaan browser lokal
mencakup 21 halaman pada 320, 390, dan 1280 px (63 kombinasi), tanpa overflow.
Sebelas asersi interaksi lulus. Browser memakai database QA terisolasi dan proxy
loopback khusus QA; tidak menggunakan akun pelanggan, mengirim pesan, membeli
kredit, atau menjalankan scan publik. Kegagalan verifikasi dalam pengujian browser
disimulasikan; validasi verifier sebenarnya diuji dengan fixture HTTP/TLS Go.

### Catatan rilis dan batas validasi

- Belum commit/push/deploy. Produksi dan worker eksternal belum diverifikasi ulang.
- Integrasi API yang memulai pemeriksaan harus menambahkan `authorised=true` pada
  form, atau `"authorised": true` pada JSON. Berlaku juga untuk alias `/pentests`.
  Seluruh contoh integrasi yang dikirim repo sudah diperbarui.
- Migrasi startup menambahkan `pentests.authorized_at`; data lama tetap berlaku,
  tetapi bukti kendali diperiksa kembali sebelum scan mandiri dijalankan.
- Dashboard utama tetap berbahasa Inggris dengan panduan EN/ID dan ringkasan hasil
  dwibahasa. PDF tetap bahasa Inggris. Tidak ada klaim lokalisasi penuh.
- Instalasi dependency frontend melaporkan 4 dependency build rentan yang
  sudah ada. Tidak diubah dalam redesign ini; ini bukan audit keamanan seluruh repo.
- Pernyataan tentang harness lebih luas, cakupan CVE baru, dan operasi produksi
  memerlukan bukti terpisah. Tidak dijadikan janji maupun ditandai selesai di sini.
