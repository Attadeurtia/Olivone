-- Coordonnées d'expéditeur affichées dans l'en-tête de la lettre (encart
-- haut-gauche). Toutes optionnelles ; le nom retombe sur le nom d'affichage du
-- compte et l'e-mail sur l'adresse du compte lorsqu'elles sont vides.
ALTER TABLE user_settings ADD COLUMN sender_name    TEXT NOT NULL DEFAULT '';
ALTER TABLE user_settings ADD COLUMN sender_address TEXT NOT NULL DEFAULT '';
ALTER TABLE user_settings ADD COLUMN sender_phone   TEXT NOT NULL DEFAULT '';
ALTER TABLE user_settings ADD COLUMN sender_city    TEXT NOT NULL DEFAULT '';
