package util

import "fmt"

func GenerateYandexLink(lat, lon float64) string {
	return fmt.Sprintf("https://yandex.ru/maps?rtext=%.6f,%.6f~%.6f,%.6f", lat, lon, lat, lon)
}
