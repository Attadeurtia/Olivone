// Package applications gère le cycle de vie d'une candidature : création avec
// génération de la lettre (Markdown + PDF), régénération, remplacement du
// Markdown (version corrigée), statut, et suppression. Chaque opération est
// filtrée par utilisateur.
package applications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/attadeurtia/olivone/internal/email"
	"github.com/attadeurtia/olivone/internal/generation"
	"github.com/attadeurtia/olivone/internal/pdftext"
	"github.com/attadeurtia/olivone/internal/prompt"
	"github.com/attadeurtia/olivone/internal/typst"
	"github.com/attadeurtia/olivone/internal/users"
)

// ErrNoMistralKey : la clé Mistral n'est pas configurée.
var ErrNoMistralKey = errors.New("clé API Mistral non configurée dans les réglages")

// ErrNotFound : candidature absente (ou n'appartenant pas à l'utilisateur).
var ErrNotFound = errors.New("candidature introuvable")

// ErrNoSMTP : configuration SMTP absente.
var ErrNoSMTP = errors.New("serveur SMTP non configuré dans les réglages")

// Application est la vue renvoyée au client (sans chemins de fichiers bruts).
type Application struct {
	ID             int64   `json:"id"`
	JobTitle       string  `json:"job_title"`
	Company        string  `json:"company"`
	RecipientEmail string  `json:"recipient_email"`
	SourceType     string  `json:"source_type"`
	Status         string  `json:"status"`
	Notes          string  `json:"notes"`
	External       bool    `json:"external"`
	CreatedAt      string  `json:"created_at"`
	SentAt         *string `json:"sent_at"`
	HasLetter      bool    `json:"has_letter"`
	HasOfferPDF    bool    `json:"has_offer_pdf"`
}

// Service orchestre la persistance et la génération.
type Service struct {
	DB      *sql.DB
	Users   *users.Service
	Typst   *typst.Renderer
	DataDir string
}

// NewService crée le service candidatures.
func NewService(db *sql.DB, u *users.Service, t *typst.Renderer, dataDir string) *Service {
	return &Service{DB: db, Users: u, Typst: t, DataDir: dataDir}
}

// CreateInput regroupe les entrées de création.
type CreateInput struct {
	JobTitle       string
	Company        string
	RecipientEmail string
	OfferText      string // texte collé
	OfferPDFPath   string // chemin d'un PDF temporaire déjà enregistré, ou ""
}

func (s *Service) userDir(userID int64) string {
	return filepath.Join(s.DataDir, "users", strconv.FormatInt(userID, 10))
}

func (s *Service) letterDir(userID, appID int64) string {
	return filepath.Join(s.userDir(userID), "letters", strconv.FormatInt(appID, 10))
}

// Create crée la candidature puis génère la lettre (Markdown + PDF).
func (s *Service) Create(ctx context.Context, userID int64, in CreateInput) (Application, error) {
	settings, err := s.Users.GetSettings(userID)
	if err != nil {
		return Application{}, err
	}
	apiKey, err := s.Users.DecryptSecret(settings.MistralAPIKeyEnc)
	if err != nil {
		return Application{}, err
	}
	if strings.TrimSpace(apiKey) == "" {
		return Application{}, ErrNoMistralKey
	}

	offerText := strings.TrimSpace(in.OfferText)
	sourceType := "text_paste"
	if in.OfferPDFPath != "" {
		sourceType = "pdf_upload"
		if txt, e := pdftext.ExtractText(in.OfferPDFPath); e == nil {
			if t := strings.TrimSpace(txt); t != "" {
				offerText = t
			}
		}
	}
	if offerText == "" {
		return Application{}, errors.New("offre vide : le texte n'a pas pu être extrait du PDF, collez le texte de l'offre")
	}

	recipient := firstNonEmpty(in.RecipientEmail, settings.DefaultRecipient, "attadeurtia@vivaldi.net")

	res, err := s.DB.Exec(
		`INSERT INTO applications(user_id, job_title, company, recipient_email, source_type, status, offer_text)
		 VALUES(?,?,?,?,?, 'brouillon', ?)`,
		userID, in.JobTitle, in.Company, recipient, sourceType, offerText,
	)
	if err != nil {
		return Application{}, err
	}
	appID, _ := res.LastInsertId()

	// Déplace le PDF d'offre vers le stockage de l'utilisateur.
	if in.OfferPDFPath != "" {
		offersDir := filepath.Join(s.userDir(userID), "offers")
		if err := os.MkdirAll(offersDir, 0o750); err == nil {
			dst := filepath.Join(offersDir, fmt.Sprintf("offer-%d.pdf", appID))
			if err := moveFile(in.OfferPDFPath, dst); err == nil {
				_, _ = s.DB.Exec(`UPDATE applications SET offer_pdf_path=? WHERE id=?`, dst, appID)
			}
		}
	}

	if err := s.generateInto(ctx, userID, appID, apiKey, settings.MistralModel, prompt.Inputs{
		AppPrompt: mustAppPrompt(s.DataDir),
		Template:  templateOf(s.Users, userID),
		Profile:   settings.ProfileMD,
		OfferText: offerText,
		JobTitle:  in.JobTitle,
		Company:   in.Company,
		Recipient: recipient,
	}); err != nil {
		// Nettoyage : ligne + fichiers.
		_, _ = s.DB.Exec(`DELETE FROM applications WHERE id = ?`, appID)
		_ = os.RemoveAll(s.letterDir(userID, appID))
		return Application{}, err
	}
	return s.Get(userID, appID)
}

// Regenerate relance la génération à partir du texte d'offre stocké.
func (s *Service) Regenerate(ctx context.Context, userID, appID int64) (Application, error) {
	app, offerText, err := s.getRaw(userID, appID)
	if err != nil {
		return Application{}, err
	}
	settings, err := s.Users.GetSettings(userID)
	if err != nil {
		return Application{}, err
	}
	apiKey, _ := s.Users.DecryptSecret(settings.MistralAPIKeyEnc)
	if strings.TrimSpace(apiKey) == "" {
		return Application{}, ErrNoMistralKey
	}
	if err := s.generateInto(ctx, userID, appID, apiKey, settings.MistralModel, prompt.Inputs{
		AppPrompt: mustAppPrompt(s.DataDir),
		Template:  templateOf(s.Users, userID),
		Profile:   settings.ProfileMD,
		OfferText: offerText,
		JobTitle:  app.JobTitle,
		Company:   app.Company,
		Recipient: app.RecipientEmail,
	}); err != nil {
		return Application{}, err
	}
	return s.Get(userID, appID)
}

// UpdateLetter remplace le Markdown (version corrigée) puis re-rend le PDF.
// L'ancienne version est conservée en letter.bak.md (règle : ne pas perdre de données).
func (s *Service) UpdateLetter(ctx context.Context, userID, appID int64, markdown string) (Application, error) {
	if _, _, err := s.getRaw(userID, appID); err != nil {
		return Application{}, err
	}
	dir := s.letterDir(userID, appID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return Application{}, err
	}
	mdPath := filepath.Join(dir, "letter.md")
	if old, err := os.ReadFile(mdPath); err == nil {
		_ = os.WriteFile(filepath.Join(dir, "letter.bak.md"), old, 0o640)
	}
	if err := os.WriteFile(mdPath, []byte(markdown), 0o640); err != nil {
		return Application{}, err
	}
	pdfPath := filepath.Join(dir, "letter.pdf")
	if err := s.Typst.RenderLetter(ctx, markdown, pdfPath); err != nil {
		return Application{}, fmt.Errorf("rendu PDF: %w", err)
	}
	if _, err := s.DB.Exec(
		`UPDATE applications SET letter_md_path=?, letter_pdf_path=? WHERE id=? AND user_id=?`,
		mdPath, pdfPath, appID, userID,
	); err != nil {
		return Application{}, err
	}
	return s.Get(userID, appID)
}

// generateInto génère la lettre, écrit md + pdf, met à jour la ligne.
func (s *Service) generateInto(ctx context.Context, userID, appID int64, apiKey, model string, in prompt.Inputs) error {
	letterMD, err := generation.Generate(ctx, apiKey, model, in)
	if err != nil {
		return err
	}
	dir := s.letterDir(userID, appID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	mdPath := filepath.Join(dir, "letter.md")
	if err := os.WriteFile(mdPath, []byte(letterMD), 0o640); err != nil {
		return err
	}
	pdfPath := filepath.Join(dir, "letter.pdf")
	if err := s.Typst.RenderLetter(ctx, letterMD, pdfPath); err != nil {
		return fmt.Errorf("rendu PDF: %w", err)
	}
	_, err = s.DB.Exec(
		`UPDATE applications SET letter_md_path=?, letter_pdf_path=? WHERE id=? AND user_id=?`,
		mdPath, pdfPath, appID, userID,
	)
	return err
}

// List renvoie les candidatures de l'utilisateur (plus récentes d'abord).
func (s *Service) List(userID int64) ([]Application, error) {
	rows, err := s.DB.Query(selectCols+` WHERE user_id=? ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Application{}
	for rows.Next() {
		a, _, _, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Get renvoie une candidature.
func (s *Service) Get(userID, appID int64) (Application, error) {
	a, _, _, err := scanApp(s.DB.QueryRow(selectCols+` WHERE id=? AND user_id=?`, appID, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Application{}, ErrNotFound
	}
	return a, err
}

// getRaw renvoie aussi le texte de l'offre (usage interne).
func (s *Service) getRaw(userID, appID int64) (Application, string, error) {
	var offerText string
	a, _, _, err := scanAppWithOffer(s.DB.QueryRow(selectColsWithOffer+` WHERE id=? AND user_id=?`, appID, userID), &offerText)
	if errors.Is(err, sql.ErrNoRows) {
		return Application{}, "", ErrNotFound
	}
	return a, offerText, err
}

// LetterMarkdown renvoie le contenu Markdown de la lettre.
func (s *Service) LetterMarkdown(userID, appID int64) (string, error) {
	var path string
	err := s.DB.QueryRow(`SELECT letter_md_path FROM applications WHERE id=? AND user_id=?`, appID, userID).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", ErrNotFound
	}
	b, err := os.ReadFile(path)
	return string(b), err
}

// LetterPDFPath renvoie le chemin du PDF (pour le servir).
func (s *Service) LetterPDFPath(userID, appID int64) (string, error) {
	var path string
	err := s.DB.QueryRow(`SELECT letter_pdf_path FROM applications WHERE id=? AND user_id=?`, appID, userID).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", ErrNotFound
	}
	return path, nil
}

// UpdateStatus met à jour le statut et/ou les notes.
func (s *Service) UpdateStatus(userID, appID int64, status, notes string) (Application, error) {
	res, err := s.DB.Exec(
		`UPDATE applications SET status=COALESCE(NULLIF(?,''), status), notes=? WHERE id=? AND user_id=?`,
		status, notes, appID, userID,
	)
	if err != nil {
		return Application{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Application{}, ErrNotFound
	}
	return s.Get(userID, appID)
}

// Delete supprime la candidature et ses fichiers.
func (s *Service) Delete(userID, appID int64) error {
	var offerPDF string
	_ = s.DB.QueryRow(`SELECT offer_pdf_path FROM applications WHERE id=? AND user_id=?`, appID, userID).Scan(&offerPDF)
	res, err := s.DB.Exec(`DELETE FROM applications WHERE id=? AND user_id=?`, appID, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	_ = os.RemoveAll(s.letterDir(userID, appID))
	if offerPDF != "" {
		_ = os.Remove(offerPDF)
	}
	return nil
}

// --- CV (par utilisateur) ---------------------------------------------------

// CVPath renvoie le chemin du CV de l'utilisateur.
func (s *Service) CVPath(userID int64) string {
	return filepath.Join(s.userDir(userID), "cv.pdf")
}

// HasCV indique si l'utilisateur a un CV enregistré.
func (s *Service) HasCV(userID int64) bool {
	_, err := os.Stat(s.CVPath(userID))
	return err == nil
}

// SaveCV enregistre le CV (déplace srcPath vers l'emplacement définitif).
func (s *Service) SaveCV(userID int64, srcPath string) error {
	if err := os.MkdirAll(s.userDir(userID), 0o750); err != nil {
		return err
	}
	return moveFile(srcPath, s.CVPath(userID))
}

// DeleteCV supprime le CV.
func (s *Service) DeleteCV(userID int64) error {
	err := os.Remove(s.CVPath(userID))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// --- Envoi ------------------------------------------------------------------

// Send envoie la lettre par e-mail (lettre PDF + CV en pièces jointes),
// enregistre le message sortant, passe le statut à « envoyée » et calcule la
// prochaine relance si celle-ci est activée.
func (s *Service) Send(userID, appID int64, senderName string) (Application, error) {
	app, err := s.Get(userID, appID)
	if err != nil {
		return Application{}, err
	}
	if !app.HasLetter {
		return Application{}, errors.New("aucune lettre à envoyer pour cette candidature")
	}

	settings, err := s.Users.GetSettings(userID)
	if err != nil {
		return Application{}, err
	}
	if strings.TrimSpace(settings.SMTPHost) == "" {
		return Application{}, ErrNoSMTP
	}
	smtpPw, _ := s.Users.DecryptSecret(settings.SMTPPasswordEnc)

	pdfPath, err := s.LetterPDFPath(userID, appID)
	if err != nil {
		return Application{}, err
	}
	attachments := []string{pdfPath}
	hasCV := s.HasCV(userID)
	if hasCV {
		attachments = append(attachments, s.CVPath(userID))
	}

	subject := "Candidature"
	if app.JobTitle != "" {
		subject = "Candidature — " + app.JobTitle
	}
	from := firstNonEmpty(settings.SMTPFrom, settings.SMTPUsername)

	msgID, err := email.Send(email.Config{
		Host:     settings.SMTPHost,
		Port:     settings.SMTPPort,
		Username: settings.SMTPUsername,
		Password: smtpPw,
		From:     from,
	}, email.Message{
		To:          app.RecipientEmail,
		Subject:     subject,
		Body:        buildEmailBody(app.JobTitle, app.Company, senderName, hasCV),
		Attachments: attachments,
	})
	if err != nil {
		return Application{}, fmt.Errorf("envoi e-mail: %w", err)
	}

	_, _ = s.DB.Exec(
		`INSERT INTO messages(application_id, direction, message_id, subject, from_addr, to_addr)
		 VALUES(?, 'outgoing', ?, ?, ?, ?)`,
		appID, msgID, subject, from, app.RecipientEmail,
	)

	var nextFollowup sql.NullString
	if settings.FollowupEnabled {
		days := settings.FollowupIntervalDays
		if days <= 0 {
			days = 14
		}
		nextFollowup = sql.NullString{
			String: time.Now().AddDate(0, 0, days).UTC().Format(time.RFC3339),
			Valid:  true,
		}
	}
	if _, err := s.DB.Exec(
		`UPDATE applications SET status='envoyée', sent_at=datetime('now'), next_followup_at=? WHERE id=? AND user_id=?`,
		nextFollowup, appID, userID,
	); err != nil {
		return Application{}, err
	}
	return s.Get(userID, appID)
}

func buildEmailBody(jobTitle, company, senderName string, hasCV bool) string {
	var b strings.Builder
	b.WriteString("Madame, Monsieur,\n\n")
	b.WriteString("Veuillez trouver ci-joint ma lettre de motivation")
	if hasCV {
		b.WriteString(" ainsi que mon CV")
	}
	if jobTitle != "" {
		b.WriteString(" pour le poste de " + jobTitle)
		if company != "" {
			b.WriteString(" au sein de " + company)
		}
	}
	b.WriteString(".\n\nJe me tiens à votre disposition pour un entretien.\n\nCordialement,\n")
	b.WriteString(senderName)
	return b.String()
}

// --- helpers ----------------------------------------------------------------

const selectCols = `SELECT id, job_title, company, recipient_email, source_type, status, notes, external, created_at, sent_at, letter_pdf_path, offer_pdf_path FROM applications`

const selectColsWithOffer = `SELECT id, job_title, company, recipient_email, source_type, status, notes, external, created_at, sent_at, letter_pdf_path, offer_pdf_path, offer_text FROM applications`

type scanner interface {
	Scan(dest ...any) error
}

func scanApp(sc scanner) (Application, string, string, error) {
	var a Application
	var ext int
	var sentAt sql.NullString
	var letterPDF, offerPDF string
	err := sc.Scan(&a.ID, &a.JobTitle, &a.Company, &a.RecipientEmail, &a.SourceType,
		&a.Status, &a.Notes, &ext, &a.CreatedAt, &sentAt, &letterPDF, &offerPDF)
	if err != nil {
		return a, "", "", err
	}
	a.External = ext == 1
	if sentAt.Valid {
		a.SentAt = &sentAt.String
	}
	a.HasLetter = letterPDF != ""
	a.HasOfferPDF = offerPDF != ""
	return a, letterPDF, offerPDF, nil
}

func scanAppWithOffer(sc scanner, offerText *string) (Application, string, string, error) {
	var a Application
	var ext int
	var sentAt sql.NullString
	var letterPDF, offerPDF string
	err := sc.Scan(&a.ID, &a.JobTitle, &a.Company, &a.RecipientEmail, &a.SourceType,
		&a.Status, &a.Notes, &ext, &a.CreatedAt, &sentAt, &letterPDF, &offerPDF, offerText)
	if err != nil {
		return a, "", "", err
	}
	a.External = ext == 1
	if sentAt.Valid {
		a.SentAt = &sentAt.String
	}
	a.HasLetter = letterPDF != ""
	a.HasOfferPDF = offerPDF != ""
	return a, letterPDF, offerPDF, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func mustAppPrompt(dataDir string) string {
	p, _ := prompt.EnsureAppPrompt(dataDir)
	return p
}

func templateOf(u *users.Service, userID int64) string {
	t, _ := u.GetDefaultTemplate(userID)
	return t
}

func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Rename peut échouer entre systèmes de fichiers : copie + suppression.
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, in, 0o640); err != nil {
		return err
	}
	return os.Remove(src)
}
