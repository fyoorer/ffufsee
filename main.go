package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/ffuf/ffuf/pkg/ffuf"
)

type PerDomainFilter struct {
	IsCalibrated bool
	Filters      map[string]ffuf.FilterProvider
}

type optRange struct {
	Min      float64
	Max      float64
	IsRange  bool
	HasDelay bool
}

type Config struct {
	AutoCalibration         bool                  `json:"autocalibration"`
	AutoCalibrationKeyword  string                `json:"autocalibration_keyword"`
	AutoCalibrationPerHost  bool                  `json:"autocalibration_perhost"`
	AutoCalibrationStrategy string                `json:"autocalibration_strategy"`
	AutoCalibrationStrings  []string              `json:"autocalibration_strings"`
	Cancel                  context.CancelFunc    `json:"-"`
	Colors                  bool                  `json:"colors"`
	CommandKeywords         []string              `json:"-"`
	CommandLine             string                `json:"cmdline"`
	ConfigFile              string                `json:"configfile"`
	Context                 context.Context       `json:"-"`
	Data                    string                `json:"postdata"`
	Debuglog                string                `json:"debuglog"`
	Delay                   optRange              `json:"delay"`
	DirSearchCompat         bool                  `json:"dirsearch_compatibility"`
	Extensions              []string              `json:"extensions"`
	FilterMode              string                `json:"fmode"`
	FollowRedirects         bool                  `json:"follow_redirects"`
	Headers                 map[string]string     `json:"headers"`
	IgnoreBody              bool                  `json:"ignorebody"`
	IgnoreWordlistComments  bool                  `json:"ignore_wordlist_comments"`
	InputMode               string                `json:"inputmode"`
	InputNum                int                   `json:"cmd_inputnum"`
	InputProviders          []InputProviderConfig `json:"inputproviders"`
	InputShell              string                `json:"inputshell"`
	Json                    bool                  `json:"json"`
	Matchers                struct {
		IsCalibrated bool `json:"IsCalibrated"`
		Mutex        struct {
		} `json:"Mutex"`
		Matchers struct {
			Status struct {
				Value string `json:"value"`
			} `json:"status"`
		} `json:"Matchers"`
		Filters struct {
		} `json:"Filters"`
		PerDomainFilters struct {
		} `json:"PerDomainFilters"`
	} `json:"matchers"`
	MatcherMode         string   `json:"mmode"`
	MaxTime             int      `json:"maxtime"`
	MaxTimeJob          int      `json:"maxtime_job"`
	Method              string   `json:"method"`
	Noninteractive      bool     `json:"noninteractive"`
	OutputDirectory     string   `json:"outputdirectory"`
	OutputFile          string   `json:"outputfile"`
	OutputFormat        string   `json:"outputformat"`
	OutputSkipEmptyFile bool     `json:"OutputSkipEmptyFile"`
	ProgressFrequency   int      `json:"-"`
	ProxyURL            string   `json:"proxyurl"`
	Quiet               bool     `json:"quiet"`
	Rate                int64    `json:"rate"`
	Recursion           bool     `json:"recursion"`
	RecursionDepth      int      `json:"recursion_depth"`
	RecursionStrategy   string   `json:"recursion_strategy"`
	ReplayProxyURL      string   `json:"replayproxyurl"`
	RequestFile         string   `json:"requestfile"`
	RequestProto        string   `json:"requestproto"`
	ScraperFile         string   `json:"scraperfile"`
	Scrapers            string   `json:"scrapers"`
	SNI                 string   `json:"sni"`
	StopOn403           bool     `json:"stop_403"`
	StopOnAll           bool     `json:"stop_all"`
	StopOnErrors        bool     `json:"stop_errors"`
	Threads             int      `json:"threads"`
	Timeout             int      `json:"timeout"`
	Url                 string   `json:"url"`
	Verbose             bool     `json:"verbose"`
	Wordlists           []string `json:"wordlists"`
	Http2               bool     `json:"http2"`
}

type InputProviderConfig struct {
	Name     string `json:"name"`
	Keyword  string `json:"keyword"`
	Value    string `json:"value"`
	Template string `json:"template"` // the templating string used for sniper mode (usually "§")
}

type JsonResult struct {
	Input            map[string]string   `json:"input"`
	Position         int                 `json:"position"`
	StatusCode       int64               `json:"status"`
	ContentLength    int64               `json:"length"`
	ContentWords     int64               `json:"words"`
	ContentLines     int64               `json:"lines"`
	ContentType      string              `json:"content-type"`
	RedirectLocation string              `json:"redirectlocation"`
	ScraperData      map[string][]string `json:"scraper"`
	Duration         time.Duration       `json:"duration"`
	ResultFile       string              `json:"resultfile"`
	Url              string              `json:"url"`
	Host             string              `json:"host"`
	HTMLColor        string              `json:"-"`
}

type htmlResult struct {
	Input            map[string]string
	Position         int
	StatusCode       int64
	ContentLength    int64
	ContentWords     int64
	ContentLines     int64
	ContentType      string
	RedirectLocation string
	ScraperData      string
	Duration         time.Duration
	ResultFile       string
	Url              string
	Host             string
	HTMLColor        string
	FfufHash         string
}

type htmlFileOutput struct {
	CommandLine string
	Time        string
	Keys        []string
	Results     []htmlResult
}

const (
	htmlTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <meta
      name="viewport"
      content="width=device-width, initial-scale=1, maximum-scale=1.0"
    />
    <title>FFUF Report - </title>

    <!-- CSS  -->
    <link
      href="https://fonts.googleapis.com/icon?family=Material+Icons"
      rel="stylesheet"
    />
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/css/materialize.min.css"
	/>
	<link 
	  rel="stylesheet" 
	  type="text/css" 
	  href="https://cdn.datatables.net/1.10.20/css/jquery.dataTables.css"
	/>
  
  </head>

  <body>
    <nav>
      <div class="nav-wrapper">
        <a href="#" class="brand-logo">FFUF</a>
        <ul id="nav-mobile" class="right hide-on-med-and-down">
        </ul>
      </div>
    </nav>

    <main class="section no-pad-bot" id="index-banner">
      <div class="container">
        <br /><br />
        <h1 class="header center ">FFUF Report</h1>
        <div class="row center">

		<pre>{{ .CommandLine }}</pre>
		<pre>{{ .Time }}</pre>

   <table id="ffufreport">
        <thead>
        <div style="display:none">
|result_raw|StatusCode{{ range $keyword := .Keys }}|{{ $keyword | printf "%s" }}{{ end }}|Url|RedirectLocation|Position|ContentLength|ContentWords|ContentLines|ContentType|Duration|Resultfile|ScraperData|FfufHash|
        </div>
          <tr>
              <th>Status</th>
{{ range .Keys }}              <th>{{ . }}</th>{{ end }}
			  <th>URL</th>
			  <th>Redirect location</th>
              <th>Position</th>
              <th>Length</th>
              <th>Words</th>
			  <th>Lines</th>
			  <th>Type</th>
              <th>Duration</th>
			  <th>Resultfile</th>
              <th>Scraper data</th>
              <th>Ffuf Hash</th>
          </tr>
        </thead>

        <tbody>
			{{range $result := .Results}}
                <div style="display:none">
|result_raw|{{ $result.StatusCode }}{{ range $keyword, $value := $result.Input }}|{{ $value | printf "%s" }}{{ end }}|{{ $result.Url }}|{{ $result.RedirectLocation }}|{{ $result.Position }}|{{ $result.ContentLength }}|{{ $result.ContentWords }}|{{ $result.ContentLines }}|{{ $result.ContentType }}|{{ $result.Duration }}|{{ $result.ResultFile }}|{{ $result.ScraperData }}|{{ $result.FfufHash }}|
                </div>
                <tr class="result-{{ $result.StatusCode }}" style="background-color: {{$result.HTMLColor}};">
                    <td><font color="black" class="status-code">{{ $result.StatusCode }}</font></td>
                    {{ range $keyword, $value := $result.Input }}
                        <td>{{ $value | printf "%s" }}</td>
                    {{ end }}
                    <td><a href="{{ $result.Url }}">{{ $result.Url }}</a></td>
                    <td><a href="{{ $result.RedirectLocation }}">{{ $result.RedirectLocation }}</a></td>
                    <td>{{ $result.Position }}</td>
                    <td>{{ $result.ContentLength }}</td>
                    <td>{{ $result.ContentWords }}</td>
					<td>{{ $result.ContentLines }}</td>
					<td>{{ $result.ContentType }}</td>
					<td>{{ $result.Duration }}</td>
                    <td>{{ $result.ResultFile }}</td>
					<td>{{ $result.ScraperData }}</td>
					<td>{{ $result.FfufHash }}</td>
                </tr>
            {{ end }}
        </tbody>
      </table>

        </div>
        <br /><br />
      </div>
    </main>

    <!--JavaScript at end of body for optimized loading-->
	<script src="https://code.jquery.com/jquery-3.4.1.min.js" integrity="sha256-CSXorXvZcTkaix6Yvo6HppcZGetbYMGWSFlBw8HfCJo=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/js/materialize.min.js"></script>
    <script type="text/javascript" charset="utf8" src="https://cdn.datatables.net/1.10.20/js/jquery.dataTables.js"></script>
    <script>
    $(document).ready(function() {
        $('#ffufreport').DataTable(
            {
                "aLengthMenu": [
                    [250, 500, 1000, 2500, -1],
                    [250, 500, 1000, 2500, "All"]
                ]
            }
        )
        $('select').formSelect();
        });
    </script>
    <style>
      body {
        display: flex;
        min-height: 100vh;
        flex-direction: column;
      }

      main {
        flex: 1 0 auto;
      }
    </style>
  </body>
</html>

	`
)

// colorizeResults returns a new slice with HTMLColor attribute
func colorizeResults(results []JsonResult) []JsonResult {
	newResults := make([]JsonResult, 0)

	for _, r := range results {
		result := r
		result.HTMLColor = "black"

		s := result.StatusCode

		if s >= 200 && s <= 299 {
			result.HTMLColor = "#adea9e"
		}

		if s >= 300 && s <= 399 {
			result.HTMLColor = "#bbbbe6"
		}

		if s >= 400 && s <= 499 {
			result.HTMLColor = "#d2cb7e"
		}

		if s >= 500 && s <= 599 {
			result.HTMLColor = "#de8dc1"
		}

		newResults = append(newResults, result)
	}

	return newResults
}

func writeHTML(config Config, results []JsonResult) error {
	results = colorizeResults(results)

	ti := time.Now()

	keywords := make([]string, 0)
	for _, inputprovider := range config.InputProviders {
		keywords = append(keywords, inputprovider.Keyword)
	}
	htmlResults := make([]htmlResult, 0)

	for _, r := range results {
		ffufhash := ""
		strinput := make(map[string]string)
		for k, v := range r.Input {
			if k == "FFUFHASH" {
				ffufhash = string(v)
			} else {
				strinput[k] = string(v)
			}
		}
		strscraper := ""
		for k, v := range r.ScraperData {
			if len(v) > 0 {
				strscraper = strscraper + "<p><b>" + html.EscapeString(k) + ":</b><br />"
				firstval := true
				for _, val := range v {
					if !firstval {
						strscraper += "<br />"
					}
					strscraper += html.EscapeString(val)
					firstval = false
				}
				strscraper += "</p>"
			}
		}
		hres := htmlResult{
			Input:            strinput,
			Position:         r.Position,
			StatusCode:       r.StatusCode,
			ContentLength:    r.ContentLength,
			ContentWords:     r.ContentWords,
			ContentLines:     r.ContentLines,
			ContentType:      r.ContentType,
			RedirectLocation: r.RedirectLocation,
			ScraperData:      strscraper,
			Duration:         r.Duration,
			ResultFile:       r.ResultFile,
			Url:              r.Url,
			Host:             r.Host,
			HTMLColor:        r.HTMLColor,
			FfufHash:         ffufhash,
		}
		htmlResults = append(htmlResults, hres)
	}
	outHTML := htmlFileOutput{
		Time:        ti.Format(time.RFC3339),
		Results:     htmlResults,
		CommandLine: config.CommandLine,
		Keys:        keywords,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templateName := "index"
		t := template.New(templateName).Delims("{{", "}}")
		_, err := t.Parse(htmlTemplate)
		if err != nil {
			fmt.Println("Failed to parse HTML template:", err)
			return
		}
		err = t.ExecuteTemplate(w, "index", outHTML)

	})

	// Start the server
	fmt.Println("Starting server...")
	fmt.Println("View the report in the browser at http://localhost:5505")
	err := http.ListenAndServe(":5505", nil)
	return err
}

func main() {
	// Check if path to the JSON file is provided
	if len(os.Args) < 2 {
		fmt.Println("Please provide the path to the JSON file as a command line argument")
		return
	}
	path := os.Args[1]

	// Load the JSON data from file
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Failed to load JSON file:", err)
		return
	}

	// Unmarshal the JSON into a map[string]json.RawMessage
	var data map[string]json.RawMessage
	err = json.Unmarshal(file, &data)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// Parse the JSON data
	var ffufResult []JsonResult
	err = json.Unmarshal(data["results"], &ffufResult)
	if err != nil {
		fmt.Println("Failed to parse JSON into ffuf result structure:", err)
		return
	}

	var ffufConfig Config
	err = json.Unmarshal(data["config"], &ffufConfig)
	if err != nil {
		fmt.Println("Failed to parse JSON into ffuf config structure:", err)
		return
	}

	err = writeHTML(ffufConfig, ffufResult)
	if err != nil {
		fmt.Println("Something went wrong:", err)
		return
	}

}
