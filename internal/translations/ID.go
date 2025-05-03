package translations

// ID contains all Indonesian translations for the chatbot
var ID = map[string]string{
	// General messages
	"language_changed": "Bahasa bot telah diubah.",
	"current_language": "Bahasa bot saat ini: *Indonesia*. Ubah dengan /lang indonesia atau /lang english.",
	"unknown_language": "Pilihan bahasa tidak dikenali. Gunakan /lang indonesia atau /lang english.",

	// Bill management
	"bill_exists":         "Sudah ada tagihan aktif di chat ini. Harap tutup terlebih dahulu sebelum membuat yang baru.",
	"bill_created":        "Membuat tagihan baru: *%s*\nBagi yang ingin berpartisipasi, silakan ketik /join",
	"no_bill":             "Tidak ada tagihan aktif di chat ini. Buat terlebih dahulu dengan /new.",
	"bill_name_set":       "Nama tagihan diubah menjadi *%s*.",
	"user_joined":         "*%s* bergabung dengan tagihan *%s*.",
	"user_already_joined": "*%s* sudah menjadi peserta dalam tagihan *%s*.",
	"bill_closed":         "Tagihan *%s* telah ditutup.",

	// Item management
	"item_added":     "Menambahkan item: *%s* dengan harga %s ke tagihan *%s*",
	"invalid_amount": "Jumlah tidak valid. Contoh: /add Nasi Goreng 25000",

	// Commands
	"new_bill_usage":     "Harap berikan nama tagihan. Contoh: /new Sarapan atau kirim /new Sarapan dengan foto tagihan.",
	"add_contact_prompt": "Untuk menambahkan peserta, silakan kirim satu atau lebih kontak WhatsApp sekarang. Bot akan menambahkan kontak tersebut sebagai peserta pada tagihan saat ini.",
	"your_id":            "ID WhatsApp Anda: %s",

	// Private message
	"private_message":        "Hasil Perhitungan Tagihan: %s\n\n%s\n\nBayar ke: %s",
	"private_message_failed": "Gagal mengirim pesan pribadi ke %s (%s).",

	// Calculation
	"calculation_result": "*Hasil Perhitungan Tagihan: %s*\n\n%s\n\nTotal: %s\nJumlah peserta: %d\nBagian per orang: %s",
	"no_participants":    "Tidak ada peserta dalam tagihan ini. Silakan gunakan /join atau /participant untuk menambahkan peserta.",
	"no_items":           "Tidak ada item dalam tagihan ini. Silakan tambahkan item dengan /add atau kirim foto tagihan.",

	// Help text
	"help_text": `*Bantuan Bot Split Bill*

Cara menggunakan WhatsApp Split Bill Bot:

1. Buat tagihan baru:
   /new <nama_tagihan>
   atau
   /new <nama_tagihan> dengan foto tagihan 📷 (kirim foto untuk ekstraksi otomatis)
2. Setiap peserta ketik /join untuk berpartisipasi
3. Tambahkan item dan jumlah:
   /add <nama_item> <jumlah>
   (Bisa dilewati jika mengirim foto tagihan)
4. Hitung pembagian:
   /calculate
5. Tutup tagihan saat selesai:
   /close (setiap peserta akan menerima pesan pribadi dengan bagian tagihannya)

*Daftar Perintah:*
/new [nama] - Buat tagihan baru
/add [item] [jumlah] - Tambahkan item ke tagihan
/join [nama_tagihan] - Gabung sebagai peserta (bisa juga untuk mengubah nama tagihan)
/participant - Tambahkan peserta dengan mengirim kontak
/calculate - Hitung dan tampilkan pembagian tagihan
/close - Tutup tagihan dan kirim pesan pribadi ke peserta
/bill - Tampilkan detail tagihan dan daftar peserta
/help - Tampilkan pesan bantuan ini
/myid - Tampilkan ID WhatsApp Anda
/lang [indonesia|english] - Ubah bahasa bot untuk chat ini

Contoh penggunaan:
1. /new <nama_tagihan> dengan foto tagihan 📷 atau /new <nama_tagihan>
2. Semua orang ketik /join atau /participant dengan lampiran kontak
3. /add <nama_item> <jumlah> (bisa dilewati jika mengirim foto tagihan)
4. /calculate
5. /close saat selesai

Tentang:
Dibuat oleh Hendro Wibowo (https://github.com/w33ladalah) dan Affandy Fahrizain (https://github.com/fhrzn)
https://github.com/w33ladalah/split-billing-whatsapp
`,
}
