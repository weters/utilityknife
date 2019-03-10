package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const serviceName = "weters/utilityknife"

type Service struct {
	mux       *http.ServeMux
	dataDir   string
	hostname  string
	ipAddress string
}

type keyValueData struct {
	Key         string
	Value       []byte
	ContentType string
}

type responseData struct {
	ServerHostname string            `json:"serverHostname"`
	ServerIP       string            `json:"serverIP"`
	ServerDatetime time.Time         `json:"serverDatetime"`
	Links          map[string]string `json:"_links"`
}

func New(dataDir string) *Service {
	s := &Service{
		dataDir:   dataDir,
		hostname:  getHostname(),
		ipAddress: getIPAddress(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.rootHandler())
	mux.HandleFunc("/json", s.jsonHandler())
	mux.HandleFunc("/echo/", s.echoHandler())
	mux.HandleFunc("/data/", s.dataHandler())

	s.mux = mux

	return s
}

func (s *Service) getResponseData() responseData {
	return responseData{
		ServerHostname: s.hostname,
		ServerIP:       s.ipAddress,
		ServerDatetime: time.Now(),
		Links: map[string]string{
			"/":     "shows server data in HTML",
			"/json": "shows server data in JSON",
			"/echo": "echos request in response",
			"/data": "basic key/value storage on server",
		},
	}
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Hostname", s.hostname)
	w.Header().Set("X-IP", s.ipAddress)
	w.Header().Set("X-Served-By", serviceName)

	s.mux.ServeHTTP(w, r)
}

func (s *Service) jsonHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if err := enc.Encode(s.getResponseData()); err != nil {
			log.Printf("error: could not encode JSON payload: %v", err)
		}
	}
}

func (s *Service) rootHandler() http.HandlerFunc {
	const html = `<!DOCTYPE html>
<html lang="en">
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Utility Knife</title>
<style>
* { box-sizing: border-box; padding: 0; margin: 0; outline: none } 
html, body { font-family: sans-serif; padding: 20px }
h1, h2, table { margin-bottom: 20px }
li { margin-left: 40px; }
table { border-collapse: collapse }
table td { border: 1px solid #000; padding: 5px }
</style>
<main>
	<h1>Utility Knife</h1>

	<table>
	<tbody>
		<tr>
			<td>Server Date &amp; Time</td>
			<td>{{.ServerDatetime.Format "January 2, 2006 3:04:05 pm MST"}}</td>
		</tr>
		<tr>
			<td>Server Hostname</td>
			<td>{{.ServerHostname}}</td>
		</tr>
		<tr>
			<td>Server IP Address</td>
			<td>{{.ServerIP}}</td>
		</tr>
	</tbody>
	</table>

	<h2>Other Endpoints</h2>

	<ul>
		{{ range $link, $desc := .Links }}
			<li><a href={{$link}}>{{$link}}</a> - {{$desc}}
		{{ end }}
	</ul>
</main>
</html>`

	tpl := template.Must(template.New("").Parse(html))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		tpl.Execute(w, s.getResponseData())
	}
}

func (s *Service) echoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Printf("error: could not dump request: %v", err)
		}

		log.Println(string(dump))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(dump)
	}
}

func (s *Service) dataHandler() http.HandlerFunc {
	lock := sync.RWMutex{}

	del := func(w http.ResponseWriter, r *http.Request, filename string) error {
		lock.Lock()
		defer lock.Unlock()

		err := os.Remove(filename)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusAccepted)
		return nil
	}

	put := func(w http.ResponseWriter, r *http.Request, filename string) error {
		lock.Lock()
		defer lock.Unlock()

		file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		enc := json.NewEncoder(file)
		if err := enc.Encode(keyValueData{
			ContentType: r.Header.Get("Content-Type"),
			Key:         r.URL.Path,
			Value:       b,
		}); err != nil {
			return err
		}

		w.WriteHeader(http.StatusCreated)
		return nil
	}

	get := func(w http.ResponseWriter, r *http.Request, filename string) error {
		lock.RLock()
		defer lock.RUnlock()

		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		dec := json.NewDecoder(file)
		var kvd keyValueData
		if err := dec.Decode(&kvd); err != nil {
			return err
		}

		w.Header().Set("Content-Type", kvd.ContentType)
		w.Write(kvd.Value)
		return nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var fn func(http.ResponseWriter, *http.Request, string) error

		switch r.Method {
		case http.MethodPut:
			fn = put
		case http.MethodGet:
			fn = get
		case http.MethodDelete:
			fn = del
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		sum := md5.Sum([]byte(r.URL.Path))
		key := hex.EncodeToString(sum[:])
		filename := filepath.Join(s.dataDir, key)

		if err := fn(w, r, filename); err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}

			log.Printf("error: could not %s %s: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	return hostname
}

func getIPAddress() string {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	udpAddr := conn.LocalAddr().(*net.UDPAddr)
	return udpAddr.IP.String()
}
