package controller

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"../model"
	"../shared/session"
	"../shared/view"
	"github.com/josephspurrier/csrfbanana"
	"googlemaps.github.io/maps"
)

var (
	mapapiKey       = "<enter Googleapi key here>"
	owmapiKey       = "<enter OpenWeatherMap api key here>"
	mapaddress      = ""
	mapcomponents   = ""
	mapbounds       = ""
	maplanguage     = ""
	mapregion       = ""
	maplatlng       = ""
	mapresultType   = ""
	maplocationType = ""
)

var jsondata map[string]interface{}

func check(err error) {
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
}

// SearchGET displays the notes in the notepad
func SearchGET(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	//userID := fmt.Sprintf("%s", sess.Values["id"])

	//notes, err := model.NotesByUserID(userID)
	//if err != nil {
	//	log.Println(err)
	//	notes = []model.Note{}
	//}

	// Display the view
	v := view.New(r)
	v.Name = "search/search"
	v.Vars["_city"] = sess.Values["_city"]
	v.Vars["_state"] = sess.Values["_state"]
	v.Vars["_zipcode"] = sess.Values["_zipcode"]
	v.Vars["_address"] = sess.Values["_address"]
	v.Vars["_currtemp"] = sess.Values["_currtemp"]
	v.Vars["_hightemp"] = sess.Values["_hightemp"]
	v.Vars["_lowtemp"] = sess.Values["_lowtemp"]
	v.Vars["_desc"] = sess.Values["_desc"]
	v.Vars["_icon"] = sess.Values["_icon"]
	v.Vars["_ifsrc"] = sess.Values["_ifsrc"]
	v.Vars["token"] = csrfbanana.Token(w, r, sess)
	v.Render(w)
}

// SearchPOST handles the note creation form submission
func SearchPOST(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// Validate with required fields
	if validate, missingField := view.Validate(r, []string{"_city"}); !validate {
		log.Println("Missing the ", missingField)
		sess.AddFlash(view.Flash{"Please enter a valid City!!", view.FlashError})
		sess.Save(r, w)
		SearchGET(w, r)
		return
	}
	if validate, missingField := view.Validate(r, []string{"_state"}); !validate {
		log.Println("Missing the ", missingField)
		sess.AddFlash(view.Flash{"Please enter a valid State!!", view.FlashError})
		sess.Save(r, w)
		SearchGET(w, r)
		return
	}
	if validate, missingField := view.Validate(r, []string{"_zipcode"}); !validate {
		log.Println("Missing the ", missingField)
		sess.AddFlash(view.Flash{"Please enter a valid Zipcode!!", view.FlashError})
		sess.Save(r, w)
		SearchGET(w, r)
		return
	}

	// Get form values
	//streetaddress := r.FormValue("_address")
	city := r.FormValue("_city")
	usstate := r.FormValue("_state")
	zipcode := r.FormValue("_zipcode")
	address := r.FormValue("_address")
	mapaddress = address

	//log.Println("City is ", city)
	//log.Println("State is ", usstate)
	//log.Println("Zipcode is ", zipcode)

	/*
			var myClient = &http.Client{Timeout: 10 * time.Second}
			respgeo, err := http.Get("https://maps.googleapis.com/maps/api/geocode/json?address=" + city + ", " + usstate)
			if err != nil {
				log.Println("Z3")
				// handle error
			}
			defer respgeo.Body.Close()
			json.NewDecoder(respgeo.Body).Decode(target)

		    //https: //maps.googleapis.com/maps/api/geocode/json?address=amsterdam

		    //https: //maps.googleapis.com/maps/api/geocode/json?latlng=52.3182742,4.7288558
	*/

	now := time.Now()
	log.Println("Time is ", now)

	startSC := model.SearchCache{zipcode, city, usstate, address, 1, "", "", "", "", "", "", now, 0}

	if zipcode != "" {
		// Get the searchcache
		searchcache, err := model.SearchCacheByZipcode(zipcode)
		//if err != nil { // If the zipcode does not exist
		//log.Println(err)
		//sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		//sess.Save(r, w)
		//http.Redirect(w, r, "/search", http.StatusFound)
		//return
		//}
		log.Println("After Search ", err, "and ", searchcache)

		if err == nil {
			log.Println("Found a Cache Zipcode ", zipcode, " and time was ", searchcache.UpdatedAt)
			dur := time.Since(searchcache.UpdatedAt)
			log.Println("Duration is ", round(dur.Minutes()))

			if round(dur.Minutes()) <= 3 {
				startSC.CurrTemp = searchcache.CurrTemp
				startSC.HighTemp = searchcache.HighTemp
				startSC.LowTemp = searchcache.LowTemp
				startSC.Description = searchcache.Description
				startSC.Icon = searchcache.Icon
				startSC.IFrameSrc = searchcache.IFrameSrc
				sess.AddFlash(view.Flash{"Got From Cache: " + strconv.Itoa(round(dur.Minutes())) + " minutes old", view.FlashError})
				sess.Save(r, w)
				//SearchGET(w, r)

				v := view.New(r)
				v.Name = "search/search"
				v.Vars["_city"] = startSC.City
				v.Vars["_state"] = startSC.State
				v.Vars["_zipcode"] = startSC.Zipcode
				v.Vars["_address"] = startSC.Address
				v.Vars["_currtemp"] = startSC.CurrTemp
				v.Vars["_hightemp"] = startSC.HighTemp
				v.Vars["_lowtemp"] = startSC.LowTemp
				v.Vars["_desc"] = startSC.Description
				v.Vars["_icon"] = startSC.Icon
				v.Vars["_ifsrc"] = startSC.IFrameSrc
				//v.Vars["_ifsrc"] = "https://weather.com/weather/today/l/" + searchcache.Zipcode + ":4:US"
				v.Vars[GetStateOptionParm(startSC.State)] = "selected"
				v.Vars["token"] = csrfbanana.Token(w, r, sess)
				v.Render(w)

				return
			} else {
				err := model.SearchCacheDelete(zipcode)
				if err != nil {
					log.Println(err)
					sess.AddFlash(view.Flash{"An error occurred on the server during delete. Please try again later.", view.FlashError})
					sess.Save(r, w)
				}
			}
		}
	}

	log.Println("Create new cache ", zipcode)

	newSC := model.SearchCache{"", "", "", "", 1, "", "", "", "", "", "", now, 0}
	currtempStr := ""
	hightempStr := ""
	lowtempStr := ""
	weatherdetailsStr := ""
	weatherphraseStr := ""
	feelslikeStr1 := ""
	feelslikeStr2 := ""
	weathericon := ""
	weatherdesc := ""
	ifsrc := ""

	//mode := "api"
	mode := r.FormValue("_rdomode")
	log.Println("Mode is", mode)

	if mode == "weather" {
		//Get via screen scrape

		//Go Get Value

		//https://weather.com/weather/today/l/30005:4:US
		//<div class="today_nowcard-temp"><span>70<sup>°</sup></span></div>
		//<div class="today_nowcard-phrase">Mostly Cloudy</div>
		//<div class="today_nowcard-feels"><span class="btn-text">Feels Like </span><span class="deg-feels" classname="deg-feels">70<sup>°</sup></span></div>
		//<span class="btn-text">H </span>
		//<span class="deg-hilo-nowcard"><span>86<sup>°</sup></span></span>
		//<span class="deg-hilo-nowcard"><span>68<sup>°</sup></span></span>
		//<span class="btn-text">UV Index </span>
		//<span>0 of 10</span>

		//log.Println("A1")
		//resp, err := http.Get("https://weather.com/weather/today/l/" + searchcache.Zipcode + ":4:US")

		tr := &http.Transport{MaxIdleConns: 10, IdleConnTimeout: 30 * time.Second, DisableCompression: true}
		client := &http.Client{Transport: tr}
		resp, err := client.Get("https://weather.com/weather/today/l/" + startSC.Zipcode + ":4:US")

		if err != nil {
			log.Println("A3")
			// handle error
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		//sbody := string(body[:resp.ContentLength])
		sbody := string(body)
		rbody := []rune(string(body))
		log.Println("The string response is:", len(sbody))
		log.Println("The rune response is:", len(rbody))
		fi, err := os.Create("httpresult.txt")
		err = ioutil.WriteFile(fi.Name(), []byte(sbody), 0644)

		//   This is for Weather.com

		//Get Current Tempurature
		currtempStr = GetTagSubString(0, "div", sbody, rbody, "\"today_nowcard-temp\"")
		log.Println("Curr Temp String: ", currtempStr)
		currtempStr = GetTagSubString(0, "span", currtempStr, []rune(currtempStr), "")
		log.Println("Curr Temp String: ", currtempStr)
		currtempStr = KeepNumbers(currtempStr)
		log.Println("Curr Temp String: ", currtempStr)

		//Get Weather Phrase
		weatherphraseStr = GetTagSubString(0, "div", sbody, rbody, "\"today_nowcard-phrase\"")
		log.Println("Phrase String: ", weatherphraseStr)
		weatherdesc = weatherphraseStr
		weathericon = ""

		//Get Feels Like
		feelslikeStr1 = GetTagSubString(0, "div", sbody, rbody, "\"today_nowcard-feels\"")
		log.Println("Feels Like: ", feelslikeStr1)
		feelslikeStr1 = GetTagSubString(0, "span", feelslikeStr1, []rune(feelslikeStr1), "")
		log.Println("Feels Like: ", feelslikeStr1)
		feelslikeStr2 = GetTagSubString(0, "div", sbody, rbody, "\"today_nowcard-feels\"")
		log.Println("Feels Like: ", feelslikeStr2)
		feelslikeStr2 = GetTagSubString(0, "span", feelslikeStr2, []rune(feelslikeStr2), "\"deg-feels\"")
		log.Println("Feels Like: ", feelslikeStr2)
		feelslikeStr2 = KeepNumbers(feelslikeStr2)
		log.Println("Feels Like: ", feelslikeStr2)

		//Get High Temp
		hightempStr = GetTagSubString(0, "span", sbody, rbody, "\"deg-hilo-nowcard\"")
		log.Println("High Temp String: ", hightempStr)
		hightempStr = KeepNumbers(hightempStr)
		log.Println("High Temp String: ", hightempStr)

		//Get Low Temp
		i := strings.Index(sbody, "\"deg-hilo-nowcard\"") + 1
		lowtempStr = GetTagSubString(i, "span", sbody, rbody, "\"deg-hilo-nowcard\"")
		log.Println("Low Temp String: ", lowtempStr)
		lowtempStr = KeepNumbers(lowtempStr)
		log.Println("Low Temp String: ", lowtempStr)

		ifsrc = "https://weather.com/weather/today/l/" + startSC.Zipcode + ":4:US"

		newSC = model.SearchCache{startSC.Zipcode, startSC.City, startSC.State, startSC.Address, 1, currtempStr, hightempStr, lowtempStr, weatherdesc, weathericon, ifsrc, now, 0}
	}

	if mode == "foreca" {
		/*   This is for Foreca.com  */
		tr := &http.Transport{MaxIdleConns: 10, IdleConnTimeout: 30 * time.Second, DisableCompression: true}
		client := &http.Client{Transport: tr}
		resp, err := client.Get("https://www.foreca.com/United_States/" + startSC.State + "/" + startSC.City)

		if err != nil {
			log.Println("A3")
			// handle error
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		//sbody := string(body[:resp.ContentLength])
		sbody := string(body)
		rbody := []rune(string(body))
		log.Println("The string response is:", len(sbody))
		log.Println("The rune response is:", len(rbody))
		fi, err := os.Create("httpresult.txt")
		err = ioutil.WriteFile(fi.Name(), []byte(sbody), 0644)

		//Get Current Tempurature
		currtempStr = GetTagSubString(0, "span", sbody, rbody, "txt-xxlarge")
		log.Println("Curr Temp StringA: ", currtempStr)
		currtempStr = GetTagSubString(0, "strong", currtempStr, []rune(currtempStr), "")
		log.Println("Curr Temp StringB: ", currtempStr)

		//Get Weather Phrase
		weatherdetailsStr = GetTagSubString(0, "div", sbody, rbody, "right txt-tight")
		//log.Println("Details String: ", weatherdetailsStr)

		weatherphraseStr = strings.TrimSpace(weatherdetailsStr[0:strings.Index(weatherdetailsStr, "<br")])
		//r := utf8.RuneCountInString(weatherdetailsStr[0 : i+ii+iii])
		//returnStr = returnStr + string(bigrune[rr+1:rrr])
		//r := utf8.RuneCountInString(weatherdetailsStr[0 : strings.Index(weatherdetailsStr, "<br")])
		//PrintRuneClump(bigrune, rrr, 4)
		log.Println("Phrase String: ", weatherphraseStr)
		weatherdesc = weatherphraseStr
		weathericon = ""

		//Get Feels Like
		feelslikeStr1 = GetTagSubString(0, "strong", weatherdetailsStr, []rune(weatherdetailsStr), "Feels Like:")
		log.Println("Feels LikeA: ", feelslikeStr1)
		feelslikeStr1 = KeepNumbers(feelslikeStr1)
		log.Println("Feels LikeB: ", feelslikeStr1)

		//Get High Temp
		hightempStr = GetTagSubString(0, "div", sbody, rbody, "minmax")
		log.Println("High Temp StringA: ", hightempStr)
		hightempStr = GetTagSubString(0, "span", hightempStr, []rune(hightempStr), "")
		log.Println("High Temp StringB: ", hightempStr)
		hightempStr = KeepNumbers(hightempStr)
		log.Println("High Temp StringC: ", hightempStr)

		//Get Low Temp
		lowtempStr = GetTagSubString(0, "div", sbody, rbody, "minmax")
		log.Println("Low Temp StringA: ", lowtempStr)
		i := strings.Index(lowtempStr, "Lo:") + 1
		lowtempStr = GetTagSubString(i, "strong", lowtempStr, []rune(lowtempStr), "")
		log.Println("Low Temp StringB: ", lowtempStr)
		lowtempStr = KeepNumbers(lowtempStr)
		log.Println("Low Temp StringC: ", lowtempStr)

		ifsrc = "https://www.foreca.com/United_States/" + startSC.State + "/" + startSC.City

		newSC = model.SearchCache{startSC.Zipcode, startSC.City, startSC.State, startSC.Address, 1, currtempStr, hightempStr, lowtempStr, weatherdesc, weathericon, ifsrc, now, 0}
		//newSC.Zipcode = zipcode
		//newSC.CurrTemp = "80"
		//newSC.HighTemp = "90"
		//newSC.LowTemp = "70"
	}

	if mode == "api" {
		//Get from API

		var client *maps.Client
		var err error
		var useGoogle bool
		var latitude string
		var longitude string

		useGoogle = false
		if useGoogle {
			log.Println("Using Key: ", mapapiKey)
			client, err = maps.NewClient(maps.WithAPIKey(mapapiKey))
			check(err)

			r := &maps.GeocodingRequest{
				Address:  mapaddress,
				Language: maplanguage,
				Region:   mapregion,
			}

			//log.Println("::: r :::", r)

			parseComponents(mapcomponents, r)
			parseBounds(mapbounds, r)
			parseLatLng(maplatlng, r)
			parseResultType(mapresultType, r)
			parseLocationType(maplocationType, r)
			//latitude := parseLat(maplatlng, r)
			//longitude := parseLon(maplatlng, r)

			resp, err := client.Geocode(context.Background(), r)
			//check(err)

			if err != nil {
				log.Println(err)
				latitude = "38.806745"
				longitude = "-121.329121"
			} else {
				log.Println("::: resp :::", resp)
				latitude = strconv.FormatFloat(resp[0].Geometry.Location.Lat, 'f', 6, 64)
				longitude = strconv.FormatFloat(resp[0].Geometry.Location.Lng, 'f', 6, 64)
			}
			log.Println("Lat/Long", latitude, "::", longitude)
		}

		/*   This is for Foreca.com  */
		err = nil
		tr2 := &http.Transport{MaxIdleConns: 10, IdleConnTimeout: 30 * time.Second, DisableCompression: true}
		if err != nil {
			log.Println("B1")
		}
		client2 := &http.Client{Transport: tr2}
		if err != nil {
			log.Println("B2")
		}

		var resp2 *http.Response
		if useGoogle {
			resp2, err = client2.Get("https://api.openweathermap.org/data/2.5/weather?lat=" + latitude + "&lon=" + longitude + "&units=imperial&appid=" + owmapiKey)
			//resp2, err := client2.Get("https://api.openweathermap.org/data/2.5/forecast?lat=" + latitude + "&lon=" + longitude + "&units=imperial&APPID=" + owmapiKey)
			//resp2, err := client2.Get("https://api.openweathermap.org/data/2.5/forecast?lat=38.806745&lon=-121.329121&units=imperial&APPID=" + owmapiKey)
		} else {
			resp2, err = client2.Get("https://api.openweathermap.org/data/2.5/weather?zip=" + zipcode + ",US&units=imperial&appid=" + owmapiKey)
			//38.806745 :: -121.329121
		}

		if err != nil {
			log.Println("A3")
			sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
			//sess.Save(r, w)
			return
		} else {
			defer resp2.Body.Close()
			body2, err := ioutil.ReadAll(resp2.Body)
			if err != nil {
				log.Println("A4")
				sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
				//sess.Save(r, w)
				return
			}
			//sbody2 := string(body2[:resp2.ContentLength])
			sbody2 := string(body2)
			rbody2 := []rune(string(body2))
			log.Println("The string response is:", len(sbody2))
			log.Println("The rune response is:", len(rbody2))
			fi, err := os.Create("httpresult2.txt")
			err = ioutil.WriteFile(fi.Name(), []byte(sbody2), 0644)

			if err := json.Unmarshal(body2, &jsondata); err != nil {
				panic(err)
			}

			//Get Current Tempurature
			weathergroup := jsondata["weather"].([]interface{})
			weathergroup1 := weathergroup[0].(map[string]interface{})
			maingroup := jsondata["main"].(map[string]interface{})
			currtempStr = strconv.Itoa(round(maingroup["temp"].(float64)))
			weatherdetailsStr = weathergroup1["main"].(string)
			weatherphraseStr = weathergroup1["description"].(string)
			weatherdesc = weatherdetailsStr + ": " + weatherphraseStr
			weathericon = "http://openweathermap.org/img/w/" + weathergroup1["icon"].(string) + ".png"
			feelslikeStr1 = ""
			hightempStr = strconv.Itoa(round(maingroup["temp_max"].(float64)))
			lowtempStr = strconv.Itoa(round(maingroup["temp_min"].(float64)))
			ifsrc = "https://openweathermap.org/find?q=" + startSC.Zipcode + ",US&units=imperial"

			newSC = model.SearchCache{startSC.Zipcode, startSC.City, startSC.State, startSC.Address, 1, currtempStr, hightempStr, lowtempStr, weatherdesc, weathericon, ifsrc, now, 0}
		}
	}

	// Get database result
	err2 := model.SearchCacheCreate(newSC)
	log.Println("After Create ", err2)
	if err2 != nil {
		log.Println(err2)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
	} else {
		sess.AddFlash(view.Flash{"Zipcode added!", view.FlashSuccess})
		sess.Save(r, w)

		v := view.New(r)
		v.Name = "search/search"
		v.Vars["_city"] = newSC.City
		v.Vars["_state"] = newSC.State
		v.Vars["_zipcode"] = newSC.Zipcode
		v.Vars["_address"] = newSC.Address
		v.Vars["_currtemp"] = newSC.CurrTemp
		v.Vars["_hightemp"] = newSC.HighTemp
		v.Vars["_lowtemp"] = newSC.LowTemp
		v.Vars["_desc"] = newSC.Description
		v.Vars["_icon"] = newSC.Icon
		v.Vars["_ifsrc"] = newSC.IFrameSrc
		v.Vars[GetStateOptionParm(newSC.State)] = "selected"
		v.Vars["token"] = csrfbanana.Token(w, r, sess)
		v.Render(w)
		//http.Redirect(w, r, "/search", http.StatusFound)
		return
	}

	// Display the same page
	//SearchGET(w, r)
	return
}

func round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

// GetTagSubString will return the contents inside the HTML tag when screen scraping
func GetTagSubString(starti int, tag string, bigstr string, bigrune []rune, smallstr string) string {
	i := starti
	r := utf8.RuneCountInString(bigstr[0:i])
	//log.Println("Inside TagSearch1 i=", i, "r=", r)

	if smallstr != "" {
		log.Println("Inside TagSearch2 smallstr=", smallstr)
		i = strings.Index(bigstr[starti:len(bigstr)], smallstr)
		//i = strings.Index(bigstr, smallstr)
		r = utf8.RuneCountInString(bigstr[0:i])
		//log.Println("Inside TagSearch3 i=", i, "r=", r)
		PrintRuneClump(bigrune, r, 4)
		//log.Println("Pulled ", bigstr[n:j])
	}

	ii := strings.Index(bigstr[i:len(bigstr)], ">")
	log.Println("Inside TagSearch5 ii=", ii)
	rr := utf8.RuneCountInString(bigstr[0 : i+ii])
	log.Println("Inside TagSearch5 rr=", rr)
	PrintRuneClump(bigrune, rr, 4)

	iii := strings.Index(bigstr[i+ii:len(bigstr)], "</"+tag+">")
	log.Println("Inside TagSearch6 iii=", iii)
	rrr := utf8.RuneCountInString(bigstr[0 : i+ii+iii])
	log.Println("Inside TagSearch6 rrr=", rrr)
	PrintRuneClump(bigrune, rrr, 4)

	returnStr := ""
	PrintRuneClump(bigrune, rr+1, rrr-rr)
	returnStr = returnStr + string(bigrune[rr+1:rrr])
	log.Println("Returning=", returnStr)
	return returnStr
}

// KeepNumbers will return a string of just the numbers and remove the other characters.
func KeepNumbers(str string) string {
	log.Println("Entering KeepNumbers")
	returnStr := ""
	runestr := []rune(str)

	for i := 0; i < len(runestr); i++ {
		_, err := strconv.Atoi(string(runestr[i]))
		if err == nil {
			returnStr = returnStr + string(runestr[i])
		}
	}
	return returnStr
}

func parseComponents(components string, r *maps.GeocodingRequest) {
	if components == "" {
		return
	}
	if r.Components == nil {
		r.Components = make(map[maps.Component]string)
	}

	c := strings.Split(components, "|")
	for _, cf := range c {
		i := strings.Split(cf, ":")
		switch i[0] {
		case "route":
			r.Components[maps.ComponentRoute] = i[1]
		case "locality":
			r.Components[maps.ComponentLocality] = i[1]
		case "administrative_area":
			r.Components[maps.ComponentAdministrativeArea] = i[1]
		case "postal_code":
			r.Components[maps.ComponentPostalCode] = i[1]
		case "country":
			r.Components[maps.ComponentCountry] = i[1]
		default:
			log.Fatalf("parseComponents: component name %#v unknown", i[0])
		}
	}
}

func parseBounds(bounds string, r *maps.GeocodingRequest) {
	if bounds != "" {
		b := strings.Split(bounds, "|")
		sw := strings.Split(b[0], ",")
		ne := strings.Split(b[1], ",")

		swLat, err := strconv.ParseFloat(sw[0], 64)
		if err != nil {
			log.Fatalf("Couldn't parse bounds: %#v", err)
		}
		swLng, err := strconv.ParseFloat(sw[1], 64)
		if err != nil {
			log.Fatalf("Couldn't parse bounds: %#v", err)
		}
		neLat, err := strconv.ParseFloat(ne[0], 64)
		if err != nil {
			log.Fatalf("Couldn't parse bounds: %#v", err)
		}
		neLng, err := strconv.ParseFloat(ne[1], 64)
		if err != nil {
			log.Fatalf("Couldn't parse bounds: %#v", err)
		}

		r.Bounds = &maps.LatLngBounds{
			NorthEast: maps.LatLng{Lat: neLat, Lng: neLng},
			SouthWest: maps.LatLng{Lat: swLat, Lng: swLng},
		}
	}
}

func parseLatLng(latlng string, r *maps.GeocodingRequest) {
	if latlng != "" {
		l := strings.Split(latlng, ",")
		lat, err := strconv.ParseFloat(l[0], 64)
		if err != nil {
			log.Fatalf("Couldn't parse latlng: %#v", err)
		}
		lng, err := strconv.ParseFloat(l[1], 64)
		if err != nil {
			log.Fatalf("Couldn't parse latlng: %#v", err)
		}
		r.LatLng = &maps.LatLng{
			Lat: lat,
			Lng: lng,
		}
	}
}

func parseLat(latlng string, r *maps.GeocodingRequest) string {
	lat := ""
	if latlng != "" {
		l := strings.Split(latlng, ",")
		lat = l[0]
	}
	return lat
}

func parseLon(latlng string, r *maps.GeocodingRequest) string {
	lng := ""
	if latlng != "" {
		l := strings.Split(latlng, ",")
		lng = l[1]
	}
	return lng
}

func parseResultType(resultType string, r *maps.GeocodingRequest) {
	if resultType != "" {
		r.ResultType = strings.Split(resultType, "|")
	}
}

func parseLocationType(locationType string, r *maps.GeocodingRequest) {
	if locationType != "" {
		for _, l := range strings.Split(locationType, "|") {
			switch l {
			case "ROOFTOP":
				r.LocationType = append(r.LocationType, maps.GeocodeAccuracyRooftop)
			case "RANGE_INTERPOLATED":
				r.LocationType = append(r.LocationType, maps.GeocodeAccuracyRangeInterpolated)
			case "GEOMETRIC_CENTER":
				r.LocationType = append(r.LocationType, maps.GeocodeAccuracyGeometricCenter)
			case "APPROXIMATE":
				r.LocationType = append(r.LocationType, maps.GeocodeAccuracyApproximate)
			}
		}

	}
}

//PrintRuneClump will print a few runes
func PrintRuneClump(runes []rune, start int, clumpsize int) {
	log.Println("Rune \"", string(runes[start:start+clumpsize]), "\"")
}

// GetStateOptionParm will return the HTML Select/Options parameter to set for the dropdown
func GetStateOptionParm(s string) string {
	var p string
	p = ""

	switch s {
	case "Alabama":
		p = "AL"
	case "Alaska":
		p = "AK"
	case "Arizona":
		p = "AZ"
	case "Arkansas":
		p = "AR"
	case "California":
		p = "CA"
	case "Colorado":
		p = "CO"
	case "Connecticut":
		p = "CT"
	case "Delaware":
		p = "DE"
	case "Florida":
		p = "FL"
	case "Georgia":
		p = "GA"
	case "Hawaii":
		p = "HW"
	case "Idaho":
		p = "ID"
	case "Illinois":
		p = "IL"
	case "Indiana":
		p = "IN"
	case "Iowa":
		p = "IW"
	case "Kansas":
		p = "KS"
	case "Kentucky":
		p = "KY"
	case "Louisiana":
		p = "LO"
	case "Maine":
		p = "ME"
	case "Maryland":
		p = "MD"
	case "Massachusetts":
		p = "MA"
	case "Michigan":
		p = "MI"
	case "Minnesota":
		p = "MN"
	case "Mississippi":
		p = "MS"
	case "Missouri":
		p = "MO"
	case "Montana":
		p = "MT"
	case "Nebraska":
		p = "NE"
	case "Nevada":
		p = "NV"
	case "New Hampshire":
		p = "NH"
	case "New Jersey":
		p = "NJ"
	case "New Mexico":
		p = "NM"
	case "New York":
		p = "NY"
	case "North Carolina":
		p = "NC"
	case "North Dakota":
		p = "ND"
	case "Ohio":
		p = "OH"
	case "Oklahoma":
		p = "OK"
	case "Oregon":
		p = "OR"
	case "Pennsylvania":
		p = "PN"
	case "Rhode Island":
		p = "RI"
	case "South Carolina":
		p = "SC"
	case "South Dakota":
		p = "SD"
	case "Tennessee":
		p = "TN"
	case "Texas":
		p = "TX"
	case "Utah":
		p = "UT"
	case "Vermont":
		p = "VE"
	case "Virginia":
		p = "VA"
	case "Washington":
		p = "WA"
	case "West Virginia":
		p = "WV"
	case "Wisconsin":
		p = "WI"
	case "Wyoming":
		p = "WY"
	default:
		p = "XX"
	}
	return "_selected" + p
}
