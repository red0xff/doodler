package main

import (
    "fmt"
    "flag"
    "time"
    "net/http"
    "io/ioutil"
    "regexp"
    "strconv"
    "encoding/json"
    "os"
    "path/filepath"
    "sync"
)

var output string
var wg sync.WaitGroup

func main() {
	var startdate, enddate string
	var image, hdimage, full bool
	flag.StringVar(&startdate, "startdate", "1998/08", "First date to scrap")
	flag.StringVar(&enddate, "enddate", time.Now().Format("2006/01"), "Last date to scrap")
	flag.BoolVar(&image, "image", false, "Scrap the doodle image (not HD resolution)")
	flag.BoolVar(&hdimage, "hd-image", false, "Scrap the doodle image in HD")
	flag.BoolVar(&full, "full", false, "Query the full format (more informations)")
	flag.StringVar(&output, "output", ".", "Directory where to save the scrapped data")
	flag.Parse()

	date_regex := regexp.MustCompile("^\\d{4}/\\d{2}$")
	if m := date_regex.MatchString(startdate) ; !m {
		panic("Invalid startdate given")
	}
	
	if m := date_regex.MatchString(enddate) ; !m {
		panic("Invalid enddate given")
	}
	
	scrap(startdate, enddate, image, hdimage, full)
	wg.Wait()
}

var re = regexp.MustCompile("\\d+")

func ParseDate(date string) (int, int) {
	date_numbers := re.FindAllString(date, 2)
	year, err := strconv.Atoi(date_numbers[0])
	if err != nil {
		panic(err)
	}
	month, err := strconv.Atoi(date_numbers[1])
	if err != nil {
		panic(err)
	}
	return year, month
}

func scrap(startdate, enddate string, image, hdimage, isfull bool) {
	startyear, startmonth := ParseDate(startdate)
	endyear, endmonth := ParseDate(enddate)

	fmt.Printf("startyear : %v\nstartmonth : %d\nendyear : %d\nendmonth : %d\n", startyear, startmonth, endyear, endmonth)
	full := 0
	if isfull {
		full = 1
	}
	// Iterating over the dates
	
	if startyear == endyear {
		os.Mkdir(filepath.Join(output, strconv.Itoa(startyear)), os.ModePerm)
		for m := startmonth ; m <= endmonth ; m++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ScrapMonth(startyear, m, full, image, hdimage)
			}()
		}
	} else {
		// 1) Iterate over the remaining months in startyear
		os.Mkdir(filepath.Join(output, strconv.Itoa(startyear)), os.ModePerm)
		for m := startmonth ; m <= 12 ; m++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ScrapMonth(startyear, m, full, image, hdimage)
			}()
		}
		
		// 2) Iterate over all the months in the exclusive range ]startyear, endyear[
		for y := startyear+1 ; y < endyear ; y++ {
			os.Mkdir(filepath.Join(output, strconv.Itoa(y)), os.ModePerm)
			for m := 1 ; m <= 12 ; m++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ScrapMonth(y, m, full, image, hdimage)
				}()
			}
		}
		
		// 3) Iterate over the months until endmonth in endyear
		os.Mkdir(filepath.Join(output, strconv.Itoa(endyear)), os.ModePerm)
		for m := 1 ; m <= endmonth ; m++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ScrapMonth(endyear, m, full, image, hdimage)
			}()
		}
	}
}

func ScrapMonth(year, month int, full int, image, hdimage bool) {
	var path string = filepath.Join(output, fmt.Sprintf("%d/%d.json", year, month))
	if _, err := os.Stat(path); err == nil {
		return
	}
	data := ScrapData(year, month, full)
	fmt.Printf("year = %v ; month = %v ; data = %v\n", year, month, data)
	SaveData(path, data)
	if image {
		image_folder := filepath.Join(output, fmt.Sprintf("%d/%d_images", year, month))
		os.Mkdir(image_folder, os.ModePerm)
	
		var json_data []map[string] interface{}
		//fmt.Printf("data = %v\n", string(data))
		//fmt.Println("hi")
		json.Unmarshal(data, &json_data)
		//fmt.Printf("json_data = %v\n", json_data)
		for _, doodle := range(json_data) {
			url := doodle["url"]
			//fmt.Printf("url = %v\ndoodle = %v\n", url, doodle)
			wg.Add(1)
			go func() {
				defer wg.Done()
				DownloadImage(url.(string), image_folder)
			}()
		}
	}
	
	if hdimage {
		image_folder := filepath.Join(output, fmt.Sprintf("%d/%d_hd_images", year, month))
		os.Mkdir(image_folder, os.ModePerm)
		var json_data []map[string] interface{}
		json.Unmarshal(data, &json_data)
		
		for _, doodle := range(json_data) {
			url := doodle["high_res_url"]
			fmt.Printf("url = %v\ndoodle = %v\n", url, doodle)
			wg.Add(1)
			go func() {
				defer wg.Done()
				DownloadImage(url.(string), image_folder)
			}()
		}
	}
}

func ScrapData(year, month int, full int) ([]byte) {
	var url string = fmt.Sprintf("https://www.google.com/doodles/json/%d/%d?full=%d", year, month, full)
	return GetRequest(url)
}

/*func ScrapMonthParsing(year, month int, full int) ([]map[string]interface{}) {
	var res []map[string]interface{}
	var url string = fmt.Sprintf("https://www.google.com/doodles/json/%d/%d?full=%d", year, month, full)
	rs, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer rs.Body.Close()
	bodyBytes, err := ioutil.ReadAll(rs.Body)
    if err != nil {
        panic(err)
    }
    json.Unmarshal(bodyBytes, &res)
    return res
}*/

func SaveData(path string, data []byte) {
	err := ioutil.WriteFile(path, data, 0644)
	if err != nil {
		panic(err)
	}
}

func DownloadImage(url, path string) {
	filename_regex := regexp.MustCompile("/[^/]*$")
	var filepath string = filepath.Join(path, filename_regex.FindString(url))
	var image_data []byte = GetRequest(url)
	// save it
	ioutil.WriteFile(filepath, image_data, 0644)
}

func GetRequest(url string) ([]byte) {
	m, err := regexp.MatchString("^//", url)
	if m {
		url = fmt.Sprintf("https:%s", url)
	}
	rs, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer rs.Body.Close()
	body, err := ioutil.ReadAll(rs.Body)
    if err != nil {
        panic(err)
    }
    return body
}
