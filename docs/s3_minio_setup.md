# S3 / MinIO Setup – UserService

Deze documentatie beschrijft hoe **MinIO (S3-compatible storage)** is ingericht en geïntegreerd met de **UserService** binnen OpenShift (en lokaal via Docker Compose).

MinIO wordt gebruikt voor het opslaan van **user media**, zoals de **profielfoto**. De backend slaat alleen de **public URL** op in de database; de daadwerkelijke upload gebeurt via een **presigned PUT URL**.

---

## Overzicht

- MinIO draait als **aparte service** (S3-compatible)
- UserService gebruikt MinIO voor:
  - Bucket beheer (bestaan/aanmaken)
  - Genereren van **presigned upload URLs** (tijdelijk geldig)
  - Public URL genereren en opslaan bij de user (`ProfilePhotoURL`)
- Upload flow is **client-side**:
  - Backend: maakt presigned URL + slaat public URL op
  - Frontend: doet de PUT upload naar MinIO

---

## Componenten

### 1. MinIO Deployment / Container

- **Image**: `minio/minio:latest`
- **API Port**: `9000`
- **Console Port**: `9001`
- **Command**: `server /data --console-address ":9001"`
- **Storage**: persistent volume (voor object data)

De MinIO credentials worden gezet via environment variables:
- `MINIO_ROOT_USER` = `S3_ACCESS_KEY`
- `MINIO_ROOT_PASSWORD` = `S3_SECRET_KEY`

---

## Environment Variables

De UserService verwacht onderstaande variabelen voor S3/MinIO:

- `S3_ENDPOINT`  
  Interne endpoint (service-to-service binnen cluster), bijv. `http://minio:9000`

- `S3_EXTERNAL_ENDPOINT` *(optioneel)*  
  Endpoint die gebruikt wordt om **presigned URLs** te maken die de client kan bereiken.  
  Als deze leeg is, gebruikt de service automatisch `S3_ENDPOINT`.

- `S3_ACCESS_KEY`  
  MinIO access key

- `S3_SECRET_KEY`  
  MinIO secret key

- `S3_BUCKET`  
  Bucket naam waar user media in komt (bijv. `user-media`)

- `S3_PUBLIC_BASE_URL`  
  Base URL voor publieke object links (bijv. MinIO route of gateway URL)

---

## Hoe de UserService S3/MinIO gebruikt (code flow)

### 1. Twee MinIO clients (internal + presign)

In `storage.NewS3()` worden **twee MinIO clients** gebouwd:

- `internal` client  
  Wordt gebruikt door de backend binnen het cluster (bucket checks/maak bucket).

- `presign` client  
  Wordt gebruikt om **presigned PUT URLs** te genereren.  
  Dit is bewust apart, zodat de presigned URL altijd een endpoint bevat dat de **frontend echt kan bereiken** (bijv. via OpenShift route).

Waarom dit nodig is:
- In OpenShift is de interne service-hostname vaak niet bereikbaar vanuit de browser.
- Daarom kan de presigned URL een **externe hostname** nodig hebben.

---

### 2. Endpoint normalisatie

`normalizeEndpoint()` zorgt dat endpoints flexibel kunnen worden meegegeven:

- Met schema: `http://...` of `https://...` → secure wordt automatisch bepaald
- Zonder schema: `minio:9000` → secure = `false`
- Een endpoint met een path (zoals `/api`) wordt geweigerd (moet puur host:port zijn)

---

### 3. Bucket automatisch aanmaken

Bij startup doet de service een `EnsureBucket()`:

1. Check of bucket bestaat
2. Als niet: maak bucket aan

In `main.go` gebeurt dit met retries (handig omdat MinIO soms later klaar is dan de app):

- meerdere pogingen met sleep ertussen
- zodra bucket ready is gaat de service door

---

## Upload flow (Profielfoto)

Endpoint: `POST /users/me/profile-photo/presign`

1. Client stuurt metadata zoals:
   - `content_type` (bijv. `image/jpeg`)
   - `ext` (bijv. `jpg`)
2. Backend:
   - valideert allowlist (jpg/jpeg/png/webp)
   - maakt object key: `users/<sub>/profile.<ext>`
   - genereert presigned PUT URL (tijdelijk geldig, bijv. 10 min)
   - maakt public URL: `S3_PUBLIC_BASE_URL + "/" + key`
   - slaat public URL op bij user in DB (`ProfilePhotoURL`)
3. Client uploadt vervolgens de foto via:
   - `PUT <upload_url>` met als body de file bytes

Belangrijk:
- Backend uploadt de file **niet** zelf.
- Backend regelt alleen presign + opslaan van de link.

---

## Docker Compose (local dev)

In `docker-compose.yml` draait MinIO als service:

- poorten:
  - `9000:9000` (S3 API)
  - `9001:9001` (console UI)
- credentials komen uit `.env` (`S3_ACCESS_KEY` / `S3_SECRET_KEY`)
- data wordt opgeslagen in volume `minio_data`

De UserService dependt op `minio` zodat MinIO eerst beschikbaar is.

---

## Beveiliging

Voor productie/serieuze omgevingen:

- zet credentials in **Kubernetes Secrets** (niet plaintext env vars)
- gebruik een externe route/gateway voor `S3_PUBLIC_BASE_URL`
- beperk bucket policy (liefst private + alleen presigned writes/reads waar nodig)
- draai MinIO niet met default credentials

---

## Samenvatting

- MinIO = object storage voor user media
- UserService:
  - maakt bucket aan bij startup
  - genereert presigned upload URLs (client-side upload)
  - bewaart alleen public URL in DB
- Externe vs interne endpoint is belangrijk voor presigned URLs in OpenShift

---
