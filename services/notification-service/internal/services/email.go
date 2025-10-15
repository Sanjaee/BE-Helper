package services

import (
	"fmt"
	"log"

	"notification-service/internal/config"

	"gopkg.in/gomail.v2"
)

// EmailService handles email operations
type EmailService struct {
	config *config.Config
}

// EmailData represents email content
type EmailData struct {
	To      string
	Subject string
	Body    string
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.Config) (*EmailService, error) {
	if cfg.SMTPUsername == "" {
		return nil, fmt.Errorf("SMTP_USERNAME is required")
	}
	if cfg.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD is required")
	}

	return &EmailService{
		config: cfg,
	}, nil
}

// SendEmail sends a generic email
func (es *EmailService) SendEmail(emailData EmailData) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", es.config.FromName, es.config.FromEmail))
	m.SetHeader("To", emailData.To)
	m.SetHeader("Subject", emailData.Subject)
	m.SetBody("text/html", emailData.Body)

	d := gomail.NewDialer(es.config.SMTPHost, es.config.SMTPPort, es.config.SMTPUsername, es.config.SMTPPassword)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("✅ Email sent successfully to: %s", emailData.To)
	return nil
}

// SendOTPEmail sends OTP verification email
func (es *EmailService) SendOTPEmail(to, username, otp string) error {
	subject := "Verifikasi Email - ZACloth"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .otp-code { background: #667eea; color: white; font-size: 32px; font-weight: bold; padding: 20px; text-align: center; border-radius: 8px; margin: 20px 0; letter-spacing: 5px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .button { background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Selamat Datang di ZACloth!</h1>
        </div>
        <div class="content">
            <h2>Halo %s!</h2>
            <p>Terima kasih telah mendaftar di ZACloth. Untuk melengkapi proses pendaftaran, silakan verifikasi email Anda dengan kode OTP berikut:</p>
            
            <div class="otp-code">%s</div>
            
            <p><strong>Kode ini berlaku selama 10 menit.</strong></p>
            
            <p>Jika Anda tidak mendaftar di ZACloth, silakan abaikan email ini.</p>
            
            <p>Terima kasih,<br>Tim ZACloth</p>
        </div>
        <div class="footer">
            <p>Email ini dikirim secara otomatis, mohon tidak membalas email ini.</p>
        </div>
    </div>
</body>
</html>`, subject, username, otp)

	return es.SendEmail(EmailData{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

// SendPasswordResetEmail sends password reset OTP email
func (es *EmailService) SendPasswordResetEmail(to, username, otp string) error {
	subject := "Reset Password - ZACloth"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .otp-code { background: #e74c3c; color: white; font-size: 32px; font-weight: bold; padding: 20px; text-align: center; border-radius: 8px; margin: 20px 0; letter-spacing: 5px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .warning { background: #fff3cd; border: 1px solid #ffeaa7; color: #856404; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Reset Password - ZACloth</h1>
        </div>
        <div class="content">
            <h2>Halo %s!</h2>
            <p>Kami menerima permintaan untuk mereset password akun ZACloth Anda. Gunakan kode verifikasi berikut untuk melanjutkan:</p>
            
            <div class="otp-code">%s</div>
            
            <div class="warning">
                <strong>⚠️ Penting:</strong>
                <ul>
                    <li>Kode ini berlaku selama 10 menit</li>
                    <li>Jangan bagikan kode ini kepada siapa pun</li>
                    <li>Jika Anda tidak meminta reset password, abaikan email ini</li>
                </ul>
            </div>
            
            <p>Jika Anda tidak meminta reset password, silakan abaikan email ini dan password Anda akan tetap aman.</p>
            
            <p>Terima kasih,<br>Tim ZACloth</p>
        </div>
        <div class="footer">
            <p>Email ini dikirim secara otomatis, mohon tidak membalas email ini.</p>
        </div>
    </div>
</body>
</html>`, subject, username, otp)

	return es.SendEmail(EmailData{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

// SendWelcomeEmail sends welcome email to new users
func (es *EmailService) SendWelcomeEmail(to, username string) error {
	subject := "Selamat Datang di ZACloth!"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .button { background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Selamat Datang di ZACloth!</h1>
        </div>
        <div class="content">
            <h2>Halo %s!</h2>
            <p>Selamat! Akun ZACloth Anda telah berhasil dibuat dan diverifikasi. Sekarang Anda dapat menikmati berbagai produk fashion terbaik dari kami.</p>
            
            <p>Berikut adalah beberapa hal yang dapat Anda lakukan:</p>
            <ul>
                <li>Jelajahi koleksi produk terbaru kami</li>
                <li>Nikmati pengalaman berbelanja yang mudah dan aman</li>
                <li>Dapatkan notifikasi tentang penawaran khusus</li>
                <li>Kelola profil dan preferensi Anda</li>
            </ul>
            
            <p>Terima kasih telah bergabung dengan ZACloth. Kami berharap Anda menikmati pengalaman berbelanja bersama kami!</p>
            
            <p>Salam hangat,<br>Tim ZACloth</p>
        </div>
        <div class="footer">
            <p>Email ini dikirim secara otomatis, mohon tidak membalas email ini.</p>
        </div>
    </div>
</body>
</html>`, subject, username)

	return es.SendEmail(EmailData{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

// SendOrderConfirmationEmail sends order confirmation email
func (es *EmailService) SendOrderConfirmationEmail(to, username, orderID string, amount float64, currency string) error {
	subject := fmt.Sprintf("Konfirmasi Pesanan #%s - ZACloth", orderID)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .order-details { background: white; padding: 20px; border-radius: 8px; margin: 20px 0; border: 1px solid #ddd; }
        .amount { font-size: 24px; font-weight: bold; color: #667eea; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Pesanan Dikonfirmasi!</h1>
        </div>
        <div class="content">
            <h2>Halo %s!</h2>
            <p>Terima kasih! Pesanan Anda telah berhasil dibuat dan sedang diproses.</p>
            
            <div class="order-details">
                <h3>Detail Pesanan:</h3>
                <p><strong>Nomor Pesanan:</strong> #%s</p>
                <p><strong>Total Pembayaran:</strong> <span class="amount">%s %.2f</span></p>
            </div>
            
            <p>Kami akan segera memproses pesanan Anda dan mengirimkan detail pengiriman melalui email terpisah.</p>
            
            <p>Terima kasih telah berbelanja di ZACloth!</p>
            
            <p>Salam,<br>Tim ZACloth</p>
        </div>
        <div class="footer">
            <p>Email ini dikirim secara otomatis, mohon tidak membalas email ini.</p>
        </div>
    </div>
</body>
</html>`, subject, username, orderID, currency, amount)

	return es.SendEmail(EmailData{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

// SendPaymentSuccessEmail sends payment success notification
func (es *EmailService) SendPaymentSuccessEmail(to, username, orderID, paymentID string, amount float64, currency string) error {
	subject := "Pembayaran Berhasil - ZACloth"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #28a745 0%%, #20c997 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .payment-details { background: white; padding: 20px; border-radius: 8px; margin: 20px 0; border: 1px solid #ddd; }
        .success { color: #28a745; font-weight: bold; }
        .amount { font-size: 24px; font-weight: bold; color: #28a745; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Pembayaran Berhasil!</h1>
        </div>
        <div class="content">
            <h2>Halo %s!</h2>
            <p class="success">✅ Pembayaran Anda telah berhasil diproses!</p>
            
            <div class="payment-details">
                <h3>Detail Pembayaran:</h3>
                <p><strong>Nomor Pesanan:</strong> #%s</p>
                <p><strong>ID Pembayaran:</strong> %s</p>
                <p><strong>Jumlah:</strong> <span class="amount">%s %.2f</span></p>
            </div>
            
            <p>Pesanan Anda sekarang sedang diproses dan akan segera dikirim. Anda akan menerima notifikasi pengiriman melalui email.</p>
            
            <p>Terima kasih telah berbelanja di ZACloth!</p>
            
            <p>Salam,<br>Tim ZACloth</p>
        </div>
        <div class="footer">
            <p>Email ini dikirim secara otomatis, mohon tidak membalas email ini.</p>
        </div>
    </div>
</body>
</html>`, subject, username, orderID, paymentID, currency, amount)

	return es.SendEmail(EmailData{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

// SendPaymentFailedEmail sends payment failure notification
func (es *EmailService) SendPaymentFailedEmail(to, username, orderID string, reason string) error {
	subject := "Pembayaran Gagal - ZACloth"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #dc3545 0%%, #c82333 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .error { background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .button { background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Pembayaran Gagal</h1>
        </div>
        <div class="content">
            <h2>Halo %s!</h2>
            <p>Maaf, pembayaran untuk pesanan Anda tidak dapat diproses.</p>
            
            <div class="error">
                <strong>Detail Pesanan:</strong><br>
                Nomor Pesanan: #%s<br>
                Alasan: %s
            </div>
            
            <p>Anda dapat mencoba melakukan pembayaran kembali dengan:</p>
            <ul>
                <li>Memastikan informasi kartu kredit/debit Anda benar</li>
                <li>Memeriksa saldo rekening Anda</li>
                <li>Mencoba metode pembayaran lain</li>
            </ul>
            
            <p>Jika masalah berlanjut, silakan hubungi customer service kami.</p>
            
            <p>Salam,<br>Tim ZACloth</p>
        </div>
        <div class="footer">
            <p>Email ini dikirim secara otomatis, mohon tidak membalas email ini.</p>
        </div>
    </div>
</body>
</html>`, subject, username, orderID, reason)

	return es.SendEmail(EmailData{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

// HealthCheck checks if email service is properly configured
func (es *EmailService) HealthCheck() error {
	if es.config.SMTPHost == "" || es.config.SMTPUsername == "" || es.config.SMTPPassword == "" {
		return fmt.Errorf("email service not properly configured")
	}
	return nil
}
