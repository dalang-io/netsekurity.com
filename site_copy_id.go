package main

import "html/template"

// Salinan Bahasa Indonesia untuk situs publik. Strukturnya identik dengan
// site_copy_en.go (diuji oleh TestSiteCopyIsComplete). Token dalam kurung
// kurawal diisi saat render dari credit_packages. Sumber setiap pernyataan ada
// di CLAIMS.md.
//
// Istilah (TestIndonesianTermsAreConsistent menjaganya):
//   - "Direksi" untuk board, bukan "dewan"; "engineering lead" tidak diterjemahkan.
//   - "read-only" tidak diterjemahkan ("hanya-baca" terdengar seperti terjemahan mesin).
//   - "asesmen" untuk assessment, "uji penetrasi" untuk penetration test.
//   - "false positive", "deploy", "pipeline", "token" dipakai apa adanya seperti
//     di kalangan engineer Indonesia.
//   - "Netsekurity" untuk merek; "NetSekurity Scanner" adalah User-Agent yang dikirim pemindai.

func commonID() commonCopy {
	return commonCopy{
		Skip: "Langsung ke konten", MenuOpen: "Buka menu", MenuClose: "Tutup menu", LangLabel: "Bahasa",
		NavHow: "Cara kerja", NavEnterprise: "Pentest menyeluruh", NavTrust: "Keamanan dan privasi", NavPricing: "Harga",
		NavSample: "Contoh hasil", NavDocs: "Dokumentasi developer",
		LogIn: "Masuk", Dashboard: "Dasbor", Start: "Mulai pemeriksaan", Buy: "Beli",
		Tagline:    "Pemeriksaan keamanan website yang praktis dan mudah diulang.",
		Operator:   "Netsekurity adalah produk Dalang Pte. Ltd. (UEN 202622331M), 160 Robinson Road, #14-04, Singapura 068914.",
		Rights:     "© 2026 Dalang Pte. Ltd.",
		ColProduct: "Produk", ColEnterprise: "Enterprise", ColResources: "Sumber daya",
		LinkStartGuide: "Panduan memulai", LinkScope: "Cakupan produk", LinkGuide: "Asesmen vs uji penetrasi", LinkLLMS: "llms.txt",
		LinkSecurityTxt: "security.txt", LinkSitemap: "Peta situs", LinkDalang: "Dalang.io", LinkPrivacy: "Kebijakan privasi", LinkTerms: "Ketentuan penggunaan", LinkTrustCenter: "Pusat keamanan dan kepercayaan",
		LinkContact: "Hubungi tim penjualan", LinkProcurement: "Lembar fakta pengadaan",
		Breadcrumb: "Beranda",
		CopyLabel:  "Salin", CopiedLabel: "Tersalin", PrintLabel: "Cetak atau simpan sebagai PDF",
		Form: formCopy{
			Title:              "Bicara dengan kami",
			Lead:               "Pertanyaan soal cakupan, volume, ketentuan pembelian, atau tingkat whitebox? Kirim pesan, atau email sales@dalang.io.",
			Name:               "Nama (opsional)",
			Email:              "Email kantor",
			Message:            "Pesan",
			MessagePlaceholder: "Apa fungsi website atau aplikasi Anda, dan bagian mana yang ingin diperiksa?",
			Send:               "Kirim pesan",
			Reply:              "Kami membalas dalam 1 hari kerja.",
		},
		Coverage: coverageCopy{
			H2:       "Persis apa yang diperiksa, dan apa yang tidak",
			Lead:     "Pemeriksaannya berupa daftar tetap yang sama untuk semua pelanggan. Tidak ada yang tersirat di luar kata-katanya.",
			InTitle:  "Termasuk di setiap asesmen",
			OutTitle: "Tidak termasuk",
			In: []kv{
				{"Pengintaian permukaan serangan", "Resolusi DNS dan daftar subdomain pasif dari certificate transparency dan sumber publik. Subdomain didaftar sebagai informasi; hanya domain yang Anda verifikasi yang dinilai."},
				{"Header keamanan", "HSTS, Content-Security-Policy, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, dan Permissions-Policy, ditambah perlindungan clickjacking."},
				{"Sidik jari web", "Header server, pengungkapan versi server, dan deteksi web application firewall, dibaca dari respons."},
				{"Sertifikat TLS", "Masa berlaku, sisa hari sampai kedaluwarsa, penerbit, dan jumlah entri SAN. Handshake yang gagal dilaporkan sebagai temuan."},
				{"File dan path yang terbuka", "13 probe, termasuk .git, .env, arsip cadangan, server-status, phpinfo, metadata SVN, dan path admin atau login."},
				{"Open redirect", "Sembilan parameter redirect yang umum, diuji apakah mengarahkan ke alamat eksternal."},
			},
			Out: []kv{
				{"Apa pun di balik login", "Area yang memerlukan autentikasi tidak diuji."},
				{"Otorisasi dan IDOR", "Memerlukan beberapa identitas; tidak diuji."},
				{"Pengujian injeksi", "Tidak ada payload SQLi, XSS, LFI/RFI, atau path traversal yang dikirim. Pemindaian menemukan paparan dan salah konfigurasi, bukan bug input yang bisa dieksploitasi."},
				{"Logika bisnis", "Pengujian alur kerja, pembayaran, dan race condition."},
				{"Pemindaian port jaringan", "Asesmen standar tidak memindai port."},
				{"Kelas lain", "SSRF, GraphQL, WebSocket, request smuggling, cache poisoning, analisis kode sumber, dan rantai eksploit bertahap."},
			},
			Note: "Temuan yang tidak muncul berarti “tidak dipindai”, bukan “terbukti aman”. Permintaan bersifat read-only: tidak ada yang ditulis, diunggah, atau dieksploitasi.",
		},
		TiersH2:   "Perlu pemeriksaan lebih mendalam?",
		TiersLead: "Untuk area setelah login, hak akses antar pengguna, dan alur transaksi, diskusikan pentest menyeluruh. Cakupan disepakati sesuai aplikasi Anda sebelum pengujian dimulai.",
		Tiers: []tier{
			{Name: "Asesmen eksternal otomatis", Price: "Mulai {single} per domain", Audience: "Terus periksa apa yang bisa dilihat internet publik.",
				Points: []string{"35+ pemeriksaan read-only", "Tingkat keparahan, skor CVSS v3.1, CWE, dan kategori OWASP per temuan", "Mulai dari dasbor atau dari CI/CD", "Kredit tidak kedaluwarsa"}},
			{Name: "Tambahan tinjauan ahli", Price: "Harga sesuai permintaan", Audience: "Minta seorang engineer memastikan temuannya.",
				Points: []string{"Seorang engineer keamanan menguji ulang setiap temuan dalam satu laporan yang sudah selesai", "False positive dihapus", "Rencana perbaikan yang ditulis untuk stack Anda"}},
			{Name: "Pentest menyeluruh", Price: "{whitebox} per aplikasi atau domain", Audience: "Uji apa yang hanya bisa dijangkau penguji yang terautentikasi dan berpengetahuan.",
				Points: []string{"Tinjauan kode sumber dan pengujian terautentikasi", "Analisis logika bisnis dan rantai eksploit", "Seorang engineer keamanan khusus dengan bantuan analisis agen", "Lingkup ditentukan per aplikasi atau domain"}},
		},
	}
}

func homeID() homeCopy {
	return homeCopy{
		Title:       "Netsekurity: Pemeriksaan Keamanan Website yang Praktis",
		Description: "Periksa website setelah pembaruan dan di antara jadwal pentest. Dapatkan temuan, bukti, dan panduan untuk developer. Ulangi saat dibutuhkan. Mulai {single}.",
		Eyebrow:     "Pemeriksaan keamanan website yang mudah diulang",
		H1:          "Website terus berubah. Periksa keamanannya secara berkala.",
		Lead:        "Ketahui masalah pada bagian publik website, pahami dampaknya, dan berikan panduan tindakan kepada developer Anda. Jalankan pemeriksaan kembali setelah pembaruan atau perbaikan.",
		SubLead:     "Mulai dari satu website dengan {single}. Anda tidak perlu menjadi ahli keamanan untuk memulai; pengelola website bisa membantu persiapan dan perbaikannya.",
		CTAPrimary:  "Mulai pemeriksaan", CTASecondary: "Lihat contoh hasil", CTATertiary: "Mulai dari satu website · {single}",
		SignInNote:   "Masuk dengan Google, tambahkan website, dan verifikasi akses sebelum membeli kredit.",
		Assurances:   []string{"Bayar per pemeriksaan", "Bukti dan panduan perbaikan", "Ulangi setelah pembaruan"},
		Card:         sampleCard{Label: "Contoh hasil", Domain: "shop.example.com", Summary: "File konfigurasi perlu diperiksa", Posture: "Prioritas: segera telusuri", Findings: []string{"Yang terlihat: alamat file konfigurasi memberikan respons berhasil.", "Mengapa penting: file yang benar-benar terbuka bisa berisi pengaturan privat atau kredensial.", "Langkah berikutnya: minta developer memvalidasi respons dan membatasi akses jika file memang terbuka."}, Note: "Contoh ilustratif. Respons berhasil saja belum membuktikan adanya rahasia yang terpapar.", Link: "Lihat bukti dan langkah berikutnya"},
		DefinitionH2: "Apa yang akan diperiksa?",
		Definition:   "Netsekurity memeriksa bagian publik website untuk mencari file yang terbuka, perlindungan browser yang belum terpasang, masalah sertifikat, dan pengalihan tautan yang berisiko. Agen analisis AI menyusun temuan dan laporan. Layanan standar adalah asesmen eksternal otomatis; tidak ada manusia yang meninjau laporan sebelum dikirim.",
		ScopeLegend: []kv{
			{"Dipindai", "Permukaan publik: halaman, file, header, sertifikat, dan redirect yang bisa diminta siapa pun di internet."},
			{"Tidak dipindai: di balik login", "Area yang memerlukan login tidak pernah diuji."},
			{"Tidak dipindai: data dan logika", "Database, pembayaran, dan logika alur kerja berada di luar jangkauan pemeriksaan eksternal yang read-only."},
		},
		TypicalH3: "Contoh masalah yang dicari",
		Typical: []string{
			"File .env atau .git yang diberikan web server kepada siapa pun yang memintanya",
			"Header keamanan yang tidak ada pada respons",
			"Sertifikat yang kedaluwarsa bulan depan",
			"Redirect yang membuat tautan phishing membawa domain Anda",
		},
		AudiencesH2: "Kapan website perlu diperiksa lagi?",
		Audiences:   []audience{{Role: "Setelah pembaruan", Tagline: "Website Anda sudah berubah.", Bullets: []string{"Halaman baru dan perubahan pengaturan dapat mengubah bagian yang bisa diakses publik.", "Periksa kondisi website saat ini, meskipun versi sebelumnya pernah diuji."}, LinkLabel: "Lihat cara kerja", Anchor: "how"}, {Role: "Setelah perbaikan", Tagline: "Periksa hasil perbaikannya.", Bullets: []string{"Jalankan kembali pemeriksaan dengan cakupan yang sama setelah developer melakukan perubahan.", "Tinjau laporan baru; setiap pemeriksaan ulang memakai 1 kredit."}, LinkLabel: "Lihat cara kerja", Anchor: "how"}, {Role: "Di antara jadwal pentest", Tagline: "Tetap periksa di sela jadwal.", Bullets: []string{"Gunakan pemeriksaan eksternal di antara jadwal pentest menyeluruh.", "Pengujian setelah login dan logika bisnis membutuhkan layanan dengan cakupan lebih dalam."}, LinkLabel: "Lihat cara kerja", Anchor: "how"}},
		ReportH2:    "Pahami masalahnya. Ketahui langkah berikutnya.",
		ReportLead:  "Anda mendapat laporan PDF berbahasa Inggris untuk setiap website yang diperiksa. Baca ringkasannya, lalu berikan bukti dan panduan perbaikan kepada developer.",
		ReportLink:  "Lihat contoh hasil pemeriksaan",
		Report:      []kv{{"Pahami masalahnya", "Lihat apa yang ditemukan dan mengapa hal itu perlu diperhatikan."}, {"Tentukan tindakan berikutnya", "Gunakan tingkat keparahan dan dampak untuk memilih masalah yang perlu ditelusuri developer terlebih dahulu."}, {"Berikan detail kepada developer", "Laporan memuat bukti, langkah reproduksi, dan panduan perbaikan."}, {"Periksa lagi setelah perubahan", "Jalankan asesmen baru dan tinjau hasilnya terhadap perbaikan. Setiap pemeriksaan ulang memakai 1 kredit."}},
		PlainH3:     "Contoh temuan dalam bahasa sehari-hari", PlainLink: "Lihat bukti dan detail teknis",
		Plain:  []kv{{"Yang terlihat", "Alamat file konfigurasi memberikan respons berhasil."}, {"Mengapa penting", "File konfigurasi yang terbuka bisa berisi pengaturan sensitif. Respons tersebut perlu divalidasi."}, {"Yang perlu diminta kepada developer", "Pastikan apakah file benar-benar terbuka, batasi aksesnya, dan periksa apakah ada kredensial yang perlu diganti."}},
		HowH2:  "Dari alamat website sampai langkah perbaikan",
		Steps:  []kv{{"Tambahkan website", "Masuk dengan Google dan masukkan alamat website. Menambahkan website dan memverifikasi akses tidak dipungut biaya."}, {"Konfirmasi akses", "Pilih record DNS, file verifikasi, atau tag halaman utama. Salin instruksi untuk pengelola website bila membutuhkan bantuan."}, {"Mulai pemeriksaan", "Tinjau cakupan dan konfirmasi izin untuk menguji hostname ini. Setiap pemeriksaan memakai 1 kredit dan memerlukan waktu 10 menit hingga 24 jam, tergantung temuannya."}, {"Tindak lanjuti hasilnya", "Unduh PDF berbahasa Inggris, tinjau prioritas, dan berikan detailnya kepada developer. Ulangi pemeriksaan setelah perbaikan."}},
		BandH2: "Mulai dari website yang digunakan pelanggan Anda.", BandBody: "Periksa sekarang, lalu ulangi setelah pembaruan atau perbaikan. Setiap pemeriksaan memiliki cakupan dan biaya per kredit yang jelas.",
		PricingH2:   "Mulai sekali. Ulangi saat dibutuhkan.",
		PricingLead: "1 kredit = 1 pemeriksaan untuk 1 hostname terverifikasi. Pemeriksaan ulang juga memakai 1 kredit. Harga dalam dolar AS; subdomain yang berbeda dihitung sebagai target terpisah.",
		PricingNote: "Kredit tidak kedaluwarsa. Pemeriksaan yang gagal mendapat pengembalian kredit otomatis. Pemindaian memerlukan waktu 10 menit hingga 24 jam, tergantung temuannya. Satu pemeriksaan dapat berjalan pada satu waktu per akun.",
		PricingCols: []string{"Paket", "Kredit", "Harga", "Per asesmen", "Pesan"},
		PricingBest: "Harga per asesmen terendah",
		PricingTalk: "Perlu bantuan memilih cakupan yang sesuai?", PricingTalkBtn: "Bicara dengan kami",
		TrustH2:   "Akses dan hasil pemeriksaan tetap terkendali",
		TrustLead: "Sebelum mulai, pahami cara kami memverifikasi website dan menjalankan pemeriksaan.",
		TrustLink: "Baca tentang keamanan dan privasi",
		Trust:     []kv{{"Akses diperiksa kembali", "Bukti DNS, file, atau meta tag diperiksa sebelum setiap pemindaian mandiri. Biarkan terpasang untuk pemeriksaan berikutnya."}, {"Eksternal dan read-only", "Pemeriksaan standar tidak menulis ke website Anda atau menguji alur setelah login."}, {"Cakupan yang jelas", "Satu hostname terverifikasi per pemeriksaan. Subdomain lain dan pengujian aplikasi yang lebih dalam memiliki cakupan terpisah."}, {"Laporan privat", "Laporan tersedia bagi akun pemesan dan operator yang berwenang, melalui akses yang terlindungi."}},
		FAQH2:     "Sebelum Anda mulai",
		FAQ: []qa{
			{"Apakah saya perlu pengalaman teknis?", `Anda bisa memesan pemeriksaan tanpa keahlian keamanan. Pengelola website mungkin perlu membantu memasang token verifikasi dan melakukan perbaikan. Dasbor menyediakan instruksi yang bisa Anda salin dan kirim kepadanya.`},
			{"Apakah Netsekurity memperbaiki website saya?", `Layanan standar memberikan temuan dan panduan perbaikan. Developer atau penyedia website Anda melakukan perubahannya. Pemeriksaan ulang memakai 1 kredit.`},
			{"Apakah semua CVE baru diperiksa?", `Tidak. CVE adalah identitas kerentanan perangkat lunak yang sudah diketahui publik. Layanan standar memeriksa daftar paparan eksternal dan masalah konfigurasi yang dipublikasikan; layanan ini tidak menjanjikan deteksi setiap CVE baru atau pemantauan terus-menerus.`},
			{"Bagaimana jika laporan tidak memiliki temuan?", `Artinya tidak ada temuan yang dilaporkan dalam cakupan pemeriksaan tersebut. Area di luar cakupan tetap belum diuji, dan pemeriksaan gagal atau tidak lengkap tidak boleh dianggap sebagai hasil bersih.`},
			{"Apa itu Netsekurity?", `Netsekurity adalah layanan asesmen keamanan web eksternal yang dioperasikan oleh Dalang Pte. Ltd. Layanan ini memeriksa situs web milik Anda dari luar, read-only dan tanpa login, lalu mengirim laporan PDF terprioritas berisi temuan berskor CVSS, bukti, dan rencana perbaikan 30/60/90 hari.`},
			{"Apakah ini uji penetrasi?", `<strong>Bukan.</strong> Ini asesmen otomatis, eksternal, tanpa autentikasi, dan read-only. Layanan ini tidak menguji apa pun di balik login, otorisasi (IDOR), logika bisnis, atau celah injeksi. Untuk risiko tersebut gunakan <a href="/id/enterprise#escalation">tinjauan ahli atau tingkat whitebox</a>, atau uji penetrasi manual. <a href="/id/asesmen-eksternal-vs-uji-penetrasi">Perbandingan lengkapnya</a> menjelaskan bedanya.`},
			{"Apa saja yang dicakup satu asesmen?", `Lebih dari 35 pemeriksaan read-only pada domain yang Anda verifikasi: DNS dan daftar subdomain pasif, enam header keamanan dan perlindungan clickjacking, pengungkapan versi server, deteksi WAF, masa berlaku dan kedaluwarsa sertifikat TLS, 13 probe file terbuka dan path admin, serta sembilan parameter open redirect. Lihat <a href="/id/cakupan-produk">cakupan produk</a> untuk daftar persisnya.`},
			{"Apa yang tidak dicakup?", `Apa pun di balik login, otorisasi dan IDOR, logika bisnis, pengujian injeksi (SQLi, XSS, LFI/RFI), pemindaian port jaringan, SSRF, GraphQL, WebSocket, request smuggling, cache poisoning, analisis kode sumber, dan rantai eksploit bertahap. Temuan yang tidak muncul berarti “tidak dipindai”, bukan “terbukti aman”.`},
			{"Bagaimana kepemilikan domain diverifikasi?", `Pilih DNS TXT, file verifikasi melalui HTTPS, atau meta tag di halaman utama publik. Token terikat pada akun dan hostname Anda. Bukti diperiksa sebelum setiap pemindaian mandiri. Pengelola website dapat mengikuti instruksi tanpa password Netsekurity Anda.`},
			{"Berapa biayanya?", `1 kredit = 1 asesmen untuk 1 domain. Satu kredit seharga {single}; pada paket {volumeCredits} kredit, setiap asesmen seharga {volume}. Harga dalam dolar AS dan kredit tidak kedaluwarsa. Tinjauan ahli berharga sesuai permintaan, dan tingkat whitebox {whitebox} per aplikasi atau domain.`},
			{"Apakah ada manusia yang terlibat?", `Tidak pada asesmen standar. Agen analisis AI menghapus duplikat, mengurutkan temuan, dan menulis laporan; tidak ada manusia yang membacanya sebelum dikirim. Jika Anda ingin seorang engineer menguji ulang setiap temuan, itulah tambahan tinjauan ahli.`},
			{"Apakah Netsekurity memakai AI, dan apa yang dilihatnya?", `Agen analisis AI menjalankan asesmen dan melihat keluaran pemindai: temuan, kode status, nilai header, dan string bukti singkat. Probe paparan hanya mencatat kode status dan ukuran, lalu membuang isi respons. Penyedia model dan wilayahnya belum disebut di situs ini: tanyakan ke <a href="mailto:security@dalang.io">security@dalang.io</a> sebelum membeli. <a href="/id/trust#third-parties">Daftar pihak ketiga</a> menjelaskan siapa lagi yang menangani data Anda.`},
			{"Bisakah dijalankan dari CI/CD?", `Bisa. Buat token API di dasbor lalu panggil <code>POST /api/v1/assessments</code> pada setiap deploy, kemudian baca hasilnya dengan <code>GET /api/v1/assessments/{id}</code>. Berfungsi dengan sistem apa pun yang bisa mengirim permintaan HTTP; <a href="/id/dokumentasi">dokumentasi</a> memuat contoh GitHub Actions, GitLab CI, Jenkins, dan delapan lainnya.`},
			{"Apakah cocok untuk aplikasi yang dibuat dengan alat pemrograman AI?", `Cocok. Asesmen eksternal memeriksa apa yang diperlihatkan sebuah situs web, bukan cara kodenya ditulis, sehingga berlaku sama untuk kode tulisan tangan maupun kode buatan asisten pemrograman AI. File environment yang terbuka dan header keamanan yang hilang terjadi pada keduanya.`},
			{"Berapa lama pemindaian berlangsung?", `Sebuah pemindaian memerlukan waktu 10 menit hingga 24 jam, tergantung temuannya. Anda boleh meninggalkan halaman: laporan muncul di dasbor Anda setelah siap.`},
			{"Siapa di balik Netsekurity?", `Netsekurity adalah produk Dalang Pte. Ltd., perusahaan Singapura (UEN 202622331M). Lihat <a href="/id/trust#operator">fakta operator</a> dan <a href="/id/enterprise#procurement">lembar fakta pengadaan</a>.`},
		},
	}
}

func enterpriseID() enterpriseCopy {
	return enterpriseCopy{
		Title:        "Netsekurity: Pentest Menyeluruh untuk Aplikasi Anda",
		Description:  "Uji area setelah login, hak akses pengguna, dan alur bisnis bersama engineer keamanan. Sepakati cakupan aplikasi dan dapatkan temuan beserta panduan perbaikan.",
		Eyebrow:      "Pentest menyeluruh",
		H1:           "Periksa lebih dalam bagaimana aplikasi Anda bisa disalahgunakan.",
		Lead:         "Untuk menguji fitur setelah login, hak akses antar pengguna, atau alur bisnis, bekerja samalah dengan engineer keamanan. Tinjauan kode sumber dan analisis berbantuan agen mendukung pengujian dengan cakupan sesuai aplikasi Anda.",
		HeroCTA:      "Diskusikan aplikasi saya",
		Jump:         []jump{{"Proses", "board"}, {"Layanan", "escalation"}, {"Untuk developer", "engineering"}, {"Pengadaan", "procurement"}, {"Kontak", "contact"}},
		BoardH2:      "Apa yang terjadi setelah Anda menghubungi kami?",
		BoardIntro:   "Ceritakan fungsi aplikasi dan hal yang ingin Anda ketahui. Developer Anda bisa membantu menyiapkan kebutuhan teknisnya.",
		Board:        []kv{{"Diskusikan aplikasinya", "Jelaskan website, alur yang penting, dan alasan Anda membutuhkan pengujian."}, {"Sepakati cakupan dan akses", "Pastikan aplikasi, akun, dan alur yang diuji, beserta harga, jadwal, dan izin pengujiannya."}, {"Telusuri bersama engineer", "Engineer keamanan khusus melakukan tinjauan kode dan pengujian terautentikasi yang disepakati, dengan analisis berbantuan agen."}, {"Rencanakan perbaikan", "Gunakan laporan untuk memahami temuan dan membahas tindakan bersama developer. Pastikan kebutuhan pengujian lanjutan tercantum dalam cakupan pekerjaan."}},
		BoardCallout: "Cakupan disepakati sebelum pekerjaan dimulai. Laporan pemeriksaan eksternal yang sudah ada dapat membantu diskusi; Anda juga bisa langsung meminta pentest menyeluruh.",
		EngH2:        "Detail teknis untuk developer dan engineering lead",
		EngIntro:     "Rincian berikut menjelaskan pemeriksaan eksternal standar. Pentest menyeluruh memiliki cakupan tersendiri yang disepakati, termasuk akses dan akun uji yang dibutuhkan.",
		Eng: []engBlock{
			{Title: "Apa yang berjalan", Paragraphs: []string{"Lebih dari 35 pemeriksaan read-only: daftar persisnya ada di bawah “Persis apa yang diperiksa” pada halaman ini, dan di halaman cakupan produk. Subdomain didaftar, bukan dinilai: hanya domain yang Anda verifikasi yang dipindai, dan setiap hostname yang ingin dinilai adalah domain tersendiri dengan kreditnya sendiri."}},
			{Title: "Bagaimana menyentuh sistem Anda", Paragraphs: []string{"Permintaan HTTP GET biasa, pencarian DNS, dan handshake TLS. Setiap permintaan HTTP membawa User-Agent “NetSekurity Scanner”, sehingga bisa dikenali di log Anda, dan probe paparan serta redirect diberi jeda 0,3 detik. Tanpa kredensial, tanpa payload injeksi, tidak ada yang ditulis atau diunggah. Probe paparan mencatat kode status HTTP dan ukuran respons lalu membuang isinya; pengecualiannya robots.txt, yang path terlarangnya diringkas sebagai bukti."}},
			{Title: "Apa yang dikembalikan", Paragraphs: []string{"Laporan PDF berbahasa Inggris untuk setiap asesmen, dapat diunduh dari dasbor. Setiap temuan memiliki tingkat keparahan, skor CVSS v3.1 bila berlaku, referensi CWE dan CVE, kategori OWASP, bukti sebagai bukti konsep (PoC), dampak, dan perbaikan. API CI/CD memulai asesmen dan melaporkan statusnya beserta jumlah temuan per tingkat keparahan; PDF-nya diunduh dari dasbor."}},
			{Title: "Menyambungkan ke CI/CD", Paragraphs: []string{"Buat token API di dasbor. Token ditampilkan sekali, disimpan hanya sebagai hash SHA-256, dan kedaluwarsa setelah 7 sampai 90 hari sesuai pilihan Anda. Satu asesmen berjalan pada satu waktu per akun. Berfungsi dengan sistem apa pun yang bisa mengirim permintaan HTTP; dokumentasi memuat contoh GitHub Actions, GitLab CI, dan Jenkins."},
				Code: apiExample},
			{Title: "False positive", Paragraphs: []string{"Sebuah probe menghitung path yang menjawab 200, 301, atau 302, sehingga aplikasi satu halaman atau pengalihan login bisa tampak seperti paparan. Agen analisis AI menghapus duplikat dan mengurutkan temuan, tetapi tidak menggantikan seorang engineer. Meminta engineer menguji ulang setiap temuan adalah tambahan tinjauan ahli."}},
		},
		ProcH2:    "Informasi perusahaan untuk direksi dan pengadaan",
		ProcIntro: "Semua di bawah ini benar saat ini. Bila proses vendor Anda memerlukan sesuatu yang belum kami cantumkan, beri tahu kami apa yang dibutuhkan.",
		ProcFacts: []kv{
			{"Pemasok", "Dalang Pte. Ltd., Singapura. UEN 202622331M. 160 Robinson Road, #14-04, Singapura 068914."},
			{"Produk", "Netsekurity: asesmen keamanan web eksternal otomatis (blackbox, tanpa autentikasi, read-only). Bukan uji penetrasi."},
			{"Satuan pembelian", "1 kredit = 1 asesmen eksternal untuk 1 domain terverifikasi."},
			{"Harga (USD)", "{packs}. Tinjauan ahli: harga sesuai permintaan. Whitebox: {whitebox} per aplikasi atau domain."},
			{"Pembayaran", "Online, dalam dolar AS, melalui Xendit. Netsekurity tidak pernah melihat atau menyimpan data kartu. Perlakuan pajak dan rincian faktur: tanyakan ke sales@dalang.io sebelum membeli."},
			{"Kredit dan pengembalian", "Kredit tidak kedaluwarsa. Pemindaian yang gagal dikembalikan otomatis."},
			{"Deliverable", "Laporan PDF berbahasa Inggris untuk setiap domain, dapat diunduh oleh akun yang memesannya."},
			{"Kontrol kepemilikan", "Konfirmasi kendali setiap hostname melalui DNS TXT, file HTTPS, atau meta tag halaman utama. Bukti diperiksa kembali sebelum setiap pemindaian mandiri; izin pengujian dikonfirmasi secara terpisah."},
			{"Akun", "Masuk dengan Google. Saat ini akun bersifat perorangan, dan kredit serta laporan milik akun yang membelinya."},
			{"Pihak ketiga", "Google (masuk), Xendit (pembayaran), Cloudflare (jaringan), Umami (analitik), Meta (hanya setelah persetujuan), dan penyedia model AI. Halaman kepercayaan merinci apa yang diterima masing-masing."},
			{"Kebijakan", "Dalang menerbitkan kebijakan privasi (dalang.io/privacy), ketentuan penggunaan (dalang.io/terms-of-use), serta pusat keamanan dan kepercayaan (dalang.io/security)."},
			{"Dukungan dan keamanan", "sales@dalang.io, dibalas dalam 1 hari kerja. Masalah keamanan: security@dalang.io. Lihat halaman kepercayaan untuk kebijakan pengungkapan."},
			{"Kontrak dan administrasi", "Butuh purchase order, ketentuan kontrak tertentu, atau kuesioner vendor diisi? Email sales@dalang.io dan beri tahu apa yang dibutuhkan proses Anda."},
		},
		ProcNote:    "Halaman ini tidak mengklaim sertifikasi atau audit independen untuk Netsekurity itu sendiri.",
		ProfileLead: "Fakta perusahaan di halaman ini bersumber dari Company Profile Dalang 2026.", ProfileLink: "Buka Company Profile (PDF)",
		EscLink: "Perbandingan tingkat ini dengan uji penetrasi",
		FAQH2:   "Pertanyaan enterprise",
		FAQ: []qa{
			{"Bisakah asesmen dijalankan berjadwal?", `Netsekurity tidak memiliki penjadwal bawaan. Panggil API dari pipeline CI/CD Anda pada setiap rilis, atau dari penjadwal apa pun yang bisa mengirim permintaan HTTP.`},
			{"Bisakah pipeline menggagalkan build berdasarkan temuan?", `Bisa, dari pipeline: baca jumlah temuan dengan <code>GET /api/v1/assessments/{id}</code> dan gagalkan job jika ada High atau Critical. Sebuah pemindaian memerlukan waktu 10 menit hingga 24 jam, jadi mulai saat deploy dan periksa dari job berikutnya. <a href="/id/dokumentasi#gate">Dokumentasi</a> memuat contoh yang berfungsi.`},
			{"Apakah laporannya layak sebagai bukti audit?", `Laporan adalah catatan bertanggal tentang apa yang ditemukan asesmen eksternal read-only pada sebuah domain, dengan cakupan yang dinyatakan. Apakah itu memenuhi kontrol tertentu, auditor Anda yang memutuskan, dan laporan ini tidak menggantikan pengujian area yang memerlukan autentikasi.`},
			{"Bisakah menilai banyak domain?", `Bisa. Setiap domain diverifikasi dan dinilai terpisah dan memakai satu kredit. Paket lebih besar menurunkan harga per asesmen; untuk volume atau ketentuan pembelian, <a href="#contact">hubungi kami</a>.`},
			{"Siapa yang bisa melihat laporan kami?", `Akun yang memesan asesmen, dan operator Netsekurity yang mengelola layanan. Laporan hanya diberikan kepada pengguna yang masuk dan tidak pernah di-cache. Lihat <a href="/id/trust#access">siapa yang bisa melihat laporan</a>.`},
			{"Bagaimana jika kami butuh orang untuk meninjau temuan?", `Pesan tambahan tinjauan ahli: seorang engineer keamanan menguji ulang setiap temuan dalam satu laporan yang sudah selesai, menghapus false positive, dan menulis rencana perbaikan untuk stack Anda. Harganya sesuai permintaan.`},
		},
	}
}

func trustID() trustCopy {
	return trustCopy{
		Title:       "Kepercayaan dan Keamanan: Cara Netsekurity Memindai dan Menangani Data",
		Description: "Cara Netsekurity memverifikasi kepemilikan, menjaga pemindaian read-only, meminimalkan data, dan membatasi akses laporan. Dioperasikan Dalang Pte. Ltd.",
		Eyebrow:     "Kepercayaan",
		H1:          "Cara Netsekurity memindai, apa yang disimpan, dan siapa yang bisa melihatnya.",
		Lead:        "Pernyataan di halaman ini menggambarkan cara layanan bekerja saat ini. Masing-masing bisa dicocokkan dengan cakupan produk dan contoh laporan.",
		Jump: []jump{{"Pemindaian", "responsible"}, {"Alur data", "flow"}, {"Yang kami simpan", "data"}, {"Pihak ketiga", "third-parties"},
			{"Siapa yang bisa melihat laporan", "access"}, {"Pengungkapan", "disclosure"}, {"Operator", "operator"}},
		RespH2:    "Pemindaian yang bertanggung jawab",
		RespIntro: "Alat keamanan berbahaya jika bisa diarahkan ke siapa saja. Kontrol berikut mencegahnya.",
		Resp: []kv{
			{"Kepemilikan lebih dulu", "Pilih DNS TXT, file verifikasi HTTPS, atau meta tag halaman utama. Bukti terikat pada akun dan hostname, lalu diperiksa kembali sebelum setiap pemindaian mandiri. Verifikasi HTTPS memakai User-Agent NetSekurity Verification dan menolak redirect. Konfirmasi izin pengujian sebelum mulai."},
			{"Read-only", "Permintaan berupa HTTP GET biasa, pencarian DNS, dan handshake TLS. Asesmen standar tidak mengirim payload injeksi dan tidak menulis apa pun ke sistem Anda."},
			{"Daftar pemeriksaan yang tetap", "Pemeriksaannya adalah daftar di halaman cakupan produk, sama untuk semua pelanggan. Satu asesmen berjalan pada satu waktu per akun."},
			{"Dapat dikenali", "Setiap permintaan HTTP membawa User-Agent “NetSekurity Scanner”, sehingga bisa dikenali di log Anda."},
			{"Diberi jeda", "Probe paparan dan redirect diberi jeda 0,3 detik."},
			{"Kegagalan dikembalikan", "Jika pemindaian tidak bisa selesai, pemindaian ditutup dan kreditnya dikembalikan otomatis."},
		},
		FlowH2:   "Ke mana data Anda mengalir",
		FlowLead: "Lima langkah dari situs web Anda sampai laporan Anda. Nomornya sesuai dengan diagram.",
		Flow: []kv{
			{"Situs web Anda", "Pemindai mengirim permintaan read-only ke domain yang Anda verifikasi. Subdomain yang ditemukan didaftar, tidak dipindai."},
			{"Pemindai", "Menjalankan daftar pemeriksaan yang tetap. Untuk probe paparan, ia mencatat kode status dan ukuran lalu membuang isi respons."},
			{"Agen analisis AI", "Menjalankan asesmen, menghapus duplikat, mengurutkan temuan, dan menulis laporan. Tidak ada manusia yang membaca laporan standar sebelum dikirim."},
			{"Laporan", "PDF berbahasa Inggris, disimpan untuk akun yang memesannya. Hanya diberikan kepada akun itu dan operator Netsekurity, tidak pernah di-cache, dan dijauhkan dari mesin pencari."},
			{"Anda", "Anda mengunduh laporan dari dasbor. Pihak ketiga yang menangani data di sepanjang jalan terdaftar di bawah."},
		},
		DataH2:    "Yang kami simpan",
		DataIntro: "Kami menyimpan apa yang dibutuhkan layanan untuk berjalan dan menampilkan hasil kepada Anda, tidak lebih.",
		DataCols:  []string{"Item", "Yang disimpan"},
		Data: []dataRow{
			{"Akun", "Nama dan alamat email Anda dari login Google."},
			{"Domain", "Domain yang Anda tambahkan beserta token verifikasinya."},
			{"Asesmen", "Temuan dengan string bukti singkat (kode status, nilai header, ukuran) dan PDF laporan."},
			{"Probe paparan", "Kode status HTTP dan ukuran respons. Isi respons dibuang."},
			{"Analisis AI", "Keluaran pemindai diproses oleh agen analisis AI untuk menghasilkan laporan. Penyedia model tercantum di bawah pihak ketiga."},
			{"Pembayaran", "Catatan faktur: paket, jumlah, dan status. Data kartu ditangani Xendit dan tidak pernah sampai ke kami."},
			{"Pesan kontak", "Apa yang Anda tulis, alamat email Anda, dan alamat IP asalnya, dipakai untuk pengendalian spam."},
			{"Analitik", "Umami, yang tanpa cookie. Pixel Meta (Facebook) hanya dimuat jika Anda menerimanya di banner cookie."},
		},
		SubH2:    "Pihak ketiga yang menangani data",
		SubIntro: "Berikut layanan luar yang terlibat. Jika ada yang Anda butuhkan tetapi tidak tercantum, atau Anda memerlukan rincian yang tidak diberikan di sini, tanyakan ke security@dalang.io.",
		SubCols:  []string{"Penyedia", "Dipakai untuk", "Yang diterima"},
		Subs: []subRow{
			{"Google", "Masuk", "Nama dan alamat email Anda sampai ke kami dari Google; Google mengetahui bahwa Anda masuk ke Netsekurity."},
			{"Xendit", "Pembayaran kartu dan bank", "Data pembayaran yang Anda masukkan di halaman Xendit. Kami menerima status faktur, tidak pernah data kartu."},
			{"Cloudflare", "Proxy jaringan di depan netsekurity.com", "Permintaan yang masuk ke situs ini, sebagaimana dilihat reverse proxy mana pun."},
			{"Umami (umami.dalang.io)", "Analitik kunjungan halaman tanpa cookie", "Kunjungan halaman. Tidak ada cookie yang dipasang."},
			{"Meta", "Pengukuran iklan, hanya jika Anda menerima banner cookie", "Tidak ada, sampai Anda menerima. Setelahnya, peristiwa kunjungan halaman dan pendaftaran."},
			{"Penyedia model AI", "Agen analisis yang menulis laporan", "Keluaran pemindai untuk asesmen Anda. Penyedia dan wilayahnya tidak dipublikasikan di situs ini; tanyakan ke security@dalang.io sebelum membeli."},
		},
		AccessH2: "Siapa yang bisa melihat laporan",
		Access: []kv{
			{"Laporan", "Hanya akun yang memesan asesmen dan operator Netsekurity yang mengelola layanan. Laporan hanya diberikan kepada pengguna yang masuk, tidak pernah di-cache oleh peramban atau CDN, dan dikecualikan dari mesin pencari."},
			{"Sesi", "Cookie bertanda tangan yang berlaku 24 jam, bertanda HttpOnly, Secure, dan SameSite=Lax."},
			{"Token API", "Ditampilkan sekali, disimpan hanya sebagai hash SHA-256, dan kedaluwarsa setelah 7 sampai 90 hari sesuai pilihan Anda."},
		},
		DiscH2:   "Melaporkan kerentanan pada sistem kami sendiri",
		DiscBody: "Jika Anda menemukan masalah keamanan di netsekurity.com atau dalang.io, email security@dalang.io.",
		Disc: []string{
			"Kami mengakui laporan dalam 2 hari kerja.",
			"Kami tidak menuntut peneliti yang beritikad baik.",
			"Kami mencantumkan kredit atas perbaikan bila diminta.",
			"Mohon jangan menguji domain pelanggan lain.",
		},
		OpH2: "Siapa yang mengoperasikan Netsekurity",
		OpBody: []string{
			"Netsekurity adalah produk Dalang Pte. Ltd., perusahaan yang terdaftar di Singapura.",
			"Grup Dalang (Dalang Pte. Ltd. di Singapura, PT Mashanz Sumber Terbuka dan PT Komponen Pusat Data di Indonesia) mengoperasikan jaringannya sendiri, AS154547, yang bisa diperiksa siapa pun di bgp.he.net/AS154547.",
		},
		OpFacts: []kv{
			{"Badan hukum", "Dalang Pte. Ltd., UEN 202622331M"},
			{"Alamat", "160 Robinson Road, #14-04, Singapura 068914"},
			{"Penjualan", "sales@dalang.io"},
			{"Keamanan", "security@dalang.io"},
			{"Jaringan grup", "AS154547 (catatan publik)"},
		},
		OpSource:    "Fakta entitas dan jaringan bersumber dari Company Profile Dalang 2026, per 20 Oktober 2026.",
		ProfileLead: "Company Profile adalah dokumen Dalang sendiri tentang grupnya.", ProfileLink: "Buka Company Profile (PDF)",
		PolicyLead: "Dalang menerbitkan kebijakannya di dalang.io.",
		LimitsH2:   "Yang tidak diklaim halaman ini",
		LimitsBody: "Halaman ini tidak mengklaim sertifikasi atau audit independen untuk Netsekurity itu sendiri, dan tidak menyatakan bahwa sebuah asesmen membuat situs aman. Halaman ini menyatakan bagaimana layanan berperilaku. Jika ada yang Anda butuhkan tetapi tidak tercantum, tanyakan kepada kami.",
	}
}

func compareID() compareCopy {
	return compareCopy{
		Title:       "Asesmen Eksternal vs Uji Penetrasi: Apa Bedanya?",
		Description: "Asesmen web eksternal bersifat otomatis dan berulang; uji penetrasi adalah upaya manual membobol sistem. Apa yang ditemukan dan terlewat, dan cara memilih.",
		Eyebrow:     "Panduan",
		H1:          "Asesmen keamanan web eksternal vs uji penetrasi: apa bedanya?",
		Lead:        "Asesmen keamanan web eksternal adalah pemeriksaan otomatis dan berulang atas apa yang diperlihatkan sebuah situs web kepada internet publik. Uji penetrasi adalah pekerjaan manual ketika seseorang mencoba membobol aplikasi, biasanya dengan kredensial, untuk menunjukkan apa yang sebenarnya bisa dilakukan penyerang. Keduanya menjawab pertanyaan yang berbeda dan paling baik bila digunakan bersama.",
		Updated:     "Terakhir diperbarui 26 Oktober 2026.",
		TableH2:     "Dua pendekatan, berdampingan",
		Cols:        []string{"Aspek", "Asesmen eksternal", "Uji penetrasi"},
		Rows: []cmpRow{
			{"Pertanyaan yang dijawab", []string{"Apa yang diperlihatkan situs ini kepada internet publik saat ini?", "Apa yang sebenarnya bisa dilakukan penyerang terhadap aplikasi ini?"}},
			{"Siapa atau apa yang melakukannya", []string{"Pemindai otomatis dengan pemilahan oleh AI; tinjauan engineer bersifat opsional.", "Seorang engineer keamanan, biasanya dibantu alat."}},
			{"Tingkat akses", []string{"Eksternal dan tanpa autentikasi: tanpa kredensial.", "Sering terautentikasi, kadang dengan kode sumber (whitebox)."}},
			{"Keluaran khas", []string{"Daftar paparan dan salah konfigurasi yang terprioritas, lengkap dengan bukti dan perbaikan.", "Jalur eksploit yang tervalidasi, termasuk celah logika bisnis dan otorisasi."}},
			{"Frekuensi yang cocok", []string{"Setiap rilis, atau berkala.", "Sebelum peluncuran besar, setelah perubahan besar, atau secara periodik."}},
			{"Yang bisa terlewat", []string{"Apa pun di balik login, otorisasi, logika bisnis, celah injeksi.", "Apa pun yang berada di luar lingkup atau waktu yang disepakati."}},
			{"Penawaran Netsekurity", []string{"Asesmen eksternal otomatis, mulai {single} per domain.", "Asesmen whitebox, {whitebox} per aplikasi atau domain; atau uji penetrasi manual dari penyedia lain."}},
		},
		GoodH2:   "Yang dikuasai asesmen eksternal",
		CannotH2: "Yang tidak bisa diberitahukannya",
		ChooseH2: "Cara memilih",
		Good: []string{
			"Rahasia dan file yang terbuka seperti .env, .git, dan arsip cadangan",
			"Header keamanan yang hilang dan perlindungan clickjacking yang lemah",
			"Sertifikat TLS yang akan kedaluwarsa atau bermasalah",
			"Open redirect yang membuat tautan phishing tampak sah",
			"Host yang terlupakan, melalui daftar subdomain pasif",
			"Regresi: pemeriksaan yang sama pada setiap rilis memperlihatkan apa yang berubah",
		},
		Cannot: []string{
			"Apakah pengguna yang login bisa melihat data pengguna lain (otorisasi, IDOR)",
			"Celah pada logika bisnis, pembayaran, dan alur kerja",
			"Bug input yang bisa dieksploitasi seperti SQL injection dan cross-site scripting",
			"Apa pun pada host selain yang Anda verifikasi",
			"Apakah sebuah situs “aman”: laporan bersih berarti “tidak ditemukan oleh pemeriksaan ini”",
		},
		Choose: []kv{
			{"Anda perlu tahu apa yang terpapar hari ini, dan terus memeriksa", "Gunakan asesmen eksternal otomatis, dan jalankan dari CI/CD."},
			{"Anda perlu seseorang memastikan temuan dan menghapus false positive", "Tambahkan tinjauan ahli pada laporan yang sudah selesai."},
			{"Anda memiliki area login, pembayaran, atau data sensitif", "Pesan asesmen whitebox atau uji penetrasi manual, dan biarkan asesmen eksternal tetap berjalan di antara pekerjaan tersebut."},
		},
		GlossH2: "Istilah yang dijelaskan",
		Gloss: []kv{
			{"Permukaan serangan", "Segala sesuatu pada sebuah sistem yang bisa dijangkau dari luar: situs web, subdomain, file, layanan."},
			{"Pengujian eksternal (blackbox)", "Pengujian dari internet publik tanpa pengetahuan internal, seperti penyerang yang tidak punya akses."},
			{"Tanpa autentikasi", "Tanpa login. Tidak ada yang berada di balik login yang diuji."},
			{"Whitebox", "Pengujian dengan pengetahuan internal, seperti kode sumber dan akun uji."},
			{"CVSS", "Common Vulnerability Scoring System: skor 0 sampai 10 untuk tingkat keparahan sebuah kerentanan."},
			{"CWE", "Common Weakness Enumeration: katalog publik jenis kelemahan perangkat lunak, misalnya CWE-601 untuk open redirect."},
			{"OWASP Top 10", "Daftar yang banyak dipakai tentang kategori risiko aplikasi web paling kritis. Memetakan temuan ke daftar ini berarti mengategorikannya; bukan berarti setiap kategori diuji."},
			{"Record DNS TXT", "Entri teks pada pengaturan DNS sebuah domain. Netsekurity memakainya untuk memastikan Anda mengendalikan domain tersebut."},
			{"False positive", "Masalah yang dilaporkan tetapi bukan masalah sebenarnya."},
		},
		FAQH2: "Pertanyaan umum",
		FAQ: []qa{
			{"Apakah asesmen eksternal cukup untuk kepatuhan?", `Asesmen ini bisa memberi bukti tentang apa yang diperlihatkan sebuah situs, tetapi skema kepatuhan umumnya meminta lebih dari sekadar pemindaian eksternal. Periksa apa yang dipersyaratkan auditor Anda, dan perlakukan asesmen sebagai salah satu masukan.`},
			{"Seberapa sering asesmen eksternal harus dijalankan?", `Sesering situs berubah. Dengan integrasi CI/CD, asesmen bisa berjalan pada setiap rilis; tanpa itu, secara berkala sesuai seberapa sering Anda melakukan deploy.`},
			{"Apakah laporan bersih berarti situs aman?", `Tidak. Artinya pemeriksaan ini tidak menemukan apa pun. Area di luar cakupan, seperti fungsi terautentikasi dan logika bisnis, tidak diperiksa.`},
		},
		CTA: "Lihat contoh laporan", CTALink: "/id/contoh-laporan",
	}
}

func sampleID() sampleCopy {
	return sampleCopy{
		Title:       "Contoh Laporan Asesmen Keamanan: Yang Didapat dari Satu Kredit",
		Description: "Contoh laporan asesmen keamanan web eksternal Netsekurity: kontrol dokumen, ringkasan temuan, register risiko, bukti, dan rencana perbaikan 30/60/90 hari.",
		Eyebrow:     "Contoh · domain fiktif, struktur laporan sebenarnya",
		H1:          "Yang Anda dapat dari satu kredit",
		Lead:        "Setiap temuan di bawah berasal dari pemeriksaan yang benar-benar dijalankan pemindai. Domain dan buktinya fiktif. Laporan Anda berupa PDF berbahasa Inggris dengan struktur ini, dapat diunduh dari dasbor; PDF lengkapnya juga memuat metodologi dan lampiran bukti.",
		CTA:         "Dapatkan untuk domain Anda · {single}", CTAAll: "Semua paket", PrintLabel: "Cetak atau simpan sebagai PDF", PDFLabel: "Unduh contoh sebagai PDF",
		DocH2: "1 · Kontrol dokumen",
		Doc: []kv{
			{"Target", "shop.example.com"},
			{"Jenis asesmen", "Eksternal · tanpa autentikasi · read-only"},
			{"Kepemilikan", "Diverifikasi lewat TXT _netsekurity"},
			{"Selesai", "2026-09-01 09:14 UTC"},
			{"Disusun oleh", "Netsekurity.com Security Agent"},
			{"Klasifikasi", "Rahasia: hanya pemilik dan operator"},
		},
		SumH2:   "2 · Ringkasan temuan",
		Badges:  []sevBadge{{"1 High", "bad"}, {"2 Medium", "warn"}, {"2 Low", "blue"}, {"4 Info", "muted"}},
		Posture: "Postur keseluruhan: perlu perhatian.", PostureBody: "Satu file yang terbuka bisa membocorkan kredensial hari ini; sisanya adalah penguatan.",
		RiskH2:   "3 · Register risiko",
		RiskCols: []string{"#", "Temuan", "Tingkat keparahan", "CVSS", "CWE", "OWASP"},
		Risks: []riskRow{
			{"F-01", "File environment terbuka di /.env", "bad", "High", "7.5", "CWE-538", "A05 · A02"},
			{"F-02", "Header keamanan hilang (4 dari 6)", "warn", "Medium", "5.3", "CWE-1021", "A05"},
			{"F-03", "Open redirect melalui ?next=", "warn", "Medium", "5.4", "CWE-601", "A01"},
			{"F-04", "Versi server terungkap (nginx/1.18.0)", "blue", "Low", "3.7", "CWE-200", "A05"},
			{"F-05", "Sertifikat TLS kedaluwarsa dalam 41 hari", "blue", "Low", "3.7", "CWE-295", "A02"},
			{"I-01…04", "Resolusi DNS · 7 subdomain · WAF terdeteksi · entri disallow di robots.txt", "muted", "Info", "—", "—", "—"},
		},
		FindH2: "4 · Temuan dan analisis",
		Findings: []finding{
			{"F-01 · File environment terbuka di /.env", "High · CVSS 7.5", "bad", []frow{
				{"Yang kami lihat", "Path /.env menjawab HTTP 200 dengan isi 1,2 KB. File dengan nama ini biasanya berisi kata sandi database, kunci API, dan rahasia penandatanganan. Kami tidak menyimpan isinya.", false},
				{"Bukti", "GET /.env -&gt; 200 1247", true},
				{"Dampak", "Siapa pun bisa membaca rahasia produksi Anda.", false},
				{"Perbaikan", `Blokir file itu di web server (nginx: <span class="nsk-mono">location ~ /\.env { deny all; }</span>), pindahkan ke luar web root, lalu ganti setiap rahasia di dalamnya: anggap sudah terbaca.`, false},
			}},
			{"F-02 · Header keamanan hilang", "Medium · CVSS 5.3", "warn", []frow{
				{"Yang kami lihat", "Respons tidak memiliki strict-transport-security, content-security-policy, referrer-policy, dan permissions-policy. x-frame-options dan x-content-type-options ada.", false},
				{"Bukti", "missing=strict-transport-security, content-security-policy, referrer-policy, permissions-policy", true},
				{"Perbaikan", `Tambahkan keempat header di proxy atau framework. Mulai HSTS dengan max-age=300, naikkan menjadi setahun setelah tidak ada yang rusak. CSP <span class="nsk-mono">default-src 'self'</span> ditambah host CDN Anda adalah versi pertama yang aman.`, false},
			}},
			{"F-03 · Open redirect melalui ?next=", "Medium · CVSS 5.4", "warn", []frow{
				{"Yang kami lihat", "Permintaan /?next=https://evil.example dijawab dengan 302 ke alamat eksternal itu.", false},
				{"Bukti", "GET /?next=https://evil.example -&gt; 302 https://evil.example", true},
				{"Dampak", "Tautan phishing bisa membawa domain asli Anda dan tetap mendarat di halaman penyerang.", false},
				{"Perbaikan", "Hanya terima path relatif yang diawali satu /, atau periksa tujuannya terhadap daftar izin host milik Anda sendiri.", false},
			}},
		},
		MoreSummary: "F-04, F-05, dan empat butir informasional",
		More: []template.HTML{
			`<strong class="font-extrabold">F-04 · Versi server terungkap.</strong> <span class="nsk-mono">server: nginx/1.18.0</span> memberi tahu penyerang CVE mana yang harus dicoba. Setel <span class="nsk-mono">server_tokens off;</span>.`,
			`<strong class="font-extrabold">F-05 · Sertifikat kedaluwarsa dalam 41 hari.</strong> Penerbit Let's Encrypt, 3 entri SAN. Pastikan perpanjangan otomatis berjalan; sertifikat yang kedaluwarsa membuat situs tidak bisa diakses.`,
			`<strong class="font-extrabold">I-01 · DNS.</strong> Apex mengarah ke 203.0.113.10. <strong class="font-extrabold">I-02 · Subdomain.</strong> 7 host ditemukan lewat certificate transparency, termasuk staging.shop.example.com: pastikan memang dimaksudkan publik. <strong class="font-extrabold">I-03 · WAF.</strong> Sidik jari Cloudflare ada. <strong class="font-extrabold">I-04 · robots.txt.</strong> Melarang /admin/, yang menegaskan bahwa path itu ada.`,
		},
		PlanH2: "5 · Rencana perbaikan (30 / 60 / 90)",
		Plan: []kv{
			{"Hari ini", "Blokir /.env. Ganti setiap rahasia di dalamnya."},
			{"Bulan ini", "Perbaiki redirect. Tambahkan HSTS, CSP, Referrer-Policy, dan Permissions-Policy."},
			{"Kuartal depan", "Sembunyikan versi server. Putuskan apakah staging. memang harus publik. Pindai ulang (1 kredit) untuk memastikan."},
		},
		LimitsH2:   "6 · Yang tidak dicakup laporan ini",
		LimitsBody: `Tidak ada yang berada di balik login, tidak ada pengujian otorisasi (IDOR) atau logika bisnis, dan tidak ada payload injeksi (SQLi, XSS, LFI). Itu memerlukan tingkat whitebox. Temuan yang tidak muncul berarti “tidak dipindai”, bukan “terbukti aman”. Daftar lengkapnya ada di halaman <a class="font-bold underline" href="/id/cakupan-produk">cakupan produk</a>.`,
	}
}

func docsID() docsCopy {
	return docsCopy{
		Title:       "Integrasi CI/CD dan Referensi API | Netsekurity",
		Description: "Mulai asesmen Netsekurity dari pipeline Anda, baca statusnya, dan tahan rilis berdasarkan temuan: panduan cepat, referensi API, dan contoh untuk {ci} sistem CI.",
		Eyebrow:     "Dokumentasi",
		H1:          "Integrasi CI/CD dan referensi API",
		Lead:        "Mulai asesmen keamanan web eksternal dari pipeline Anda pada setiap deploy, lalu baca statusnya dan jumlah temuannya. Dua permintaan HTTP, diautentikasi dengan token API.",
		Jump:        []jump{{"Panduan cepat", "quickstart"}, {"Referensi API", "api"}, {"Tahan rilis", "gate"}, {"Contoh CI", "ci"}, {"Catatan keamanan", "notes"}},
		QuickH2:     "Panduan cepat",
		Quick: []template.HTML{
			`<strong>Tambahkan dan verifikasi domain Anda.</strong> Di dasbor (yang berbahasa Inggris): My websites → tambahkan domain → pilih DNS TXT, file HTTPS, atau meta tag halaman utama → periksa. Domain harus tampil sebagai <em>Verified</em>.`,
			`<strong>Buat token API.</strong> Di dasbor: Automation → pilih masa berlakunya (7, 14, 30, 60, atau 90 hari) → buat. Salin sekarang: token hanya ditampilkan sekali.`,
			`<strong>Simpan sebagai secret</strong> di CI Anda, misalnya <code>NETSEKURITY_API_TOKEN</code>.`,
			`<strong>Panggil endpoint</strong> saat deploy. Netsekurity mengantrekan dan menjalankan asesmen; laporan muncul di dasbor Anda, dan statusnya bisa dibaca dari API.`,
		},
		QuickNote: "1 kredit per asesmen. Eksternal dan read-only, jadi aman diarahkan ke produksi.",
		APIH2:     "Referensi API HTTP",
		StartH3:   "POST /api/v1/assessments",
		StartLead: "Mulai sebuah asesmen. Nama aslinya, /api/v1/pentests, tetap berfungsi dan berperilaku sama.",
		ParamCols: []string{"Parameter", "Tipe", "Deskripsi"},
		Params: []paramRow{
			{"authorised", "boolean JSON / form true", "Wajib: true menyatakan izin untuk menguji hostname ini sesuai cakupan eksternal read-only yang dipublikasikan. Dikirim setiap kali mulai; bukti verifikasi diperiksa lagi sebelum kredit digunakan."},
			{"X-API-Token", "header", "Wajib. Token API Anda."},
			{"domain", "form atau JSON", "Wajib. Domain terverifikasi milik akun pemilik token."},
			{"mode", "form atau JSON", `Opsional. <code class="nsk-mono">standard</code> adalah satu-satunya nilai: 1 kredit, read-only.`},
		},
		StartResp: "# 200 OK\n" +
			"{\"assessment_id\":\"pt_1f2e…\",\"pentest_id\":\"pt_1f2e…\",\"domain\":\"app.example.com\",\"mode\":\"standard\",\"status\":\"queued\"}\n" +
			"# 401 token tidak ada, tidak valid, atau kedaluwarsa\n" +
			"# 400 domain tidak ada, mode tidak valid, atau konfirmasi izin belum diberikan\n" +
			"# 402 kredit tidak cukup\n" +
			"# 409 bukti kendali hostname belum terkonfirmasi, atau sebuah asesmen sudah antre atau berjalan\n" +
			"# 429 dibatasi laju",
		ResultH3:   "GET /api/v1/assessments/{assessment_id}",
		ResultLead: "Baca status asesmen yang Anda mulai, dan jumlah temuan per tingkat keparahan setelah selesai.",
		ResultResp: "# 200 OK, saat pemindaian berjalan\n" +
			"{\"assessment_id\":\"pt_1f2e…\",\"domain\":\"app.example.com\",\"status\":\"running\",\"started_at\":\"2026-10-26T09:00:12Z\",\"completed_at\":null,\"findings\":null}\n" +
			"# 200 OK, setelah selesai dan laporan memuat rincian tingkat keparahan\n" +
			"{\"assessment_id\":\"pt_1f2e…\",\"domain\":\"app.example.com\",\"status\":\"completed\",\"started_at\":\"2026-10-26T09:00:12Z\",\"completed_at\":\"2026-10-26T09:14:40Z\",\"findings\":{\"critical\":0,\"high\":1,\"medium\":2,\"low\":2,\"info\":4}}\n" +
			"# 401 token tidak ada, tidak valid, atau kedaluwarsa\n" +
			"# 404 tidak ada asesmen seperti itu pada akun pemilik token",
		ResultParam: "assessment_id yang dikembalikan saat Anda memulainya. Hanya asesmen milik akun Anda yang terlihat: id lain dijawab 404.",
		ResultNote:  `<code>findings</code> bernilai <code>null</code> sampai laporan dikirim bersama rincian tingkat keparahan. <code>null</code> tidak berarti tidak ada temuan. PDF tidak disajikan di sini: unduh dari dasbor setelah masuk.`,
		Statuses: []kv{
			{"queued", "Diterima dan menunggu pemindai."},
			{"running", "Pemindai sedang mengerjakannya. Sebuah pemindaian memerlukan waktu 10 menit hingga 24 jam, tergantung temuannya."},
			{"completed", "Laporan sudah siap di dasbor."},
			{"failed", "Pemindaian tidak bisa selesai. Kredit dikembalikan otomatis."},
		},
		GateH3:   "Tahan rilis berdasarkan temuan",
		GateLead: "Job ini memulai asesmen, menunggunya hingga 30 menit, dan gagal bila ada temuan High atau Critical, bila pemindaian gagal, bila statusnya tidak bisa dibaca, atau bila asesmen masih berjalan saat batas tunggu habis.",
		GateCode: gateScript("tidak ada temuan High atau Critical", "tidak bisa memulai asesmen", "asesmen gagal", "asesmen masih berjalan setelah 30 menit", "tidak bisa membaca status asesmen"),
		GateNote: "Sebuah pemindaian memerlukan waktu 10 menit hingga 24 jam, jadi job yang menunggu sering akan menunggu lama. Biasanya lebih baik memulai asesmen saat deploy dan menjalankan pemeriksaan ini dari job berikutnya.",
		CIH2:     "Contoh CI",
		CILead:   "Sistem apa pun yang bisa mengirim permintaan HTTP akan berfungsi. Contoh ini memulai asesmen; tambahkan pemeriksaan di atas di tempat Anda membutuhkan hasilnya.",
		CINotes: []template.HTML{
			`Tambahkan <code>NETSEKURITY_API_TOKEN</code> di Settings → Secrets and variables → Actions.`,
			`Tambahkan <code>NETSEKURITY_API_TOKEN</code> di Settings → CI/CD → Variables (masked).`,
			`Buat kredensial <em>Secret text</em> bernama <code>nsk-api-token</code>.`,
			`Tambahkan <code>NETSEKURITY_API_TOKEN</code> di Project → Settings → Environment Variables.`,
			`Tambahkan <code>NETSEKURITY_API_TOKEN</code> di Pipelines → Library → Variable group (jadikan rahasia).`,
			`Tambahkan <code>NETSEKURITY_API_TOKEN</code> di Repository → Settings → Pipelines → Repository variables (secured).`,
			`Tambahkan secret terenkripsi dengan <code>travis encrypt NETSEKURITY_API_TOKEN=xxx</code>.`,
			`Definisikan <code>env.NSK_API_TOKEN</code> di konfigurasi build (Parameters, tipe Password).`,
			`Tambahkan <code>NETSEKURITY_API_TOKEN</code> di Pipeline → Settings → Environment variables (secret).`,
			`Setel <code>NETSEKURITY_API_TOKEN</code> di environment pipeline sebagai secret.`,
			`Setel <code>NETSEKURITY_API_TOKEN</code> di environment pipeline sebagai secret.`,
		},
		NotesH2: "Catatan konfigurasi dan keamanan",
		Notes: []template.HTML{
			`<strong>Token kedaluwarsa.</strong> Pilih masa berlaku yang sesuai dengan ritme deploy Anda, dan rotasi dengan membuat token baru lalu menghapus yang lama.`,
			`<strong>Simpan token sebagai secret CI.</strong> Jangan pernah meng-commit-nya ke source control. Token disimpan hanya sebagai hash SHA-256 di sisi kami, dan tidak pernah diterima lewat URL.`,
			`<strong>Satu token per pipeline</strong> sudah cukup. Hapus token mana pun dari dasbor kapan saja.`,
			`<strong>Satu asesmen pada satu waktu.</strong> Hanya satu asesmen yang antre atau berjalan per akun; API menjawab 409 bila sudah ada yang aktif.`,
			`<strong>Kredit.</strong> Satu asesmen satu kredit. Saldo yang tidak cukup mengembalikan HTTP 402.`,
			`<strong>Polling.</strong> Endpoint status dibatasi per alamat klien. Cek sekali per menit, bukan dalam loop rapat.`,
			`<strong>Cakupan.</strong> Eksternal, tanpa autentikasi, dan read-only. Lihat <a href="/id/cakupan-produk">cakupan produk</a> untuk apa persisnya yang diperiksa dan yang tidak.`,
		},
	}
}

func scopeID() scopeCopy {
	return scopeCopy{
		Title:       "Cakupan Produk dan Metodologi: Apa yang Diperiksa Netsekurity",
		Description: "Persis apa yang diperiksa asesmen keamanan web eksternal Netsekurity, apa yang tidak, cara temuan diberi skor, dan deliverable-nya, beserta tingkat eskalasi whitebox.",
		Crumb:       "Cakupan produk",
	}
}
