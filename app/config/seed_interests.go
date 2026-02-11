package config

import (
	"log"

	"group1-userservice/app/models"
)

func SeedInterests() {
	keys := []string{
		"Gezondheidszorg en Welzijn",
		"Handel en Dienstverlening",
		"ICT",
		"Justitie, Veiligheid en Openbaar Bestuur",
		"Milieu en Agrarische Sector",
		"Media en Communicatie",
		"Onderwijs, Cultuur en Wetenschap",
		"Techniek, Productie en Bouw",
		"Toerisme, Recreatie en Horeca",
		"Transport en Logistiek",
		"Behoefte aan Investering",
		"Interesse om te Investeren",
	}

	for _, key := range keys {
		interest := models.Interest{Key: key}
		if err := DB.FirstOrCreate(&interest, models.Interest{Key: key}).Error; err != nil {
			log.Fatalf("failed to seed interest %s: %v", key, err)
		}
	}
}
