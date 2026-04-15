package utils

import (
	"time"
)

const timezoneBogota = "America/Bogota"

var locationBogota *time.Location

func init() {
	var err error
	locationBogota, err = time.LoadLocation(timezoneBogota)
	if err != nil {
		// Fallback a UTC-5 si no se puede cargar la location
		locationBogota = time.FixedZone("Bogota", -5*60*60)
	}
}

// AhoraBogota retorna la hora actual en timezone de Bogotá
func AhoraBogota() time.Time {
	return time.Now().In(locationBogota)
}

// FormatearBogota formatea una hora en formato legible para Bogotá
func FormatearBogota(t time.Time) string {
	return t.In(locationBogota).Format("2006-01-02 15:04:05")
}

// ConvertirABogota convierte cualquier time.Time a timezone Bogotá
func ConvertirABogota(t time.Time) time.Time {
	return t.In(locationBogota)
}
