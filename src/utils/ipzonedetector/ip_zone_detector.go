package ipzonedetector

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/utils/config"
)

const (
	API_TEMPLATE = "https://timezoneapi.io/api/ip/?ip=%s&token=%s"
)

type IPZoneDetector struct {
	Token       string
	IsCloudMode bool
	Debug       bool `json:"-"`
}

func NewIPZoneDetector() *IPZoneDetector {
	conf := config.GetInstance()
	token := conf.GetIPZoneDetectorToken()
	isCloudMode := conf.IsCloudMode()
	return &IPZoneDetector{
		Token:       token,
		IsCloudMode: isCloudMode,
		Debug:       false,
	}
}

func (r *IPZoneDetector) CloseDebug() {
	r.Debug = false
}

func (r *IPZoneDetector) OpenDebug() {
	r.Debug = true
}

/*
The response data like:
```json
{
    "meta": {
        "code": "200",
        "execution_time": "0.002857 seconds"
    },
    "data": {
        "ip": "139.59.4.229",
        "city": "Bengaluru",
        "postal": "560100",
        "state": "Karnataka",
        "state_code": "KA",
        "country": "India",
        "country_code": "IN",
        "location": "12.9833,77.5833",
        "timezone": {
            "id": "Asia/Kolkata",
            "location": "22.53333,88.36666",
            "country_code": "IN",
            "country_name": "India",
            "iso3166_1_alpha_2": "IN",
            "iso3166_1_alpha_3": "IND",
            "un_m49_code": "356",
            "itu": "IND",
            "marc": "ii",
            "wmo": "IN",
            "ds": "IND",
            "phone_prefix": "91",
            "fifa": "IND",
            "fips": "IN",
            "gual": "115",
            "ioc": "IND",
            "currency_alpha_code": "INR",
            "currency_country_name": "INDIA",
            "currency_minor_unit": "2",
            "currency_name": "Indian Rupee",
            "currency_code": "356",
            "independent": "Yes",
            "capital": "New Delhi",
            "continent": "AS",
            "tld": ".in",
            "languages": "en-IN,hi,bn,te,mr,ta,ur,gu,kn,ml,or,pa,as,bh,sat,ks,ne,sd,kok,doi,mni,sit,sa,fr,lus,inc",
            "geoname_id": "1269750",
            "edgar": "K7"
        },
        "datetime": {
            "date": "10/30/2023",
            "date_time": "10/30/2023 16:46:40",
            "date_time_txt": "Monday, October 30, 2023 16:46:40",
            "date_time_wti": "Mon, 30 Oct 2023 16:46:40 +0530",
            "date_time_ymd": "2023-10-30T16:46:40+05:30",
            "time": "16:46:40",
            "month": "10",
            "month_wilz": "10",
            "month_abbr": "Oct",
            "month_full": "October",
            "month_days": "31",
            "day": "30",
            "day_wilz": "30",
            "day_abbr": "Mon",
            "day_full": "Monday",
            "year": "2023",
            "year_abbr": "23",
            "hour_12_wolz": "4",
            "hour_12_wilz": "04",
            "hour_24_wolz": "16",
            "hour_24_wilz": "16",
            "hour_am_pm": "pm",
            "minutes": "46",
            "seconds": "40",
            "week": "44",
            "offset_seconds": "19800",
            "offset_minutes": "330",
            "offset_hours": "5.5",
            "offset_gmt": "+05:30",
            "offset_tzid": "Asia/Kolkata",
            "offset_tzab": "IST",
            "offset_tzfull": "India Standard Time",
            "tz_string": "IST6",
            "dst": "false",
            "dst_observes": "true",
            "timeday_spe": "late_afternoon",
            "timeday_gen": "afternoon"
        }
    }
}
```
*/

func (r *IPZoneDetector) GetCountryCode(ipAddress string) (string, error) {
	// self-host need skip this method.
	if !r.IsCloudMode {
		return "", nil
	}
	client := resty.New()
	uri := fmt.Sprintf(API_TEMPLATE, ipAddress, r.Token)
	resp, errInGet := client.R().Get(uri)
	if r.Debug {
		log.Printf("[IPZoneDetector.GetCountryCode()]  uri: %+v \n", uri)
		log.Printf("[IPZoneDetector.GetCountryCode()]  response: %+v, err: %+v \n", resp, errInGet)
		log.Printf("[IPZoneDetector.GetCountryCode()]  resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInGet != nil {
		return "", errInGet
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return "", errors.New(resp.String())
	}

	var ipData map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &ipData)
	if errInUnMarshal != nil {
		return "", errInUnMarshal
	}
	// parse data field
	ipDataFieldRaw, hitDataField := ipData["data"]
	if !hitDataField {
		return "", errors.New("can not get ip zone info by ip address: " + ipAddress)
	}
	ipDataField, assertIPDataFieldPass := ipDataFieldRaw.(map[string]interface{})
	if !assertIPDataFieldPass {
		return "", errors.New("can not assert ipdata.data field by ip address: " + ipAddress)
	}
	// get country code field
	countryCodeFieldRaw, hitCountryCodeFieldRaw := ipDataField["country_code"]
	if !hitCountryCodeFieldRaw {
		return "", errors.New("can not get ipdata.data.country_code field by ip address: " + ipAddress)
	}
	countryCodeField, assertCountryCodeFieldPass := countryCodeFieldRaw.(string)
	if !assertCountryCodeFieldPass {
		return "", errors.New("can not assert ipdata.data.country_code field by ip address: " + ipAddress)
	}
	return countryCodeField, nil
}
