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
			builder.WriteString(fmt.Sprintf("📏 %0.f м | ", distToMe))
		}
		builder.WriteString(fmt.Sprintf("±%0.f м\n", loc.Accuracy))

		if loc.Address != nil {
			builder.WriteString(fmt.Sprintf("📍 %s\n", *loc.Address))
		} else {
			builder.WriteString("📍 Адрес не определен\n")
		}
	}

	if lastUpdate.Data.Battery != nil {
		if lastUpdate.Data.Battery.Charging {
			builder.WriteString("⚡")
		} else if lastUpdate.Data.Battery.Level > 30 {
			builder.WriteString("🔋")
		} else {
			builder.WriteString("🪫")
		}

		builder.WriteString(fmt.Sprintf(" %d%%\n", lastUpdate.Data.Battery.Level))
	}

	return builder.String()
}
