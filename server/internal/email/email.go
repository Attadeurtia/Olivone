// Package email envoie des e-mails via SMTP avec pièces jointes. Il gère aussi
// bien un serveur de test local (Mailpit, sans TLS ni authentification) qu'un
// serveur réel (STARTTLS ou SSL + authentification).
package email

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/wneessen/go-mail"
)

// Config décrit le serveur SMTP de l'utilisateur.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Message décrit l'e-mail à envoyer.
type Message struct {
	To          string
	Subject     string
	Body        string
	Attachments []string // chemins de fichiers
	MessageID   string   // valeur sans chevrons ; générée si vide
}

// NewMessageID produit un identifiant de message unique.
func NewMessageID(from string) string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	domain := "olivone.local"
	if i := strings.LastIndex(from, "@"); i >= 0 && i+1 < len(from) {
		domain = from[i+1:]
	}
	return fmt.Sprintf("%s@%s", hex.EncodeToString(b), domain)
}

// Send envoie l'e-mail et renvoie le Message-ID complet (avec chevrons) utilisé.
func Send(cfg Config, msg Message) (string, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return "", fmt.Errorf("serveur SMTP non configuré")
	}
	from := cfg.From
	if from == "" {
		from = cfg.Username
	}

	m := mail.NewMsg()
	if err := m.From(from); err != nil {
		return "", fmt.Errorf("expéditeur invalide (%q): %w", from, err)
	}
	if err := m.To(msg.To); err != nil {
		return "", fmt.Errorf("destinataire invalide (%q): %w", msg.To, err)
	}
	m.Subject(msg.Subject)
	m.SetBodyString(mail.TypeTextPlain, msg.Body)

	msgID := msg.MessageID
	if msgID == "" {
		msgID = NewMessageID(from)
	}
	m.SetMessageIDWithValue(msgID)

	for _, path := range msg.Attachments {
		if path == "" {
			continue
		}
		m.AttachFile(path) // les erreurs de lecture remontent à l'envoi
	}

	port := cfg.Port
	if port == 0 {
		port = 587
	}
	opts := []mail.Option{mail.WithPort(port), mail.WithTimeout(30 * time.Second)}
	if strings.TrimSpace(cfg.Username) != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(cfg.Username),
			mail.WithPassword(cfg.Password),
		)
		if port == 465 {
			opts = append(opts, mail.WithSSL())
		}
	} else {
		// Serveur local de test (ex. Mailpit) : ni TLS ni authentification.
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS))
	}

	c, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return "", err
	}
	if err := c.DialAndSend(m); err != nil {
		return "", err
	}
	return "<" + msgID + ">", nil
}
