package util

import "fmt"

// https://yandex.ru/dev/yandex-apps-launch-maps/doc/ru/concepts/yandexmaps-ios-app

func GenerateYandexLinkForPoint(lat, lon float64) string {
	return fmt.Sprintf("https://maps.yandex.ru?pt=%.6f,%.6f&z=17", lon, lat)
}

func GenerateYandexLinkForRoute(lat1, lon1, lat2, lon2 float64, rtt string) string {
	return fmt.Sprintf("https://maps.yandex.ru?rtext=%.6f,%.6f~%.6f,%.6f&rtt=%s", lon1, lat1, lon2, lat2, rtt)
}
