# goweather
Api for weather. Using weatherapi and redis as cache

tasklist:
- [ ] handle cache in a range such that any long/lat pair fits to a city. hashmap
- [x] get current weather based on long/lat
- [x] get forecast based on long/lat
- [x] cache long/lat (not range)

Endpoints:
http://localhost:8080/current
```json
{
	"longitude": 100.501762,
	"latitude": 13.756331
}
```

http://localhost:8080/forecast
```json
{
	"longitude": 100.501762,
	"latitude": 13.756331,
}
```
