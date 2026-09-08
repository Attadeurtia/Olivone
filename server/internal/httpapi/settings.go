package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/attadeurtia/olivone/internal/users"
)

// settingsView est la représentation renvoyée au client : elle ne contient
// JAMAIS de secret, seulement des drapeaux « défini / non défini ».
type settingsView struct {
	DefaultRecipient     string `json:"default_recipient"`
	MistralModel         string `json:"mistral_model"`
	MistralKeySet        bool   `json:"mistral_key_set"`
	SMTPHost             string `json:"smtp_host"`
	SMTPPort             int    `json:"smtp_port"`
	SMTPUsername         string `json:"smtp_username"`
	SMTPFrom             string `json:"smtp_from"`
	SMTPPasswordSet      bool   `json:"smtp_password_set"`
	IMAPHost             string `json:"imap_host"`
	IMAPPort             int    `json:"imap_port"`
	IMAPUsername         string `json:"imap_username"`
	IMAPPasswordSet      bool   `json:"imap_password_set"`
	AutoSendEnabled      bool   `json:"auto_send_enabled"`
	FollowupEnabled      bool   `json:"followup_enabled"`
	FollowupIntervalDays int    `json:"followup_interval_days"`
	ThemePref            string `json:"theme_pref"`
	ProfileMD            string `json:"profile_md"`
	Signature            string `json:"signature"`
	CVSet                bool   `json:"cv_set"`
}

func toSettingsView(st users.Settings, cvSet bool) settingsView {
	return settingsView{
		DefaultRecipient:     st.DefaultRecipient,
		MistralModel:         st.MistralModel,
		MistralKeySet:        st.MistralAPIKeyEnc != "",
		SMTPHost:             st.SMTPHost,
		SMTPPort:             st.SMTPPort,
		SMTPUsername:         st.SMTPUsername,
		SMTPFrom:             st.SMTPFrom,
		SMTPPasswordSet:      st.SMTPPasswordEnc != "",
		IMAPHost:             st.IMAPHost,
		IMAPPort:             st.IMAPPort,
		IMAPUsername:         st.IMAPUsername,
		IMAPPasswordSet:      st.IMAPPasswordEnc != "",
		AutoSendEnabled:      st.AutoSendEnabled,
		FollowupEnabled:      st.FollowupEnabled,
		FollowupIntervalDays: st.FollowupIntervalDays,
		ThemePref:            st.ThemePref,
		ProfileMD:            st.ProfileMD,
		Signature:            st.Signature,
		CVSet:                cvSet,
	}
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	st, err := s.users.GetSettings(u.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "erreur interne"})
		return
	}
	writeJSON(w, http.StatusOK, toSettingsView(st, s.apps.HasCV(u.ID)))
}

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	var in users.SettingsInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "requête invalide"})
		return
	}
	if err := s.users.UpdateSettings(u.ID, in); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "erreur interne"})
		return
	}
	st, err := s.users.GetSettings(u.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "erreur interne"})
		return
	}
	writeJSON(w, http.StatusOK, toSettingsView(st, s.apps.HasCV(u.ID)))
}
