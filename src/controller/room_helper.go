package controller

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/ipzonedetector"
)

// note this method will only log the error, but not return the error
func (controller *Controller) getUserCountryCode(c *gin.Context) string {
	conf := config.GetInstance()
	if !conf.IsCloudMode() {
		return ""
	}
	clientIP := c.ClientIP()

	// check if it is in cache
	ipZone, errInGetIPZone := controller.Cache.IPZoneCache.GetIPZone(clientIP)
	if errInGetIPZone != nil {
		log.Printf("can not get user ip info from cache: " + errInGetIPZone.Error())
		return ""
	}
	if ipZone != "" {
		return ipZone
	}

	// ok, we query it from API
	ipZoneDetector := ipzonedetector.NewIPZoneDetector()
	countryCode, errInGetCountryCode := ipZoneDetector.GetCountryCode(clientIP)
	if errInGetCountryCode != nil {
		log.Printf("can not get user ip info from ipzonedetector: " + errInGetCountryCode.Error())
		return ""
	}

	// set cache
	controller.Cache.IPZoneCache.SetIPZone(clientIP, countryCode)

	return countryCode
}
