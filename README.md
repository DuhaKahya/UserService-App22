# UserService – Documentatie

Deze service is verantwoordelijk voor **gebruikersbeheer** binnen het platform.  
Keycloak regelt **authenticatie** (tokens/identity), en onze eigen database regelt **profieldata + preferences** (zoals notificaties, interesses, discovery settings en profielfoto-URL).

---

## 1. Architectuur in het kort

- **Keycloak** = wie ben je? (identity, access/refresh tokens, `sub`)
- **UserService DB** = wat is jouw profiel en jouw instellingen?
- De koppeling gebeurt via **Keycloak `sub` (KeycloakID)** die we opslaan bij de user in onze database.

---

## 2. Authenticatie & autorisatie

### JWT verificatie (middleware)
Voor alle routes onder `/users/me/*`:
1. Client stuurt `Authorization: Bearer <token>`
2. Middleware decodeert het token via Keycloak
3. `sub` wordt uitgelezen en als `user_id` in de Gin context gezet
4. Controllers gebruiken die `sub` om de juiste user te laden

Waarom dit belangrijk is:
- De client hoeft **geen user-id** mee te sturen.
- Je kan alleen je **eigen** gegevens ophalen/wijzigen via je token.

---

## 3. Registratie (POST `/users/register`)

Flow:
1. Input wordt gevalideerd
2. Wachtwoord wordt gehasht voor opslag in de lokale DB
3. User wordt opgeslagen in onze database
4. Daarna wordt dezelfde user aangemaakt in Keycloak met het **plain** wachtwoord
5. Keycloak geeft een **UUID** terug (`sub`)
6. Dit wordt opgeslagen als `KeycloakID` in onze `users` tabel

Doel:
- Keycloak beheert login/tokens
- Onze DB beheert alle extra user data + instellingen
- `KeycloakID` is de sleutel tussen beide

---

## 4. Login (POST `/auth/login`) + Refresh (POST `/auth/refresh`)

### Login flow
1. User wordt opgezocht via e-mail in onze database
2. Wachtwoord hash wordt gecheckt (bcrypt)
3. Daarna vraagt de service bij Keycloak een **access token** + **refresh token** op
4. Token is vereist voor beveiligde routes

### Refresh flow
- Met de refresh token kan de client een nieuwe access token ophalen.

---

## 5. Profiel (PUT `/users/me`)

1. Token → `sub` via middleware
2. User wordt geladen via `KeycloakID`
3. Update wordt toegepast op de user in onze database
4. Response geeft user terug (zonder wachtwoord)

---

## 6. Notificatievoorkeuren

### Update (PUT `/users/me/notification-preferences`)
1. Token → `sub`
2. User wordt geladen via `KeycloakID`
3. JSON input wordt gevalideerd
4. **System alerts** zijn “locked” en mogen niet worden aangepast (wordt expliciet geblokkeerd)
5. Settings worden geüpdatet en teruggegeven (zonder system alert in public response)

### Intern (GET `/internal/users/{email}/notification-preferences`)
- Voor andere services, beveiligd met `X-Service-Token`
- Geeft uitgebreider antwoord terug (incl. system alerts)

---

## 7. Interesses

- Bij startup wordt een master lijst interests **geseed** in de database.
- User endpoints:
  - GET `/users/me/interests`
  - PUT `/users/me/interests` (array `{id, value}`)
- Interne endpoint (service-to-service):
  - GET `/internal/users/{email}/interests` (met `X-Service-Token`)

---

## 8. Discovery preferences (radius)

- User kan zijn **zoekradius in km** beheren:
  - GET `/users/me/discovery-preferences`
  - PUT `/users/me/discovery-preferences`
- Als er nog geen record bestaat, worden defaults teruggegeven.
- Interne endpoint:
  - GET `/internal/users/{email}/discovery-preferences` (met `X-Service-Token`)

---

## 9. Password reset flow

### Forgot (POST `/auth/forgot-password`)
1. Email komt binnen
2. Service maakt (indien mogelijk) een reset token aan
3. Er wordt een **system alert** verstuurd naar de NotificationService (service-to-service)
4. Endpoint returnt altijd een “ok” boodschap (voorkomt user enumeration)

### Reset (POST `/auth/reset-password`)
1. Token + nieuw wachtwoord
2. Token wordt gevalideerd
3. Wachtwoord wordt aangepast (incl. Keycloak reset)

---

## 10. Profielfoto upload (Presigned URL naar MinIO/S3)

Endpoint: POST `/users/me/profile-photo/presign`

Flow:
1. Client stuurt content type + ext (bijv. `image/jpeg`, `jpg`)
2. Service genereert een **presigned PUT upload URL** voor MinIO/S3
3. Service maakt ook een **public URL**
4. Die public URL wordt direct opgeslagen in de user (`ProfilePhotoURL`)
5. Client uploadt daarna de afbeelding via **HTTP PUT** naar de presigned `upload_url`

Belangrijk:
- Upload zelf gebeurt dus **client-side** met de presigned URL.
- De backend regelt alleen presign + opslaan van de public link.

---

## 11. Internal service-to-service endpoints

Alles onder `/internal/*` vereist:
- Header: `X-Service-Token`

Voorbeelden:
- GET `/internal/users/:email`
- GET `/internal/users/:email/notification-preferences`
- GET `/internal/users/:email/interests`
- GET `/internal/users/:email/discovery-preferences`

---

## 11.5. Badges (Achievements)

De UserService ondersteunt een eenvoudig badgesysteem.  
Badges worden opgeslagen in twee tabellen (`badges` en `user_badges`) en worden automatisch **geseed** bij startup.

### Automatisch toegekende badges (deze zijn gevestigd in de UserService, vanuit andere services komen nog meer badges)
- `profile_photo_uploaded` – bij upload van een profielfoto  
- `profile_complete` – wanneer alle verplichte profielvelden zijn ingevuld

Dit gebeurt automatisch in:
- `PUT /users/me`
- `POST /users/me/profile-photo`

### Badge endpoints
- **GET `/users/me/badges`** – geeft alle badges van de ingelogde gebruiker  
- **GET `/users/id/{id}/badges`** – badges opvragen via user ID  
- **POST `/internal/badges/award`** – interne endpoint voor andere services (vereist `X-Service-Token`)

Alle badges kunnen slechts **één keer** worden verdiend; duplicaten worden voorkomen in de database.

---

## 12. Monitoring (Prometheus)

- Middleware meet request metrics (counts/duration/outcomes)
- Metrics endpoint wordt aangeboden via een aparte metrics server op `/metrics`

---
