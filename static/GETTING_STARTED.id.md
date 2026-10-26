# Mulai pemeriksaan keamanan website pertama Anda

Anda bisa memesan pemeriksaan tanpa keahlian keamanan. Anda memerlukan izin untuk menguji website dan akses ke pengaturan website, hosting, atau DNS. Pengelola website bisa membantu verifikasi dan perbaikannya.

## Sebelum membeli

Layanan standar memeriksa bagian publik website tanpa login. Pemeriksaan mencari paparan eksternal dan masalah konfigurasi yang tercantum dalam [cakupan produk](/id/cakupan-produk). Fitur setelah login, hak akses antar akun, dan alur bisnis tidak diuji.

1 kredit membayar 1 pemeriksaan untuk 1 hostname, termasuk pemeriksaan ulang. Subdomain berbeda dihitung sebagai target terpisah. Lihat [harga saat ini](/id/#pricing) sebelum membeli. Menambahkan website dan memverifikasi akses tidak dipungut biaya.

## Tambahkan website

[Masuk](/login) dengan Google, buka My websites di dasbor berbahasa Inggris, lalu masukkan alamatnya. Gunakan hostname yang ingin diperiksa: `www.example.com` dan `api.example.com` adalah target terpisah.

## Konfirmasi akses

Pilih salah satu metode pada kartu website:

- **Pengaturan website:** salin meta tag yang diberikan ke bagian head halaman utama publik. Pilihan ini sesuai untuk platform yang menyediakan pengaturan tersebut.
- **File hosting:** unduh file verifikasi, lalu tempatkan di `/.well-known/netsekurity-verification.txt`. API dapat menyediakan isi yang sama melalui endpoint tersebut.
- **Pengaturan domain:** tambahkan record DNS TXT yang diberikan. Ikuti nama record yang ditampilkan dan perhatikan apakah penyedia DNS menambahkan nama domain secara otomatis.

Verifikasi file dan halaman utama memerlukan HTTPS yang valid serta respons langsung tanpa redirect atau login. Website yang mengarah ke hostname lain dapat memakai verifikasi DNS atau memverifikasi hostname tujuan.

Tekan tombol pemeriksaan untuk metode yang dipilih. Biarkan bukti terpasang: Netsekurity memeriksanya kembali sebelum setiap pemindaian mandiri. Jika bukti tidak dapat dikonfirmasi, pemindaian tidak dimulai dan kredit tidak terpakai. Periksa instruksinya lalu coba lagi; perubahan DNS mungkin memerlukan waktu untuk terlihat.

**Website dikelola orang lain?** Buka pilihan tersebut pada kartu dan salin instruksi untuk pengelola. Tanyakan metode yang digunakan, lalu konfirmasi sendiri. Jangan mengirim password atau akses masuk Netsekurity Anda.

## Mulai pemeriksaan

Beli kredit saat siap. Tinjau cakupan eksternal read-only dan konfirmasi bahwa Anda berwenang menguji hostname ini. Pemindaian memerlukan waktu 10 menit hingga 24 jam, tergantung temuannya. Satu pemindaian dapat berjalan pada satu waktu per akun. Anda boleh meninggalkan dasbor selama proses berjalan.

## Baca hasil dan rencanakan perbaikan

Buka Security scans. Ringkasan hasil menjelaskan tindakan berikutnya dalam bahasa Inggris dan Indonesia; laporan yang dapat diunduh berupa PDF berbahasa Inggris. Berikan laporan secara aman kepada developer untuk melihat bukti dan panduan perbaikannya. Tautan unduh membutuhkan akun pemesan atau operator berwenang; meneruskan alamat dasbor tidak memberikan akses kepada orang lain.

Layanan standar memberikan temuan dan panduan. Developer Anda melakukan perbaikan. Setelah perubahan, jalankan pemeriksaan baru dengan 1 kredit dan tinjau laporannya.

Tidak adanya temuan bukan berarti setiap bagian aplikasi aman. Pemeriksaan gagal bukan hasil bersih; kredit pemindaian mandiri yang gagal dikembalikan. Jika jumlah temuan tidak tersedia, baca laporan dan jangan menganggap angka yang tidak tersedia sebagai nol.

[Lihat contoh hasil](/id/contoh-laporan), [diskusikan pentest menyeluruh](/id/enterprise), atau [baca panduan integrasi developer](/id/dokumentasi).
