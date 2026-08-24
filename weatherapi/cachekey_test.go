package weatherapi

import "testing"

// The bug this guards against: coordinate requests were keyed by the city name
// from the reverse-geocode lookup, so every request whose lookup produced no
// city shared one key and received the first-cached payload.
func TestCoordCacheKeyIsDistinctPerLocation(t *testing.T) {
	points := map[string]struct{ lat, lon float32 }{
		"coral springs": {26.27119, -80.27060},
		"miami":         {25.77427, -80.19366},
		"tokyo":         {35.68950, 139.69170},
		"london":        {51.50740, -0.12780},
		"bangkok":       {13.75398, 100.50144},
		"pacific ocean": {-30.00000, -140.00000},
		"null island":   {0, 0},
	}

	seen := make(map[string]string, len(points))
	for name, p := range points {
		key := coordCacheKey("current", p.lat, p.lon)
		if other, dup := seen[key]; dup {
			t.Errorf("%q and %q share cache key %q — one would be served the other's weather",
				name, other, key)
		}
		seen[key] = name
	}
}

// A request with no coordinates (an empty body) must not share a key with a
// real location.  Previously both produced "_current".
func TestZeroCoordsDoNotCollideWithRealLocation(t *testing.T) {
	empty := coordCacheKey("current", 0, 0)
	tokyo := coordCacheKey("current", 35.6895, 139.6917)
	if empty == tokyo {
		t.Fatalf("empty request and Tokyo share key %q", empty)
	}
}

// Nearby points should still share an entry — the cache must stay useful.
func TestNearbyCoordsShareCacheKey(t *testing.T) {
	a := coordCacheKey("current", 13.75398, 100.50144)
	b := coordCacheKey("current", 13.75401, 100.50139)
	if a != b {
		t.Errorf("points ~5 m apart got different keys %q and %q", a, b)
	}
}

func TestCurrentAndForecastDoNotShareKey(t *testing.T) {
	cur := coordCacheKey("current", 13.75398, 100.50144)
	fc := forecastCoordCacheKey(13.75398, 100.50144, 3)
	if cur == fc {
		t.Fatalf("current and forecast share key %q", cur)
	}
}

// The forecast key omitted the day count, so a cached 3-day response was
// served to a 5-day request.
func TestForecastKeyVariesByDays(t *testing.T) {
	three := forecastCoordCacheKey(13.75398, 100.50144, 3)
	five := forecastCoordCacheKey(13.75398, 100.50144, 5)
	if three == five {
		t.Fatalf("3-day and 5-day forecasts share key %q", three)
	}
}

// An unset day count must normalize to the default before keying, or the same
// response is cached twice.
func TestForecastKeyNormalizesDefaultDays(t *testing.T) {
	unset := forecastCoordCacheKey(13.75398, 100.50144, 0)
	explicit := forecastCoordCacheKey(13.75398, 100.50144, defaultForecastDays)
	if unset != explicit {
		t.Fatalf("unset days %q != explicit default %q", unset, explicit)
	}
}

func TestForecastQueryKeyVariesByDays(t *testing.T) {
	if forecastQueryCacheKey("Bangkok", 3) == forecastQueryCacheKey("Bangkok", 5) {
		t.Fatal("3-day and 5-day query forecasts share a key")
	}
	if forecastQueryCacheKey("Bangkok", 0) != forecastQueryCacheKey("Bangkok", defaultForecastDays) {
		t.Fatal("unset days did not normalize for query forecasts")
	}
}

func TestQueryCacheKeysAreDistinct(t *testing.T) {
	if queryCacheKey("current", "Bangkok") == queryCacheKey("current", "Tokyo") {
		t.Fatal("distinct city queries share a cache key")
	}
	if queryCacheKey("current", "Bangkok") == queryCacheKey("forecast", "Bangkok") {
		t.Fatal("current and forecast query keys collide")
	}
}

func TestNormalizeDays(t *testing.T) {
	for _, tc := range []struct{ in, want int }{
		{-1, defaultForecastDays}, {0, defaultForecastDays}, {1, 1}, {3, 3}, {5, 5},
	} {
		if got := normalizeDays(tc.in); got != tc.want {
			t.Errorf("normalizeDays(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
