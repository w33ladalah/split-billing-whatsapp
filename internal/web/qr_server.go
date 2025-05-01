package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sync"
)

var (
	qrCode     string
	qrCodeLock sync.RWMutex
	paired     bool
	pairedLock sync.RWMutex
)

// SetPaired updates the paired status
func SetPaired(val bool) {
	pairedLock.Lock()
	defer pairedLock.Unlock()
	paired = val
}

// GetPaired returns the paired status
func GetPaired() bool {
	pairedLock.RLock()
	defer pairedLock.RUnlock()
	return paired
}

// SetQRCode updates the current QR code string
func SetQRCode(code string) {
	qrCodeLock.Lock()
	defer qrCodeLock.Unlock()
	qrCode = code
}

// GetQRCode returns the current QR code string
func GetQRCode() string {
	qrCodeLock.RLock()
	defer qrCodeLock.RUnlock()
	return qrCode
}

// ServeQR starts an HTTP server for QR code display
func ServeQR(addr string) {
	http.HandleFunc("/qr", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"qr": GetQRCode(),
			"paired": GetPaired(),
		})
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.New("qr").Parse(htmlPage))
		_ = tmpl.Execute(w, nil)
	})

	go http.ListenAndServe(addr, nil)
}

const htmlPage = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>WhatsApp QR Code</title>
	<script src="https://cdn.jsdelivr.net/npm/qrcode@1.5.1/build/qrcode.min.js"></script>
	<style>
		body { font-family: sans-serif; text-align: center; margin-top: 40px; }
		#qr { margin: 20px auto; }
	</style>
</head>
<body>
	<h2>Scan this QR code with WhatsApp</h2>
	<div id="qr"></div>
	<div id="text" style="margin-top: 20px; color: #888;"></div>
	<script>
	async function fetchQR() {
		const res = await fetch('/qr');
		const data = await res.json();
		const qr = data.qr;
		const paired = data.paired;
		const qrDiv = document.getElementById('qr');
		const textDiv = document.getElementById('text');
		qrDiv.innerHTML = '';
		if (paired) {
			textDiv.innerText = '✅ Successfully paired! You can close this window.';
			return;
		}
		if (qr) {
			const canvas = document.createElement('canvas');
			QRCode.toCanvas(canvas, qr, { width: 256 }, function (error) {
				if (error) qrDiv.innerText = 'QR error: ' + error;
			});
			qrDiv.appendChild(canvas);
			textDiv.innerText = '';
		} else {
			textDiv.innerText = 'Waiting for QR code...';
		}
	}
	setInterval(fetchQR, 2000);
	fetchQR();
	</script>
</body>
</html>`
