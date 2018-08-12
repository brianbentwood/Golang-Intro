package controller

import (
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
)

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

	startSC := model.SearchCache{zipcode, city, usstate, 1, "TBD", "TBD", "TBD", now, 0}

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

			if round(dur.Minutes()) <= 30 {
				startSC.CurrTemp = searchcache.CurrTemp
				startSC.HighTemp = searchcache.HighTemp
				startSC.LowTemp = searchcache.LowTemp
				sess.AddFlash(view.Flash{"Got From Cache: " + searchcache.CurrTemp, view.FlashError})
				//sess.Values["_currtemp"] = searchcache.CurrTemp
				//sess.Values["_hightemp"] = searchcache.HighTemp
				//sess.Values["_lowtemp"] = searchcache.LowTemp
				//sess.Values["token"] = csrfbanana.Token(w, r, sess)
				sess.Save(r, w)
				//SearchGET(w, r)

				v := view.New(r)
				v.Name = "search/search"
				v.Vars["_city"] = startSC.City
				v.Vars["_state"] = startSC.State
				v.Vars["_zipcode"] = startSC.Zipcode
				v.Vars["_currtemp"] = startSC.CurrTemp
				v.Vars["_hightemp"] = startSC.HighTemp
				v.Vars["_lowtemp"] = startSC.LowTemp
				v.Vars["token"] = csrfbanana.Token(w, r, sess)
				//v.Vars["_ifsrc"] = "https://weather.com/weather/today/l/" + searchcache.Zipcode + ":4:US"
				v.Vars["_ifsrc"] = "https://www.foreca.com/United_States/" + startSC.State + "/" + startSC.City
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
	//resp, err := client.Get("https://weather.com/weather/today/l/" + startSC.Zipcode + ":4:US")
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

	//for j := 0; j < len(rbody); j++ {
	//	log.Println("Pulled \"", string(rbody[j]), "\" [", j, "]")
	//}
	//return

	/*   This is for Weather.com

	//Get Current Tempurature
	currtempStr := GetTagSubString(0, "div", sbody, "\"today_nowcard-temp\"")
	log.Println("Curr Temp String: ", currtempStr)
	currtempStr = GetTagSubString(0, "span", currtempStr, "")
	log.Println("Curr Temp String: ", currtempStr)
	currtempStr = KeepNumbers(currtempStr)
	log.Println("Curr Temp String: ", currtempStr)

	//Get Weather Phrase
	weatherphraseStr := GetTagSubString(0, "div", sbody, "\"today_nowcard-phrase\"")
	log.Println("Phrase String: ", weatherphraseStr)

	//Get Feels Like
	feelslikeStr1 := GetTagSubString(0, "div", sbody, "\"today_nowcard-feels\"")
	log.Println("Feels Like: ", feelslikeStr1)
	feelslikeStr1 = GetTagSubString(0, "span", feelslikeStr1, "")
	log.Println("Feels Like: ", feelslikeStr1)
	feelslikeStr2 := GetTagSubString(0, "div", sbody, "\"today_nowcard-feels\"")
	log.Println("Feels Like: ", feelslikeStr2)
	feelslikeStr2 = GetTagSubString(0, "span", feelslikeStr2, "\"deg-feels\"")
	log.Println("Feels Like: ", feelslikeStr2)
	feelslikeStr2 = KeepNumbers(feelslikeStr2)
	log.Println("Feels Like: ", feelslikeStr2)

	//Get High Temp
	hightempStr := GetTagSubString(0, "span", sbody, "\"deg-hilo-nowcard\"")
	log.Println("High Temp String: ", hightempStr)
	hightempStr = KeepNumbers(hightempStr)
	log.Println("High Temp String: ", hightempStr)

	//Get Low Temp
	i := strings.Index(sbody, "\"deg-hilo-nowcard\"") + 1
	lowtempStr := GetTagSubString(int32(i), "span", sbody, "\"deg-hilo-nowcard\"")
	log.Println("Low Temp String: ", lowtempStr)
	lowtempStr = KeepNumbers(lowtempStr)
	log.Println("Low Temp String: ", lowtempStr)

	*/

	/*   This is for Foreca.com  */

	//Get Current Tempurature
	currtempStr := GetTagSubString(0, "span", sbody, rbody, "txt-xxlarge")
	log.Println("Curr Temp StringA: ", currtempStr)
	currtempStr = GetTagSubString(0, "strong", currtempStr, []rune(currtempStr), "")
	log.Println("Curr Temp StringB: ", currtempStr)

	//Get Weather Phrase
	weatherdetailsStr := GetTagSubString(0, "div", sbody, rbody, "right txt-tight")
	//log.Println("Details String: ", weatherdetailsStr)

	weatherphraseStr := strings.TrimSpace(weatherdetailsStr[0:strings.Index(weatherdetailsStr, "<br")])
	//r := utf8.RuneCountInString(weatherdetailsStr[0 : strings.Index(weatherdetailsStr, "<br")])
	//PrintRuneClump(bigrune, rrr, 4)
	log.Println("Phrase String: ", weatherphraseStr)

	//Get Feels Like
	feelslikeStr1 := GetTagSubString(0, "strong", weatherdetailsStr, []rune(weatherdetailsStr), "Feels Like:")
	log.Println("Feels LikeA: ", feelslikeStr1)
	feelslikeStr1 = KeepNumbers(feelslikeStr1)
	log.Println("Feels LikeB: ", feelslikeStr1)

	//Get High Temp
	hightempStr := GetTagSubString(0, "div", sbody, rbody, "minmax")
	log.Println("High Temp StringA: ", hightempStr)
	hightempStr = GetTagSubString(0, "span", hightempStr, []rune(hightempStr), "")
	log.Println("High Temp StringB: ", hightempStr)
	hightempStr = KeepNumbers(hightempStr)
	log.Println("High Temp StringC: ", hightempStr)

	//Get Low Temp
	lowtempStr := GetTagSubString(0, "div", sbody, rbody, "minmax")
	log.Println("Low Temp StringA: ", lowtempStr)
	i := strings.Index(lowtempStr, "Lo:") + 1
	lowtempStr = GetTagSubString(i, "strong", lowtempStr, []rune(lowtempStr), "")
	log.Println("Low Temp StringB: ", lowtempStr)
	lowtempStr = KeepNumbers(lowtempStr)
	log.Println("Low Temp StringC: ", lowtempStr)

	newSC := model.SearchCache{startSC.Zipcode, startSC.City, startSC.State, 1, currtempStr, hightempStr, lowtempStr, now, 0}
	//newSC.Zipcode = zipcode
	//newSC.CurrTemp = "80"
	//newSC.HighTemp = "90"
	//newSC.LowTemp = "70"

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
		v.Vars["_currtemp"] = newSC.CurrTemp
		v.Vars["_hightemp"] = newSC.HighTemp
		v.Vars["_lowtemp"] = newSC.LowTemp
		v.Vars["token"] = csrfbanana.Token(w, r, sess)
		//v.Vars["_ifsrc"] = "https://weather.com/weather/today/l/" + newSC.Zipcode + ":4:US"
		v.Vars["_ifsrc"] = "https://www.foreca.com/United_States/" + newSC.State + "/" + newSC.City
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

//PrintRuneClump will print a few runes
func PrintRuneClump(runes []rune, start int, clumpsize int) {
	log.Println("Rune \"", string(runes[start:start+clumpsize]), "\"")
}
