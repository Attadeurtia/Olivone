-- Migration initiale : schéma complet d'Olivone.
-- Les colonnes *_enc contiennent des secrets chiffrés (AES-GCM) au repos.

-- Utilisateurs (multi-utilisateurs, comptes gérés par un admin).
CREATE TABLE users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    display_name  TEXT    NOT NULL DEFAULT '',
    is_admin      INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- Réglages par utilisateur : clé Mistral, e-mail (SMTP/IMAP), profil, options.
CREATE TABLE user_settings (
    user_id                INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    default_recipient      TEXT    NOT NULL DEFAULT '',
    mistral_api_key_enc    TEXT    NOT NULL DEFAULT '',
    mistral_model          TEXT    NOT NULL DEFAULT 'mistral-small-latest',
    smtp_host              TEXT    NOT NULL DEFAULT '',
    smtp_port              INTEGER NOT NULL DEFAULT 587,
    smtp_username          TEXT    NOT NULL DEFAULT '',
    smtp_password_enc      TEXT    NOT NULL DEFAULT '',
    smtp_from              TEXT    NOT NULL DEFAULT '',
    imap_host              TEXT    NOT NULL DEFAULT '',
    imap_port              INTEGER NOT NULL DEFAULT 993,
    imap_username          TEXT    NOT NULL DEFAULT '',
    imap_password_enc      TEXT    NOT NULL DEFAULT '',
    auto_send_enabled      INTEGER NOT NULL DEFAULT 0,   -- OFF par défaut (règle 12)
    followup_enabled       INTEGER NOT NULL DEFAULT 0,   -- OFF par défaut (règle 13)
    followup_interval_days INTEGER NOT NULL DEFAULT 14,
    theme_pref             TEXT    NOT NULL DEFAULT 'auto',
    profile_md             TEXT    NOT NULL DEFAULT '',
    signature              TEXT    NOT NULL DEFAULT ''
);

-- Templates de prompt personnels (éditables dans l'UI en Markdown).
CREATE TABLE templates (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT    NOT NULL DEFAULT 'default',
    content_md TEXT    NOT NULL DEFAULT '',
    is_default INTEGER NOT NULL DEFAULT 1,
    updated_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_templates_user ON templates(user_id);

-- Candidatures.
CREATE TABLE applications (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id          INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_title        TEXT    NOT NULL DEFAULT '',
    company          TEXT    NOT NULL DEFAULT '',
    recipient_email  TEXT    NOT NULL DEFAULT '',
    source_type      TEXT    NOT NULL DEFAULT 'text_paste', -- pdf_upload|text_paste|external|...
    status           TEXT    NOT NULL DEFAULT 'brouillon',
    offer_text       TEXT    NOT NULL DEFAULT '',
    offer_pdf_path   TEXT    NOT NULL DEFAULT '',
    letter_md_path   TEXT    NOT NULL DEFAULT '',
    letter_pdf_path  TEXT    NOT NULL DEFAULT '',
    cv_path          TEXT    NOT NULL DEFAULT '',
    notes            TEXT    NOT NULL DEFAULT '',
    external         INTEGER NOT NULL DEFAULT 0,
    created_at       TEXT    NOT NULL DEFAULT (datetime('now')),
    sent_at          TEXT,
    last_followup_at TEXT,
    next_followup_at TEXT
);
CREATE INDEX idx_applications_user   ON applications(user_id);
CREATE INDEX idx_applications_status ON applications(status);

-- Messages e-mail liés à une candidature (pour matcher les réponses).
CREATE TABLE messages (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    application_id INTEGER NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    direction      TEXT    NOT NULL,                 -- outgoing|incoming
    message_id     TEXT    NOT NULL DEFAULT '',
    in_reply_to    TEXT    NOT NULL DEFAULT '',
    subject        TEXT    NOT NULL DEFAULT '',
    from_addr      TEXT    NOT NULL DEFAULT '',
    to_addr        TEXT    NOT NULL DEFAULT '',
    at             TEXT    NOT NULL DEFAULT (datetime('now')),
    snippet        TEXT    NOT NULL DEFAULT '',
    imap_uid       INTEGER
);
CREATE INDEX idx_messages_app   ON messages(application_id);
CREATE INDEX idx_messages_msgid ON messages(message_id);

-- Sessions serveur (cookie -> hash du jeton).
CREATE TABLE sessions (
    token_hash TEXT    PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT    NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT    NOT NULL,
    last_seen  TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_sessions_user ON sessions(user_id);
