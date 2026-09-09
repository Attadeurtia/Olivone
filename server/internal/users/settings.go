package users

import (
	"strings"

	"github.com/attadeurtia/olivone/internal/crypto"
)

// Settings reflète la ligne user_settings. Les champs *Enc restent chiffrés ;
// ils ne doivent jamais être renvoyés tels quels au client.
type Settings struct {
	DefaultRecipient     string
	MistralAPIKeyEnc     string
	MistralModel         string
	SMTPHost             string
	SMTPPort             int
	SMTPUsername         string
	SMTPPasswordEnc      string
	SMTPFrom             string
	IMAPHost             string
	IMAPPort             int
	IMAPUsername         string
	IMAPPasswordEnc      string
	AutoSendEnabled      bool
	FollowupEnabled      bool
	FollowupIntervalDays int
	ThemePref            string
	ProfileMD            string
	Signature            string
	SenderName           string
	SenderAddress        string
	SenderPhone          string
	SenderCity           string
}

// GetSettings lit les réglages d'un utilisateur.
func (s *Service) GetSettings(userID int64) (Settings, error) {
	var st Settings
	var autoSend, followup int
	err := s.DB.QueryRow(`SELECT
		default_recipient, mistral_api_key_enc, mistral_model,
		smtp_host, smtp_port, smtp_username, smtp_password_enc, smtp_from,
		imap_host, imap_port, imap_username, imap_password_enc,
		auto_send_enabled, followup_enabled, followup_interval_days,
		theme_pref, profile_md, signature,
		sender_name, sender_address, sender_phone, sender_city
		FROM user_settings WHERE user_id = ?`, userID).Scan(
		&st.DefaultRecipient, &st.MistralAPIKeyEnc, &st.MistralModel,
		&st.SMTPHost, &st.SMTPPort, &st.SMTPUsername, &st.SMTPPasswordEnc, &st.SMTPFrom,
		&st.IMAPHost, &st.IMAPPort, &st.IMAPUsername, &st.IMAPPasswordEnc,
		&autoSend, &followup, &st.FollowupIntervalDays,
		&st.ThemePref, &st.ProfileMD, &st.Signature,
		&st.SenderName, &st.SenderAddress, &st.SenderPhone, &st.SenderCity,
	)
	st.AutoSendEnabled = autoSend == 1
	st.FollowupEnabled = followup == 1
	return st, err
}

// DecryptSecret déchiffre un champ *Enc (usage interne : M2/M3).
func (s *Service) DecryptSecret(enc string) (string, error) {
	return crypto.DecryptString(s.Key, enc)
}

// SettingsInput : champs modifiables. Un pointeur nil signifie « ne pas
// changer » ; une valeur (même vide) écrase. Les secrets fournis sont chiffrés.
type SettingsInput struct {
	DefaultRecipient     *string `json:"default_recipient"`
	MistralAPIKey        *string `json:"mistral_api_key"` // secret
	MistralModel         *string `json:"mistral_model"`
	SMTPHost             *string `json:"smtp_host"`
	SMTPPort             *int    `json:"smtp_port"`
	SMTPUsername         *string `json:"smtp_username"`
	SMTPPassword         *string `json:"smtp_password"` // secret
	SMTPFrom             *string `json:"smtp_from"`
	IMAPHost             *string `json:"imap_host"`
	IMAPPort             *int    `json:"imap_port"`
	IMAPUsername         *string `json:"imap_username"`
	IMAPPassword         *string `json:"imap_password"` // secret
	AutoSendEnabled      *bool   `json:"auto_send_enabled"`
	FollowupEnabled      *bool   `json:"followup_enabled"`
	FollowupIntervalDays *int    `json:"followup_interval_days"`
	ThemePref            *string `json:"theme_pref"`
	ProfileMD            *string `json:"profile_md"`
	Signature            *string `json:"signature"`
	SenderName           *string `json:"sender_name"`
	SenderAddress        *string `json:"sender_address"`
	SenderPhone          *string `json:"sender_phone"`
	SenderCity           *string `json:"sender_city"`
}

// UpdateSettings applique les champs fournis (UPDATE partiel).
func (s *Service) UpdateSettings(userID int64, in SettingsInput) error {
	set := []string{}
	args := []any{}
	add := func(col string, val any) {
		set = append(set, col+" = ?")
		args = append(args, val)
	}

	if in.DefaultRecipient != nil {
		add("default_recipient", *in.DefaultRecipient)
	}
	if in.MistralModel != nil {
		add("mistral_model", *in.MistralModel)
	}
	if in.SMTPHost != nil {
		add("smtp_host", *in.SMTPHost)
	}
	if in.SMTPPort != nil {
		add("smtp_port", *in.SMTPPort)
	}
	if in.SMTPUsername != nil {
		add("smtp_username", *in.SMTPUsername)
	}
	if in.SMTPFrom != nil {
		add("smtp_from", *in.SMTPFrom)
	}
	if in.IMAPHost != nil {
		add("imap_host", *in.IMAPHost)
	}
	if in.IMAPPort != nil {
		add("imap_port", *in.IMAPPort)
	}
	if in.IMAPUsername != nil {
		add("imap_username", *in.IMAPUsername)
	}
	if in.AutoSendEnabled != nil {
		add("auto_send_enabled", boolToInt(*in.AutoSendEnabled))
	}
	if in.FollowupEnabled != nil {
		add("followup_enabled", boolToInt(*in.FollowupEnabled))
	}
	if in.FollowupIntervalDays != nil {
		add("followup_interval_days", *in.FollowupIntervalDays)
	}
	if in.ThemePref != nil {
		add("theme_pref", *in.ThemePref)
	}
	if in.ProfileMD != nil {
		add("profile_md", *in.ProfileMD)
	}
	if in.Signature != nil {
		add("signature", *in.Signature)
	}
	if in.SenderName != nil {
		add("sender_name", *in.SenderName)
	}
	if in.SenderAddress != nil {
		add("sender_address", *in.SenderAddress)
	}
	if in.SenderPhone != nil {
		add("sender_phone", *in.SenderPhone)
	}
	if in.SenderCity != nil {
		add("sender_city", *in.SenderCity)
	}

	// Secrets : chiffrés avant stockage.
	if in.MistralAPIKey != nil {
		enc, err := crypto.EncryptString(s.Key, *in.MistralAPIKey)
		if err != nil {
			return err
		}
		add("mistral_api_key_enc", enc)
	}
	if in.SMTPPassword != nil {
		enc, err := crypto.EncryptString(s.Key, *in.SMTPPassword)
		if err != nil {
			return err
		}
		add("smtp_password_enc", enc)
	}
	if in.IMAPPassword != nil {
		enc, err := crypto.EncryptString(s.Key, *in.IMAPPassword)
		if err != nil {
			return err
		}
		add("imap_password_enc", enc)
	}

	if len(set) == 0 {
		return nil
	}
	args = append(args, userID)
	_, err := s.DB.Exec(
		"UPDATE user_settings SET "+strings.Join(set, ", ")+" WHERE user_id = ?", args...,
	)
	return err
}
