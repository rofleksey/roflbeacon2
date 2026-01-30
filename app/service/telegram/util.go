package telegram

import (
	"fmt"
	"roflbeacon2/app/api"
	"roflbeacon2/pkg/database"
	"roflbeacon2/pkg/util"
	"strings"
)

func (s *Service) resetState() {
	s.state = BotState{
		Stage: "idle",
	}
}

func (s *Service) formatUpdate(acc *database.Account, lastUpdate database.Update, myLastLocation *api.LocationData) string {
	var builder strings.Builder

	loc := lastUpdate.Data.Location

	builder.WriteString("*")
	builder.WriteString(acc.Name)
	builder.WriteString("* (")
	builder.WriteString(util.TimeAgo(lastUpdate.Created))
	builder.WriteString(")\n")

	if loc == nil {
		builder.WriteString("⚠️ Местоположение не определено")
	} else {
		mapLink := util.GenerateYandexLinkForPoint(loc.Latitude, loc.Longitude)

		builder.WriteString(fmt.Sprintf("[На карте](%s)", mapLink))
		if myLastLocation != nil {
			routeLink := util.GenerateYandexLinkForRoute(myLastLocation.Latitude, myLastLocation.Longitude, loc.Latitude, loc.Longitude, "mt")
			builder.WriteString(fmt.Sprintf(" | [Маршрут до меня](%s)", routeLink))
		}

		builder.WriteString("\n")

		if myLastLocation != nil {
			distToMe := util.HaversineDistance(myLastLocation.Latitude, myLastLocation.Longitude, loc.Latitude, loc.Longitude)

			if distToMe >= 1000 {
				builder.WriteString(fmt.Sprintf("📐 %.1f км | ", distToMe/1000))
			} else {
				builder.WriteString(fmt.Sprintf("📐 %.0f м | ", distToMe))
			}
		}
		builder.WriteString(fmt.Sprintf("±%.0f м\n", loc.Accuracy))

		if loc.Address != nil {
			builder.WriteString(fmt.Sprintf("📍 %s\n", *loc.Address))
		} else {
			builder.WriteString("📍 Адрес не определен\n")
		}
	}

	return builder.String()
}
