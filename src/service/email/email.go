package email

import (
	"bytes"
	"crypto/tls"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/dpv"
	"fmt"
	"html/template"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	text_template "text/template"
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

func (s *Service) createSMTPClient() (*smtp.Client, error) {
	addr := net.JoinHostPort(s.Config.Email.SMTPHost, fmt.Sprint(s.Config.Email.SMTPPort))
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.Config.Email.SMTPHost,
	}

	var client *smtp.Client

	if s.Config.Email.SMTPPort == 465 {
		// Implicit TLS
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		client, err = smtp.NewClient(conn, s.Config.Email.SMTPHost)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		// Explicit TLS (STARTTLS), usually port 587 or 25
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		client, err = smtp.NewClient(conn, s.Config.Email.SMTPHost)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create SMTP client: %w", err)
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate
	auth := smtp.PlainAuth("",
		s.Config.Email.SMTPUsername,
		s.Config.Email.SMTPPassword,
		s.Config.Email.SMTPHost)

	if err := client.Auth(auth); err != nil {
		client.Close()
		return nil, fmt.Errorf("SMTP authentication failed: %w", err)
	}

	return client, nil
}

func (s *Service) SendEmailValidationEmail(data ValidationData) error {
	if s.Config.Email.SMTPHost == "" {
		return nil
	}

	client, err := s.createSMTPClient()
	if err != nil {
		return err
	}
	defer client.Quit()

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
	if s.Config.Email.SMTPHost == "" {
		return nil
	}

	client, err := s.createSMTPClient()
	if err != nil {
		return err
	}
	defer client.Quit()

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



var (
	textBaseTemplate *text_template.Template
	htmlBaseTemplate *template.Template
)

func init() {
	textBaseTemplate = text_template.Must(text_template.New("text").Parse(`DEUTSCHER PARKOUR VERBAND
{{.Title}}

{{.Content}}

---

Über die DPV-Mitgliederverwaltung:
Die DPV-Mitgliederverwaltung ist das offizielle System des Deutschen Parkour Verbandes zur Verwaltung von Mitgliedschaften, Vereinen und Organisationen. Mit diesem System kannst Du Deine Mitgliedschaft beantragen, Vereinsdaten verwalten und an der Parkour-Community in Deutschland teilnehmen.

Bei Fragen wende Dich an: info@parkour-deutschland.de

© {{.Year}} Deutscher Parkour Verband`))

	htmlBaseTemplate = template.Must(template.New("html").Parse(`<!DOCTYPE html>
<html lang="de">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px">
    <div style="background-color: #2c5aa0; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0">
        <h1 style="margin: 0; font-size: 24px;">Deutscher Parkour Verband</h1>
        <h2 style="margin: 10px 0 0; font-size: 18px; font-weight: normal;">{{.Title}}</h2>
    </div>
    <div style="background-color: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px">
{{.Content}}
    </div>
    <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666">
        <p><strong>Über die DPV-Mitgliederverwaltung:</strong><br>
        Die DPV-Mitgliederverwaltung ist das offizielle System des Deutschen Parkour Verbandes zur Verwaltung von Mitgliedschaften, Vereinen und Organisationen. Mit diesem System kannst Du Deine Mitgliedschaft beantragen, Vereinsdaten verwalten und an der Parkour-Community in Deutschland teilnehmen.</p>
        
        <p>Bei Fragen wende Dich an: <a href="mailto:info@parkour-deutschland.de">info@parkour-deutschland.de</a></p>
        
        <p>© {{.Year}} Deutscher Parkour Verband</p>
    </div>
</body>
</html>`))
}

type templateData struct {
	Title   string
	Content interface{}
	Year    int
}

func (s *Service) wrapText(title, content string) string {
	var buf bytes.Buffer
	textBaseTemplate.Execute(&buf, templateData{
		Title:   title,
		Content: content,
		Year:    time.Now().Year(),
	})
	return buf.String()
}

func (s *Service) wrapHTML(title, content string) string {
	var buf bytes.Buffer
	htmlBaseTemplate.Execute(&buf, templateData{
		Title:   title,
		Content: template.HTML(content),
		Year:    time.Now().Year(),
	})
	return buf.String()
}

func (s *Service) sendGenericEmail(to, subject, textBody, htmlBody string) error {
	if s.Config == nil || s.Config.Email.SMTPHost == "" {
		return nil
	}

	client, err := s.createSMTPClient()
	if err != nil {
		return err
	}
	defer client.Quit()

	if err = client.Mail(s.Config.Email.FromAddress); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	messageID := fmt.Sprintf("<%d.%s@parkour-deutschland.de>", time.Now().UnixNano(), to)
	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
	message := fmt.Sprintf("Message-ID: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\nFrom: %s <%s>\r\nTo: <%s>\r\nSubject: %s\r\nContent-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n\r\n--%s\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n%s\r\n\r\n--%s--",
		messageID, time.Now().Format(time.RFC1123Z), s.Config.Email.FromName, s.Config.Email.FromAddress, to, s.encodeSubjectIfNeeded(subject), boundary, boundary, textBody, boundary, s.quotedPrintableEncode(htmlBody), boundary)

	_, err = writer.Write([]byte(message))
	if err != nil {
		return err
	}
	return writer.Close()
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *Service) SendWelcomeEmail(user *entities.User) error {
	subject := "Willkommen beim Deutschen Parkour Verband"
	textBody := fmt.Sprintf(`Hallo %s %s,

vielen Dank für Deine Registrierung in der DPV-Mitgliederverwaltung!
Du kannst Dich ab sofort mit Deiner E-Mail-Adresse %s anmelden.

In der Mitgliederverwaltung kannst Du Deine Vereine anlegen, Mitgliedschaften beantragen und Bestandsmeldungen abgeben.

Hier geht es zum Login: %s

Viel Erfolg bei Deiner Vereinsarbeit!`,
		user.FirstName, user.LastName, user.Email, s.Config.Settings.BaseURL)

	htmlBody := fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>vielen Dank für Deine Registrierung in der DPV-Mitgliederverwaltung!</p>
        <p>Du kannst Dich ab sofort mit Deiner E-Mail-Adresse <strong>%s</strong> anmelden.</p>
        <p style="text-align: center;">
            <a href="%s" style="display: inline-block; background-color: #2c5aa0; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 20px 0"><span style="color: white">Zum Login</span></a>
        </p>
        <p>In der Mitgliederverwaltung kannst Du Deine Vereine anlegen, Mitgliedschaften beantragen und Bestandsmeldungen abgeben.</p>
        <p>Viel Erfolg bei Deiner Vereinsarbeit!</p>`,
		user.FirstName, user.LastName, user.Email, s.Config.Settings.BaseURL)

	return s.sendGenericEmail(user.Email, subject, s.wrapText(subject, textBody), s.wrapHTML(subject, htmlBody))
}

// SendApplicationReceiptEmail sends an application receipt confirmation to the applicant
func (s *Service) SendApplicationReceiptEmail(user *entities.User, club *entities.Club) error {
	var subject, textBody, htmlBody string

	if club != nil {
		subject = "Eingang Deines Mitgliedsantrags für " + club.Name + " " + club.LegalForm
		textBody = fmt.Sprintf(`Hallo %s %s,

vielen Dank für den Mitgliedsantrag für Deinen Verein %s %s.
Wir werden diesen in Kürze prüfen und uns bei Dir melden.`,
			user.FirstName, user.LastName, club.Name, club.LegalForm)
		htmlBody = fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>vielen Dank für den Mitgliedsantrag für Deinen Verein <strong>%s %s</strong>.</p>
        <p>Wir werden diesen in Kürze prüfen und uns bei Dir melden.</p>`,
			user.FirstName, user.LastName, club.Name, club.LegalForm)
	} else {
		subject = "Eingang Deines Mitgliedsantrags"
		textBody = fmt.Sprintf(`Hallo %s %s,

vielen Dank für Deinen Mitgliedsantrag als Aktivmitglied.
Wir werden diesen in Kürze prüfen und uns bei Dir melden.`,
			user.FirstName, user.LastName)
		htmlBody = fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>vielen Dank für Deinen Mitgliedsantrag als Aktivmitglied.</p>
        <p>Wir werden diesen in Kürze prüfen und uns bei Dir melden.</p>`,
			user.FirstName, user.LastName)
	}

	return s.sendGenericEmail(user.Email, subject, s.wrapText(subject, textBody), s.wrapHTML(subject, htmlBody))
}

// SendApplicationNoticeEmail sends a notice email to the DPV board
func (s *Service) SendApplicationNoticeEmail(user *entities.User, club *entities.Club) error {
	var subject, textBody, htmlBody string
	to := "info@parkour-deutschland.de"

	if club != nil {
		subject = "Neuer Mitgliedsantrag: " + club.Name + " " + club.LegalForm
		textBody = fmt.Sprintf(`Hallo,

es liegt ein neuer Mitgliedsantrag für den Verein %s %s von %s %s vor.
Bitte in der Verwaltungsoberfläche prüfen.`,
			club.Name, club.LegalForm, user.FirstName, user.LastName)
		htmlBody = fmt.Sprintf(`        <p>Hallo,</p>
        <p>es liegt ein neuer Mitgliedsantrag für den Verein <strong>%s %s</strong> von <strong>%s %s</strong> vor.</p>
        <p>Bitte in der Verwaltungsoberfläche prüfen.</p>`,
			club.Name, club.LegalForm, user.FirstName, user.LastName)
	} else {
		subject = "Neuer Mitgliedsantrag: Aktivmitglied"
		textBody = fmt.Sprintf(`Hallo,

es liegt ein neuer Mitgliedsantrag als Aktivmitglied von %s %s vor.
Bitte in der Verwaltungsoberfläche prüfen.`,
			user.FirstName, user.LastName)
		htmlBody = fmt.Sprintf(`        <p>Hallo,</p>
        <p>es liegt ein neuer Mitgliedsantrag als Aktivmitglied von <strong>%s %s</strong> vor.</p>
        <p>Bitte in der Verwaltungsoberfläche prüfen.</p>`,
			user.FirstName, user.LastName)
	}

	return s.sendGenericEmail(to, subject, s.wrapText(subject, textBody), s.wrapHTML(subject, htmlBody))
}

// SendApplicationAcceptedEmail sends general acceptance emails to those applying
func (s *Service) SendApplicationAcceptedEmail(user *entities.User, club *entities.Club) error {
	if club != nil {
		for _, v := range club.Vorstand {
			if v.AuthorizedRepresentative && v.Email != "" {
				subject := fmt.Sprintf("Mitgliedsantrag für %s %s angenommen", club.Name, club.LegalForm)
				textBody := fmt.Sprintf(`Hallo %s %s,

der Mitgliedsantrag für Deinen Verein %s %s wurde soeben vom Deutschen Parkour Verband angenommen.
Wir freuen uns, Euch als Mitglied begrüßen zu dürfen! 

Die Mitgliedschaft ist nun aktiv. Gemeinsam können wir die Entwicklung des Parkoursports in Deutschland vorantreiben.`, v.Firstname, v.Lastname, club.Name, club.LegalForm)
				htmlBody := fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>der Mitgliedsantrag für Deinen Verein <strong>%s %s</strong> wurde soeben vom Deutschen Parkour Verband angenommen.</p>
        <p>Wir freuen uns, Euch als Mitglied begrüßen zu dürfen!</p>
        <p>Die Mitgliedschaft ist nun aktiv. Gemeinsam können wir die Entwicklung des Parkoursports in Deutschland vorantreiben.</p>`, v.Firstname, v.Lastname, club.Name, club.LegalForm)

				wText := s.wrapText(subject, textBody)
				wHTML := s.wrapHTML(subject, htmlBody)
				_ = s.sendGenericEmail(v.Email, subject, wText, wHTML)
			}
		}
		return nil
	} else {
		subject := "Mitgliedsantrag angenommen"
		textBody := fmt.Sprintf(`Hallo %s %s,

Dein Mitgliedsantrag als Aktivmitglied wurde soeben vom Deutschen Parkour Verband angenommen.
Wir freuen uns sehr, Dich als neues Mitglied in unserer Community begrüßen zu dürfen!

Deine Mitgliedschaft ist ab sofort aktiv. Gemeinsam können wir den Parkoursport stärken und unsere Ziele verwirklichen.`, user.FirstName, user.LastName)
		htmlBody := fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>Dein Mitgliedsantrag als Aktivmitglied wurde soeben vom Deutschen Parkour Verband angenommen.</p>
        <p>Wir freuen uns sehr, Dich als neues Mitglied in unserer Community begrüßen zu dürfen!</p>
        <p>Deine Mitgliedschaft ist ab sofort aktiv. Gemeinsam können wir den Parkoursport stärken und unsere Ziele verwirklichen.</p>`, user.FirstName, user.LastName)

		return s.sendGenericEmail(user.Email, subject, s.wrapText(subject, textBody), s.wrapHTML(subject, htmlBody))
	}
}

// SendMembershipBeganEmail sends an onboarding email with the vision
func (s *Service) SendMembershipBeganEmail(user *entities.User, club *entities.Club) error {
	if club != nil {
		for _, v := range club.Vorstand {
			if v.AuthorizedRepresentative && v.Email != "" {
				subject := fmt.Sprintf("Willkommen im DPV! Die Mitgliedschaft für %s %s hat begonnen", club.Name, club.LegalForm)
				textBody := fmt.Sprintf(`Hallo %s %s,

die Mitgliedschaft für Deinen Verein %s %s im Deutschen Parkour Verband hat nun offiziell begonnen! 
Wir freuen uns unglaublich über diesen gemeinsamen Meilenstein.

Durch den Anschluss an unseren Verband stärken wir die Selbstvertretung der Parkourszene in Deutschland weiter. Unsere Mission ist es, die Interessen frei zu bündeln und die Entwicklung des Sports gemeinsam und demokratisch voranzutreiben.
Wir sind froh, Euch auf diesem Weg an unserer Seite zu haben und sind fest entschlossen, die Förderung unseres Sports – sei es durch Zugang zu Fördermitteln, Anerkennungen oder den Bau neuer Anlagen – weiter auszubauen.

Herzlich willkommen im Team!`, v.Firstname, v.Lastname, club.Name, club.LegalForm)
				htmlBody := fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>die Mitgliedschaft für Deinen Verein <strong>%s %s</strong> im Deutschen Parkour Verband hat nun offiziell begonnen!</p>
        <p>Wir freuen uns unglaublich über diesen gemeinsamen Meilenstein.</p>
        <p>Durch den Anschluss an unseren Verband stärken wir die Selbstvertretung der Parkourszene in Deutschland weiter. Unsere Mission ist es, die Interessen frei zu bündeln und die Entwicklung des Sports gemeinsam und demokratisch voranzutreiben.</p>
        <p>Wir sind froh, Euch auf diesem Weg an unserer Seite zu haben und sind fest entschlossen, die Förderung unseres Sports – sei es durch Zugang zu Fördermitteln, Anerkennungen oder den Bau neuer Anlagen – weiter auszubauen.</p>
        <p><strong>Herzlich willkommen im Team!</strong></p>`, v.Firstname, v.Lastname, club.Name, club.LegalForm)

				wText := s.wrapText(subject, textBody)
				wHTML := s.wrapHTML(subject, htmlBody)
				_ = s.sendGenericEmail(v.Email, subject, wText, wHTML)
			}
		}
		return nil
	} else {
		subject := "Willkommen im DPV! Deine Mitgliedschaft hat begonnen"
		textBody := fmt.Sprintf(`Hallo %s %s,

Deine Mitgliedschaft im Deutschen Parkour Verband hat nun offiziell begonnen!
Wir sind der demokratische Dachverband und stehen für die freie und kreative Entfaltung des Parkoursports.

Ein großer Teil unserer Arbeit besteht darin, die Szene bundesweit zu vernetzen, das vielfältige Wissen der Community zu bündeln und Anlaufstelle für Politik, Sportbünde und für alle zu sein, die Parkour selbst entdecken wollen. 
Mit Deinem Beitrag stärkst Du unsere Stimme für Parkour in Deutschland entscheidend, um den Sport weiterzuentwickeln und eine offizielle Anerkennung als eigenständige Sportart im DOSB zu erzielen.

Schön, dass Du Teil der Mission bist!`, user.FirstName, user.LastName)
		htmlBody := fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>Deine Mitgliedschaft im Deutschen Parkour Verband hat nun offiziell <strong>begonnen!</strong></p>
        <p>Wir sind der demokratische Dachverband und stehen für die freie und kreative Entfaltung des Parkoursports.</p>
        <p>Ein großer Teil unserer Arbeit besteht darin, die Szene bundesweit zu vernetzen, das vielfältige Wissen der Community zu bündeln und Anlaufstelle für Politik, Sportbünde und für alle zu sein, die Parkour selbst entdecken wollen.</p>
        <p>Mit Deinem Beitrag stärkst Du unsere Stimme für Parkour in Deutschland entscheidend, um den Sport weiterzuentwickeln und eine offizielle Anerkennung als eigenständige Sportart im DOSB zu erzielen.</p>
        <p><strong>Schön, dass Du Teil der Mission bist!</strong></p>`, user.FirstName, user.LastName)
		return s.sendGenericEmail(user.Email, subject, s.wrapText(subject, textBody), s.wrapHTML(subject, htmlBody))
	}
}

// SendMembershipEndedEmail sends a goodbye email
func (s *Service) SendMembershipEndedEmail(user *entities.User, club *entities.Club) error {
	if club != nil {
		for _, v := range club.Vorstand {
			if v.AuthorizedRepresentative && v.Email != "" {
				subject := fmt.Sprintf("Ende der Mitgliedschaft für %s %s", club.Name, club.LegalForm)
				textBody := fmt.Sprintf(`Hallo %s %s,

die Mitgliedschaft für %s %s im Deutschen Parkour Verband ist beendet.
Wir möchten uns herzlich für die gemeinsame Zeit bedanken!

Es ist natürlich immer schade, wenn sich unsere Wege trennen, da jede Stimme den Parkoursport ein Stück präsenter und stärker macht. Wir hoffen, wir konnten Euch in der gemeinsamen Zeit eine gute Plattform bieten.

Solltet Ihr in Zukunft wieder aktiv dabei sein wollen, stehen unsere Türen natürlich immer offen. Wir wünschen Euch alles Gute!`, v.Firstname, v.Lastname, club.Name, club.LegalForm)
				htmlBody := fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>die Mitgliedschaft für <strong>%s %s</strong> im Deutschen Parkour Verband ist beendet.</p>
        <p>Wir möchten uns herzlich für die gemeinsame Zeit bedanken!</p>
        <p>Es ist natürlich immer schade, wenn sich unsere Wege trennen, da jede Stimme den Parkoursport ein Stück präsenter und stärker macht. Wir hoffen, wir konnten Euch in der gemeinsamen Zeit eine gute Plattform bieten.</p>
        <p>Solltet Ihr in Zukunft wieder aktiv dabei sein wollen, stehen unsere Türen natürlich immer offen. Wir wünschen Euch alles Gute!</p>`, v.Firstname, v.Lastname, club.Name, club.LegalForm)

				wText := s.wrapText(subject, textBody)
				wHTML := s.wrapHTML(subject, htmlBody)
				_ = s.sendGenericEmail(v.Email, subject, wText, wHTML)
			}
		}
		return nil
	} else {
		subject := "Ende Deiner Mitgliedschaft im DPV"
		textBody := fmt.Sprintf(`Hallo %s %s,

Deine Aktivmitgliedschaft im Deutschen Parkour Verband ist beendet.
Wir danken Dir für die Zeit, in der Du unsere Ziele unterstützt hast!

Jeder Unterstützer hat uns einen bedeutenden Schritt weiter gebracht, Parkour als eigenständige und normfreie Sportart zu vertreten und zu fördern. Wir bedauern Deinen Weggang sehr.

Vielleicht findest Du ja in Zukunft wieder den Weg in unseren Verband – Du bist jederzeit herzlich willkommen!
Bis dahin wünschen wir Dir viel Erfolg und weiche Landungen.`, user.FirstName, user.LastName)
		htmlBody := fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>Deine Aktivmitgliedschaft im Deutschen Parkour Verband ist beendet.</p>
        <p>Wir danken Dir für die Zeit, in der Du unsere Ziele unterstützt hast!</p>
        <p>Jeder Unterstützer hat uns einen bedeutenden Schritt weiter gebracht, Parkour als eigenständige und normfreie Sportart zu vertreten und zu fördern. Wir bedauern Deinen Weggang sehr.</p>
        <p>Vielleicht findest Du ja in Zukunft wieder den Weg in unseren Verband – Du bist jederzeit herzlich willkommen!</p>
        <p>Bis dahin wünschen wir Dir viel Erfolg und weiche Landungen.</p>`, user.FirstName, user.LastName)
		return s.sendGenericEmail(user.Email, subject, s.wrapText(subject, textBody), s.wrapHTML(subject, htmlBody))
	}
}

func (s *Service) generateValidationEmail(data ValidationData) string {
	messageID := fmt.Sprintf("<%d.%s@parkour-deutschland.de>",
		time.Now().Unix(), data.User.Key)
	berlinLocation, _ := time.LoadLocation("Europe/Berlin")
	expiryBerlin := data.ExpiryTime.In(berlinLocation)

	targetEmail := data.NewEmail
	if targetEmail == "" {
		targetEmail = data.User.Email
	}

	subject := "E-Mail-Adresse bestätigen"
	explanationText := fmt.Sprintf("Du hast Dich kürzlich bei der DPV-Mitgliederverwaltung mit der E-Mail-Adresse %s registriert.", data.User.Email)
	actionText := "Deine E-Mail-Adresse zu bestätigen"

	if data.IsEmailChange {
		subject = "Neue E-Mail-Adresse bestätigen"
		actionText = fmt.Sprintf("Deine neue E-Mail-Adresse (%s) zu bestätigen", data.NewEmail)
		explanationText = fmt.Sprintf("Du hast eine Änderung Deiner E-Mail-Adresse von %s zu %s beantragt.", data.User.Email, data.NewEmail)
	}

	encodedSubject := s.encodeSubjectIfNeeded(subject + " - Deutscher Parkour Verband")
	boundary := fmt.Sprintf("boundary_%d_%s", time.Now().Unix(), data.User.Key)

	textBody := fmt.Sprintf(`Hallo %s %s,

%s

Um %s, öffne bitte den folgenden Link in Deinem Browser:

%s

Alternativ kannst Du diesen Link kopieren und in Deinen Browser einfügen:

%s

WICHTIG: Dieser Link ist nur bis zum %s gültig.

Falls Du diese Anfrage nicht gestellt hast, ignoriere diese E-Mail einfach.`,
		data.User.FirstName, data.User.LastName,
		explanationText,
		actionText,
		data.ValidationURL, data.ValidationURL,
		expiryBerlin.Format("02.01.2006 um 15:04 Uhr"))

	htmlBody := fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>%s</p>
        <p>Um %s, klicke bitte auf den folgenden Button:</p>
        <p style="text-align: center;">
            <a href="%s" style="display: inline-block; background-color: #2c5aa0; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 20px 0"><span style="color: white">%s</span></a>
        </p>
        <div style="margin-top: 20px; padding: 15px; background-color: #e3f2fd; border-radius: 5px; font-size: 14px; word-break: break-all">
            <strong>Alternativ kannst Du diesen Link kopieren und in Deinen Browser einfügen:</strong><br>
            <a href="%s">%s</a>
        </div>
        <p><strong>Wichtig:</strong> Dieser Link ist nur bis zum <strong>%s</strong> gültig.</p>
        <p>Falls Du diese Anfrage nicht gestellt hast, ignoriere diese E-Mail einfach.</p>`,
		data.User.FirstName, data.User.LastName,
		explanationText,
		actionText,
		data.ValidationURL, subject,
		data.ValidationURL, data.ValidationURL,
		expiryBerlin.Format("02.01.2006 um 15:04 Uhr"))

	fullText := s.wrapText(subject, textBody)
	fullHTML := s.wrapHTML(subject, htmlBody)

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
		fullText,
		boundary,
		s.quotedPrintableEncode(fullHTML),
		boundary)
}

func (s *Service) generatePasswordResetEmail(data PasswordResetData) string {
	messageID := fmt.Sprintf("<%d.%s@parkour-deutschland.de>",
		time.Now().Unix(), data.User.Key)
	berlinLocation, _ := time.LoadLocation("Europe/Berlin")
	expiryBerlin := data.ExpiryTime.In(berlinLocation)
	subject := "Passwort zurücksetzen"
	encodedSubject := s.encodeSubjectIfNeeded(subject + " - Deutscher Parkour Verband")
	boundary := fmt.Sprintf("boundary_%d_%s", time.Now().Unix(), data.User.Key)

	textBody := fmt.Sprintf(`Hallo %s %s,

Du hast eine Anfrage zum Zurücksetzen Deines Passworts gestellt.

Um Dein Passwort zurückzusetzen, öffne bitte den folgenden Link in Deinem Browser:

%s

Alternativ kannst Du diesen Link kopieren und in Deinen Browser einfügen:

%s

WICHTIG: Dieser Link ist nur bis zum %s gültig.

Falls Du diese Anfrage nicht gestellt hast, ignoriere diese E-Mail einfach.`,
		data.User.FirstName, data.User.LastName,
		data.ResetURL, data.ResetURL,
		expiryBerlin.Format("02.01.2006 um 15:04 Uhr"))

	htmlBody := fmt.Sprintf(`        <p>Hallo %s %s,</p>
        <p>Du hast eine Anfrage zum Zurücksetzen Deines Passworts gestellt.</p>
        <p>Um Dein Passwort zurückzusetzen, klicke bitte auf den folgenden Button:</p>
        <p style="text-align: center;">
            <a href="%s" style="display: inline-block; background-color: #2c5aa0; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 20px 0"><span style="color: white">Passwort zurücksetzen</span></a>
        </p>
        <div style="margin-top: 20px; padding: 15px; background-color: #e3f2fd; border-radius: 5px; font-size: 14px; word-break: break-all">
            <strong>Alternativ kannst Du diesen Link kopieren und in Deinen Browser einfügen:</strong><br>
            <a href="%s">%s</a>
        </div>
        <p><strong>Wichtig:</strong> Dieser Link ist nur bis zum <strong>%s</strong> gültig.</p>
        <p>Falls Du diese Anfrage nicht gestellt hast, ignoriere diese E-Mail einfach.</p>`,
		data.User.FirstName, data.User.LastName,
		data.ResetURL, data.ResetURL, data.ResetURL,
		expiryBerlin.Format("02.01.2006 um 15:04 Uhr"))

	fullText := s.wrapText(subject, textBody)
	fullHTML := s.wrapHTML(subject, htmlBody)

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
		fullText,
		boundary,
		s.quotedPrintableEncode(fullHTML),
		boundary)
}

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

func (s *Service) quotedPrintableEncode(input string) string {
	var buf bytes.Buffer
	writer := quotedprintable.NewWriter(&buf)

	_, err := writer.Write([]byte(input))
	if err != nil {
		return input
	}

	err = writer.Close()
	if err != nil {
		return input
	}

	return buf.String()
}
