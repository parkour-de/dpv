package email

import (
	"bytes"
	"crypto/tls"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/dpv"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net/smtp"
	"time"
)

type Service struct {
	Config *dpv.Config
}

func NewService(config *dpv.Config) *Service {
	return &Service{Config: config}
}

type ValidationData struct {
	User          *entities.User
	ValidationURL string
	ExpiryTime    time.Time
	NewEmail      string
	IsEmailChange bool
}

// Data for password reset email
type PasswordResetData struct {
	User       *entities.User
	ResetURL   string
	ExpiryTime time.Time
}

func (s *Service) SendEmailValidationEmail(data ValidationData) error {
	// Configure SMTP
	auth := smtp.PlainAuth("",
		s.Config.Email.SMTPUsername,
		s.Config.Email.SMTPPassword,
		s.Config.Email.SMTPHost)

	// Create TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.Config.Email.SMTPHost,
	}

	// Connect to server
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", s.Config.Email.SMTPHost, s.Config.Email.SMTPPort), tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.Config.Email.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender and recipient
	if err = client.Mail(s.Config.Email.FromAddress); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	targetEmail := data.NewEmail
	if targetEmail == "" {
		targetEmail = data.User.Email
	}

	if err = client.Rcpt(targetEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send email
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	message := s.generateValidationEmail(data)
	if _, err = writer.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write email data: %w", err)
	}

	return writer.Close()
}

// SendPasswordResetEmail sends a password reset email to the user
func (s *Service) SendPasswordResetEmail(data PasswordResetData) error {
	auth := smtp.PlainAuth("",
		s.Config.Email.SMTPUsername,
		s.Config.Email.SMTPPassword,
		s.Config.Email.SMTPHost)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.Config.Email.SMTPHost,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", s.Config.Email.SMTPHost, s.Config.Email.SMTPPort), tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.Config.Email.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	if err = client.Mail(s.Config.Email.FromAddress); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(data.User.Email); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	message := s.generatePasswordResetEmail(data)
	if _, err = writer.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write email data: %w", err)
	}

	return writer.Close()
}

func (s *Service) generateValidationEmail(data ValidationData) string {
	// Generate Message-ID
	messageID := fmt.Sprintf("<%d.%s@parkour-deutschland.de>",
		time.Now().Unix(), data.User.Key)

	// Format expiry time in Berlin timezone
	berlinLocation, _ := time.LoadLocation("Europe/Berlin")
	expiryBerlin := data.ExpiryTime.In(berlinLocation)

	targetEmail := data.NewEmail
	if targetEmail == "" {
		targetEmail = data.User.Email
	}

	subject := "E-Mail-Adresse bestätigen - Deutscher Parkour Verband"
	if data.IsEmailChange {
		subject = "Neue E-Mail-Adresse bestätigen - Deutscher Parkour Verband"
	}

	// Encode subject only if necessary
	encodedSubject := s.encodeSubjectIfNeeded(subject)

	actionText := "Ihre E-Mail-Adresse zu bestätigen"
	explanationText := fmt.Sprintf("Sie haben sich kürzlich bei der DPV-Mitgliederverwaltung mit der E-Mail-Adresse %s registriert.", data.User.Email)

	if data.IsEmailChange {
		actionText = fmt.Sprintf("Ihre neue E-Mail-Adresse (%s) zu bestätigen", data.NewEmail)
		explanationText = fmt.Sprintf("Sie haben eine Änderung Ihrer E-Mail-Adresse von %s zu %s beantragt.", data.User.Email, data.NewEmail)
	}

	// Generate boundary for multipart
	boundary := fmt.Sprintf("boundary_%d_%s", time.Now().Unix(), data.User.Key)

	// Generate text and HTML parts
	textBody := s.generateTextEmail(data, actionText, explanationText, expiryBerlin)
	htmlBody := s.generateHTMLEmail(data, actionText, explanationText, expiryBerlin)

	// Create multipart email with proper headers
	return fmt.Sprintf(`Message-ID: %s
Date: %s
MIME-Version: 1.0
From: %s <%s>
To: <%s>
Subject: %s
Content-Type: multipart/alternative; boundary="%s"

This is a multi-part message in MIME format.

--%s
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: 8bit

%s

--%s
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

%s

--%s--`,
		messageID,
		time.Now().Format(time.RFC1123Z),
		s.Config.Email.FromName,
		s.Config.Email.FromAddress,
		targetEmail,
		encodedSubject,
		boundary,
		boundary,
		textBody,
		boundary,
		s.quotedPrintableEncode(htmlBody),
		boundary)
}

// generatePasswordResetEmail creates the password reset email
func (s *Service) generatePasswordResetEmail(data PasswordResetData) string {
	messageID := fmt.Sprintf("<%d.%s@parkour-deutschland.de>",
		time.Now().Unix(), data.User.Key)
	berlinLocation, _ := time.LoadLocation("Europe/Berlin")
	expiryBerlin := data.ExpiryTime.In(berlinLocation)
	subject := "Passwort zurücksetzen - Deutscher Parkour Verband"
	encodedSubject := s.encodeSubjectIfNeeded(subject)
	boundary := fmt.Sprintf("boundary_%d_%s", time.Now().Unix(), data.User.Key)

	textBody := fmt.Sprintf(`DEUTSCHER PARKOUR VERBAND
Passwort zurücksetzen

Hallo %s %s,

Sie haben eine Anfrage zum Zurücksetzen Ihres Passworts gestellt.

Um Ihr Passwort zurückzusetzen, öffnen Sie bitte den folgenden Link in Ihrem Browser:

%s

Alternativ können Sie diesen Link kopieren und in Ihren Browser einfügen:

%s

WICHTIG: Dieser Link ist nur bis zum %s gültig.

Falls Sie diese Anfrage nicht gestellt haben, ignorieren Sie diese E-Mail einfach.

© %d Deutscher Parkour Verband`,
		data.User.Vorname, data.User.Name,
		data.ResetURL, data.ResetURL,
		expiryBerlin.Format("02.01.2006 um 15:04 Uhr"),
		time.Now().Year())

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="de">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Passwort zurücksetzen</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px">
    <div style="background-color: #2c5aa0; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0">
        <h1>Deutscher Parkour Verband</h1>
        <h2>Passwort zurücksetzen</h2>
    </div>
    <div style="background-color: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px">
        <p>Hallo %s %s,</p>
        <p>Sie haben eine Anfrage zum Zurücksetzen Ihres Passworts gestellt.</p>
        <p>Um Ihr Passwort zurückzusetzen, klicken Sie bitte auf den folgenden Button:</p>
        <p style="text-align: center;">
            <a href="%s" style="display: inline-block; background-color: #2c5aa0; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 20px 0"><span style="color: white">Passwort zurücksetzen</span></a>
        </p>
        <div style="margin-top: 20px; padding: 15px; background-color: #e3f2fd; border-radius: 5px; font-size: 14px; word-break: break-all">
            <strong>Alternativ können Sie diesen Link kopieren und in Ihren Browser einfügen:</strong><br>
            <a href="%s">%s</a>
        </div>
        <p><strong>Wichtig:</strong> Dieser Link ist nur bis zum <strong>%s</strong> gültig.</p>
        <p>Falls Sie diese Anfrage nicht gestellt haben, ignorieren Sie diese E-Mail einfach.</p>
    </div>
    <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666">
        <p><strong>Über die DPV-Mitgliederverwaltung:</strong><br>
        Die DPV-Mitgliederverwaltung ist das offizielle System des Deutschen Parkour Verbandes zur Verwaltung von Mitgliedschaften, Vereinen und Organisationen. Mit diesem System können Sie Ihre Mitgliedschaft beantragen, Vereinsdaten verwalten und an der Parkour-Community in Deutschland teilnehmen.</p>
        
        <p>Bei Fragen wenden Sie sich an: <a href="mailto:info@parkour-deutschland.de">info@parkour-deutschland.de</a></p>
        
        <p>© %d Deutscher Parkour Verband</p>
    </div>
</body>
</html>`,
		data.User.Vorname, data.User.Name,
		data.ResetURL, data.ResetURL, data.ResetURL,
		expiryBerlin.Format("02.01.2006 um 15:04 Uhr"),
		time.Now().Year())

	return fmt.Sprintf(`Message-ID: %s
Date: %s
MIME-Version: 1.0
From: %s <%s>
To: <%s>
Subject: %s
Content-Type: multipart/alternative; boundary="%s"

This is a multi-part message in MIME format.

--%s
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: 8bit

%s

--%s
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

%s

--%s--`,
		messageID,
		time.Now().Format(time.RFC1123Z),
		s.Config.Email.FromName,
		s.Config.Email.FromAddress,
		data.User.Email,
		encodedSubject,
		boundary,
		boundary,
		textBody,
		boundary,
		s.quotedPrintableEncode(htmlBody),
		boundary)
}

// encodeSubjectIfNeeded encodes subject only if it contains non-ASCII characters
func (s *Service) encodeSubjectIfNeeded(subject string) string {
	needsEncoding := false
	for _, r := range subject {
		if r > 127 {
			needsEncoding = true
			break
		}
	}

	if needsEncoding {
		return mime.QEncoding.Encode("UTF-8", subject)
	}

	return subject
}

// generateTextEmail creates the plain text version
func (s *Service) generateTextEmail(data ValidationData, actionText, explanationText string, expiryBerlin time.Time) string {
	targetEmail := data.NewEmail
	if targetEmail == "" {
		targetEmail = data.User.Email
	}

	return fmt.Sprintf(`DEUTSCHER PARKOUR VERBAND
E-Mail-Bestätigung

Hallo %s %s,

%s

Um %s, öffnen Sie bitte den folgenden Link in Ihrem Browser:

%s

Alternativ können Sie diesen Link kopieren und in Ihren Browser einfügen:

%s

WICHTIG: Dieser Link ist nur bis zum %s gültig.

Falls Sie diese Anfrage nicht gestellt haben, ignorieren Sie diese E-Mail einfach.

---

Über die DPV-Mitgliederverwaltung:
Die DPV-Mitgliederverwaltung ist das offizielle System des Deutschen Parkour Verbandes zur Verwaltung von Mitgliedschaften, Vereinen und Organisationen. Mit diesem System können Sie Ihre Mitgliedschaft beantragen, Vereinsdaten verwalten und an der Parkour-Community in Deutschland teilnehmen.

Bei Fragen wenden Sie sich an: info@parkour-deutschland.de

© %d Deutscher Parkour Verband`,
		data.User.Vorname, data.User.Name,
		explanationText,
		actionText,
		data.ValidationURL, data.ValidationURL,
		expiryBerlin.Format("02.01.2006 um 15:04 Uhr"),
		time.Now().Year())
}

// generateHTMLEmail creates the HTML version (existing function, cleaned up)
func (s *Service) generateHTMLEmail(data ValidationData, actionText, explanationText string, expiryBerlin time.Time) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="de">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>E-Mail bestätigen</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px">
    <div style="background-color: #2c5aa0; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0">
        <h1>Deutscher Parkour Verband</h1>
        <h2>E-Mail-Bestätigung</h2>
    </div>
    <div style="background-color: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px">
        <p>Hallo %s %s,</p>
        
        <p>%s</p>
        
        <p>Um %s, klicken Sie bitte auf den folgenden Button:</p>
        
        <p style="text-align: center;">
            <a href="%s" style="display: inline-block; background-color: #2c5aa0; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 20px 0"><span style="color: white">E-Mail-Adresse bestätigen</span></a>
        </p>
        
        <div style="margin-top: 20px; padding: 15px; background-color: #e3f2fd; border-radius: 5px; font-size: 14px; word-break: break-all">
            <strong>Alternativ können Sie diesen Link kopieren und in Ihren Browser einfügen:</strong><br>
            <a href="%s">%s</a>
        </div>
        
        <p><strong>Wichtig:</strong> Dieser Link ist nur bis zum <strong>%s</strong> gültig.</p>
        
        <p>Falls Sie diese Anfrage nicht gestellt haben, ignorieren Sie diese E-Mail einfach.</p>
    </div>
    <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666">
        <p><strong>Über die DPV-Mitgliederverwaltung:</strong><br>
        Die DPV-Mitgliederverwaltung ist das offizielle System des Deutschen Parkour Verbandes zur Verwaltung von Mitgliedschaften, Vereinen und Organisationen. Mit diesem System können Sie Ihre Mitgliedschaft beantragen, Vereinsdaten verwalten und an der Parkour-Community in Deutschland teilnehmen.</p>
        
        <p>Bei Fragen wenden Sie sich an: <a href="mailto:info@parkour-deutschland.de">info@parkour-deutschland.de</a></p>
        
        <p>© %d Deutscher Parkour Verband</p>
    </div>
</body>
</html>`,
		data.User.Vorname, data.User.Name,
		explanationText,
		actionText,
		data.ValidationURL, data.ValidationURL, data.ValidationURL,
		expiryBerlin.Format("02.01.2006 um 15:04 Uhr"),
		time.Now().Year())
}

func (s *Service) quotedPrintableEncode(input string) string {
	var buf bytes.Buffer
	writer := quotedprintable.NewWriter(&buf)

	// Write the input string
	_, err := writer.Write([]byte(input))
	if err != nil {
		// Fallback to simple encoding if something goes wrong
		return input
	}

	// Close the writer to flush any remaining data
	err = writer.Close()
	if err != nil {
		return input
	}

	return buf.String()
}
