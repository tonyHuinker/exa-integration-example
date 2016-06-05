package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/tonyHuinker/ehop"
)

func ConvertResponseToJSONArray(resp *http.Response) map[string]interface{} {
	// Depending on the request, you may not need an array
	var mapp = make(map[string]interface{}, 0)
	//var mapp = make([]map[string]interface{}, 0)
	if err := json.NewDecoder(resp.Body).Decode(&mapp); err != nil {
		fmt.Printf("Could not parse results: %q", err.Error())
	}
	defer resp.Body.Close()
	return mapp
}

type Page struct {
	Title    string
	Body     []byte
	Records  []string
	Launches int
	LoadAvg  float64
	LoginAvg float64
}

func inputHandler(w http.ResponseWriter, r *http.Request) {
	p, _ := loadPage("input")
	renderTemplate(w, "input", p)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("body")
	title := r.FormValue("body")

	//Do all kinds of EXA searching here....
	myhop := ehop.NewEDAfromKey("keys")
	query := `{ "filter": { "operator": "and", "rules": [ { "field": "user", "operand": "` + user + `", "operator": "=" } ] }, "from": -90000000, "limit": 1000, "types": [ "~ica_close" ] }`
	response, _ := ehop.CreateEhopRequest("POST", "records/search", query, myhop)
	jsonData := ConvertResponseToJSONArray(response)
	records := ""
	loadTime := float64(0)
	loadTimeTotal := float64(0)
	loginTime := float64(0)
	loginTimeTotal := float64(0)
	program := ""
	clientType := ""
	launches := 0
	for key, value := range jsonData {
		if key == "records" {
			for _, value2 := range value.([]interface{}) {
				for key3, value3 := range value2.(map[string]interface{}) {
					if key3 == "_source" {
						for key4, value4 := range value3.(map[string]interface{}) {
							if key4 == "clientType" {

								clientType = value4.(string)
								launches = launches + 1
							}
							if key4 == "program" {
								program = value4.(string)

							}
							if key4 == "loadTime" {
								loadTime = value4.(float64)
								loadTimeTotal = loadTime + value4.(float64)
							}
							if key4 == "loginTime" {
								loginTime = value4.(float64)
								loginTimeTotal = loginTime + value4.(float64)
							}
						}
						records = records + "Launch of application " + program + " using " + clientType + " had a login time of " + strconv.FormatFloat(loginTime, 'f', -1, 64) + "sec and a load time of " + strconv.FormatFloat(loadTime, 'f', -1, 64) + "sec,"
						loginTime = 0
						loadTime = 0
					}
				}
			}
		}
	}

	//Create new page based on results of EXA search
	p := &Page{Title: title, Body: []byte("Results"), Records: strings.Split(records, ","), LoadAvg: loadTimeTotal / float64(launches), LoginAvg: loginTimeTotal / float64(launches), Launches: launches}
	renderTemplate(w, "view", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, p)
}
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	http.HandleFunc("/input/", inputHandler)
	http.HandleFunc("/search/", searchHandler)
	http.ListenAndServe(":8080", nil)
}
