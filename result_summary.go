package main

// These summaries interpret the reported counts only. They never invent a
// finding, infer coverage, or translate technical evidence that was not supplied.
func resultNextStep(reported, critical, high, medium, low int) (en, id string) {
	if reported != 1 {
		return "Finding counts are unavailable. Read the PDF with your developer before deciding what needs attention.", "Jumlah temuan belum tersedia. Baca PDF bersama developer sebelum menentukan masalah yang perlu ditangani."
	}
	if critical+high > 0 {
		return "Ask your developer to review the Critical and High findings promptly. Check the evidence and plan repairs before lower-priority work.", "Minta developer segera meninjau temuan Critical dan High. Periksa buktinya dan rencanakan perbaikan sebelum pekerjaan dengan prioritas lebih rendah."
	}
	if medium+low > 0 {
		return "Review the findings with your developer and schedule repairs according to their impact. Read the evidence before making changes.", "Tinjau temuan bersama developer dan jadwalkan perbaikan sesuai dampaknya. Baca bukti sebelum melakukan perubahan."
	}
	return "No Critical, High, Medium or Low findings were reported. Review any informational notes and the scope; this does not establish that the whole application is safe.", "Tidak ada temuan Critical, High, Medium, atau Low yang dilaporkan. Tinjau catatan informasi dan cakupannya; hasil ini tidak membuktikan seluruh aplikasi aman."
}
