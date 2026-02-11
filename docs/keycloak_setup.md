# Keycloak Setup – UserService

Deze documentatie beschrijft hoe **Keycloak** is ingericht en geïntegreerd met de **UserService** binnen OpenShift.

Keycloak verzorgt **authenticatie en autorisatie**, terwijl de UserService verantwoordelijk is voor **profieldata en gebruikersinstellingen**.

---

## Overzicht

- Keycloak draait als **aparte deployment** binnen OpenShift
- Authenticatie gebeurt via **OpenID Connect (OIDC)**
- UserService valideert JWT tokens en gebruikt de **Keycloak `sub`** als koppeling naar de lokale gebruiker
- Keycloak en UserService delen dezelfde PostgreSQL database (gescheiden tabellen)

---

## Componenten

### 1. Keycloak Deployment

Keycloak draait op basis van de officiële image:

- **Image**: `quay.io/keycloak/keycloak:26.0`
- **Mode**: `start-dev --import-realm`
- **Poort**: 8080
- **Database**: PostgreSQL
- **Health checks**:
  - Liveness probe op `/`
  - Readiness probe op `/realms/master`

De deployment is opgenomen in de UserService kustomization en wordt automatisch mee uitgerold.

---

### 2. Keycloak Service

- **Type**: ClusterIP
- Expose’t Keycloak intern op poort 8080
- Wordt gebruikt door de UserService voor:
  - Token verificatie
  - Admin acties (user aanmaken / wachtwoord reset)

---

### 3. Keycloak Route

- OpenShift Route met **edge TLS termination**
- Geeft externe toegang tot:
  - Keycloak Admin Console
  - OpenID Connect endpoints

---

## Environment Variables

### Keycloak Container

Belangrijkste configuratie:

- `KC_BOOTSTRAP_ADMIN_USERNAME`: admin  
- `KC_BOOTSTRAP_ADMIN_PASSWORD`: admin 
- `KC_DB`: postgres  
- `KC_DB_URL`: `jdbc:postgresql://userservice-postgresql:5431/userservice`  
- `KC_DB_USERNAME`: postgres  
- `KC_DB_PASSWORD`: postgres
- `KC_PROXY`: edge  
- `KC_HTTP_ENABLED`: true  
- `KC_HOSTNAME_STRICT`: false  

---

### UserService Container

De UserService communiceert met Keycloak via:

- `KEYCLOAK_URL`: externe Keycloak route
- `KEYCLOAK_REALM`: `simpleslideshow`
- `KEYCLOAK_CLIENT_ID`: `simpleslideshow-client`
- `KEYCLOAK_CLIENT_SECRET`: ***secret***
- `KEYCLOAK_ADMIN_USER`: admin
- `KEYCLOAK_ADMIN_PASS`: admin

---

## Realm & Client configuratie

Keycloak is ingericht met:

### Realm
- **Naam**: `simpleslideshow`
- Bevat alle gebruikers van het platform

### Client
- **Client ID**: `simpleslideshow-client`
- **Protocol**: OpenID Connect
- **Type**: Confidential
- Wordt gebruikt door:
  - UserService login
  - Token refresh
  - JWT validatie

De client secret wordt gebruikt door de UserService om tokens op te vragen.

---

## Gebruikersbeheer

Gebruikers kunnen worden aangemaakt via:

- Keycloak Admin Console (handmatig)
- UserService API (`/users/register`)

Bij registratie via de UserService:
1. User wordt aangemaakt in de lokale database
2. User wordt aangemaakt in Keycloak
3. Keycloak `sub` (UUID) wordt opgeslagen als `KeycloakID`

Deze ID vormt de **koppeling tussen Keycloak en de UserService database**.

---

## Database

- **PostgreSQL** wordt gedeeld met de UserService
- Keycloak gebruikt eigen tabellen (prefix `KC_`)
- UserService en Keycloak blijven logisch gescheiden

---

## Beveiliging

De huidige configuratie is geschikt voor **test/acceptatie**.

Voor productie:
- Gebruik Kubernetes Secrets
- Verander alle default credentials
- Gebruik `start` i.p.v. `start-dev`
- Zet `KC_HOSTNAME_STRICT_HTTPS=true`
- Beperk redirect URI’s en CORS origins

---

## Integratie met andere services

Andere services kunnen Keycloak gebruiken door:

- JWT tokens te accepteren
- Dezelfde `KEYCLOAK_REALM` en `KEYCLOAK_URL` te gebruiken
- Tokens te valideren via Keycloak public keys

---

## Samenvatting

- Keycloak regelt **wie je bent**
- UserService regelt **wat van jou is**
- De koppeling loopt via de **Keycloak `sub`**
- OpenShift verzorgt routing, TLS en isolatie
- Alles is voorbereid op uitbreiding met meerdere microservices

---
