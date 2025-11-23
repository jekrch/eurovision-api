package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"eurovision-api/models"

	"github.com/sirupsen/logrus"
)

/**
 * identifies the location of the given IP address using ipapi.co and returns
 * IPLocation object.
 */
func fetchIPLocationFromIP(ip string) (ipLocation models.IPLocation, err error) {
	locationURL := fmt.Sprintf("https://ipapi.co/%s/json/", ip)
	resp, err := http.Get(locationURL)
	if err != nil {
		logrus.Printf("Error getting location: %v", err)
	} else {
		var ipLocation models.IPLocation
		if err := json.NewDecoder(resp.Body).Decode(&ipLocation); err != nil {
			log.Printf("Error decoding location: %v", err)
		}
		resp.Body.Close()
	}
	return ipLocation, err
}
