package weatherapi

import "fmt"

// defaultForecastDays is the forecast length used when a request does not ask
// for a specific number of days.  It must be applied before the cache key is
// built so that an unset "days" and an explicit "days": 3 share one entry.
const defaultForecastDays = 3

// coordKeyFormat rounds coordinates to two decimals when building a cache key.
// 0.01 degrees is roughly 1.1 km (see the range note in rediscache), so nearby
// requests still share an entry while distinct locations never do.
const coordKeyFormat = "%s_%.2f,%.2f"

// normalizeDays applies the default forecast length.  Callers must use it both
// when building the cache key and when calling the upstream API, otherwise the
// same response is cached under two different keys.
func normalizeDays(days int) int {
	if days <= 0 {
		return defaultForecastDays
	}
	return days
}

// coordCacheKey builds a cache key from the requested coordinates.
//
// Coordinate requests were previously keyed by the city name returned by the
// reverse-geocoding service (`GetCity(...).City + "_current"`).  That serves
// the wrong location's weather in two ways:
//
//  1. When the lookup yields no city — an ocean or remote coordinate, a
//     transport error, a rate-limited response, or any body that does not
//     unmarshal — City is "" and every such request collapses onto a single
//     shared key ("_current" / "_forecast").  Whichever request populates it
//     first pins that payload for the whole TTL, and unrelated coordinates all
//     receive it.
//  2. Even on success the mapping is lossy: the service reports "Miami" for
//     Coral Springs (26.27119, -80.2706), ~55 km away, so two distinct places
//     share one entry while the upstream data is fetched for whichever exact
//     coordinates happened to populate it.
//
// Keying by the request coordinates avoids both, and removes the
// reverse-geocode call from the request path entirely.
func coordCacheKey(prefix string, lat, lon float32) string {
	return fmt.Sprintf(coordKeyFormat, prefix, lat, lon)
}

// forecastCoordCacheKey keys a coordinate forecast by point AND length.  The
// forecast key previously omitted the day count, so a 3-day response was
// served to a later 5-day request (and vice versa).
func forecastCoordCacheKey(lat, lon float32, days int) string {
	return fmt.Sprintf("%s_%dd", coordCacheKey("forecast", lat, lon), normalizeDays(days))
}

// queryCacheKey keys a city/zip request.  The query is included verbatim; it is
// escaped only when interpolated into the upstream URL.
func queryCacheKey(prefix, query string) string {
	return fmt.Sprintf("%s_%s", prefix, query)
}

// forecastQueryCacheKey keys a city/zip forecast by query AND length, for the
// same reason as forecastCoordCacheKey.
func forecastQueryCacheKey(query string, days int) string {
	return fmt.Sprintf("%s_%dd", queryCacheKey("forecast", query), normalizeDays(days))
}
