package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nathan-osman/go-sunrise"
)

const defaultHttpListenPort int = 2911

var LAT float64 = 53.19
var LON float64 = 19.76

func main() {
	log.Println("## sunmonitor ☀️ starting ##")
	port, _ := strconv.Atoi(os.Getenv("HTTP_LISTEN_PORT"))
	if port == 0 {
		port = defaultHttpListenPort
	}

	LAT, _ = strconv.ParseFloat(os.Getenv("LOCATION_LATITUDE"), 64)
	LON, _ = strconv.ParseFloat(os.Getenv("LOCATION_LONGITUDE"), 64)

	log.Println("config loaded: LAT=", LAT, " LON=", LON)
	log.Println("listening on port", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/suncheck", handleSunIsDownCheck)
	mux.HandleFunc("/suncheck/after/{afterminutes}", handleSunIsDownCheck)
	mux.HandleFunc("/suncheck/after/{afterminutes}/before/{beforeminutes}", handleSunIsDownCheck)

	log.Println("server will use following path patterns:")
	log.Println("  - /suncheck: no delay/accelerate")
	log.Println("  - /suncheck/after/{afterminutes}: delay after sunset")
	log.Println("  - /suncheck/after/{afterminutes}/before/{beforeminutes}: delay after sunset, accelerate before sunrise")
	log.Println("## sunmonitor ☀️ ready ##")

	s := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		IdleTimeout:  30 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}

func handleSunIsDownCheck(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	delayAfterSunset, _ := strconv.Atoi(req.PathValue("afterminutes"))
	accelerateBeforeSunrise, _ := strconv.Atoi(req.PathValue("beforeminutes"))

	if checkIfSunIsDown(delayAfterSunset, accelerateBeforeSunrise) {
		io.WriteString(w, "DARK")
	} else {
		io.WriteString(w, "SUN IS UP")
	}

}

func checkIfSunIsDown(delayAfterSunset, accelerateBeforeSunrise int) bool {
	return checkIfSunIsDownAt(time.Now(), delayAfterSunset, accelerateBeforeSunrise)
}

func checkIfSunIsDownAt(now time.Time, delayAfterSunset, accelerateBeforeSunrise int) bool {
	year := now.Year()
	month := now.Month()
	day := now.Day()

	rise, set := sunrise.SunriseSunset(
		LAT, LON,
		year, month, day,
	)

	if now.Add(time.Duration(accelerateBeforeSunrise)*time.Minute).Before(rise) || now.Add(-time.Minute*time.Duration(delayAfterSunset)).After(set) {
		// It is DARK OUTSIDE
		return true
	}

	return false
}
