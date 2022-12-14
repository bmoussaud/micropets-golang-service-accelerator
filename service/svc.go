package petkind

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	. "moussaud.org/petkind/internal"
)

// PetKind Struct
type PetKind struct {
	Index int
	Name  string
	Kind  string
	Age   int
	URL   string
	From  string
	URI   string
}

// petkind Struct
type PetKinds struct {
	Total    int
	Hostname string
	PetKinds []PetKind `json:"Pets"`
}

var calls = 0

var shift = 0

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func db_authentication(r *http.Request) {
	span := NewServerSpan(r, "db_authentication")
	defer span.Finish()

	if GlobalConfig.Service.Delay.Period > 0 {
		time.Sleep(time.Duration(2) * time.Millisecond)
	}
}

func db() PetKinds {
	host, err := os.Hostname()
	if err != nil {
		host = "Unknown"
	}

	petkind := PetKinds{4,
		host,
		[]PetKind{
			{70, "Kaa", "An Indian Python", 4,
				"http://personnages-disney.com/Images/Vignettes%20bleues%20V4/Kaa.png", GlobalConfig.Service.From, "/petkind/v1/data/0"},
			{71, "Nagini", "python-viper", 11,
				"https://figurinepop.com/public/2018/12/nagini1.jpg", GlobalConfig.Service.From, "/petkind/v1/data/1"},
			{72, "Roger", "King Cobras", 12,
				"https://upload.wikimedia.org/wikipedia/commons/4/4d/12_-_The_Mystical_King_Cobra_and_Coffee_Forests.jpg", GlobalConfig.Service.From, "/petkind/v1/data/2"},
			{73, "Pataya", "Green Snake", 27,
				"https://images.itnewsinfo.com/lmi/articles/grande/000000070110.jpg", GlobalConfig.Service.From, "/petkind/v1/data/3"}}}

	return petkind
}

func index(w http.ResponseWriter, r *http.Request) {
	span := NewServerSpan(r, "index")
	defer span.Finish()

	time.Sleep(time.Duration(3) * time.Millisecond)

	setupResponse(&w, r)

	calls = calls + 1
	petkind := db()

	time.Sleep(time.Duration(len(petkind.PetKinds)) * time.Millisecond)
	db_authentication(r)

	if GlobalConfig.Service.Mode == "RANDOM_NUMBER" {
		total := rand.Intn(petkind.Total) + 1
		fmt.Printf("total %d\n", total)
		for i := 1; i < total; i++ {
			petkind.PetKinds = petkind.PetKinds[:len(petkind.PetKinds)-1]
			petkind.Total = petkind.Total - 1
		}
	}

	js, err := json.Marshal(petkind)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if GlobalConfig.Service.Delay.Period > 0 {
		y := float64(calls+shift) * math.Pi / float64(2*GlobalConfig.Service.Delay.Period)
		sin_y := math.Sin(y)
		abs_y := math.Abs(sin_y)
		sleep := int(abs_y * GlobalConfig.Service.Delay.Amplitude * 1000.0)
		fmt.Printf("waitTime %d - %f - %f - %f  -> sleep %d ms  \n", calls, y, math.Sin(y), abs_y, sleep)
		start := time.Now()
		time.Sleep(time.Duration(sleep) * time.Millisecond)
		elapsed := time.Since(start)
		fmt.Printf("Current Unix Time: %s\n", elapsed)
	}

	if GlobalConfig.Service.FrequencyError > 0 && calls%GlobalConfig.Service.FrequencyError == 0 {
		fmt.Printf("Fails this call (%d)", calls)
		http.Error(w, "Unexpected Error when querying the petkind repository", http.StatusServiceUnavailable)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func single(w http.ResponseWriter, r *http.Request) {

	span := NewServerSpan(r, "single")
	defer span.Finish()

	setupResponse(&w, r)
	time.Sleep(time.Duration(10) * time.Millisecond)

	db_authentication(r)
	petkind := db()

	re := regexp.MustCompile(`/`)
	submatchall := re.Split(r.URL.Path, -1)
	id, _ := strconv.Atoi(submatchall[4])

	if id >= len(petkind.PetKinds) {
		http.Error(w, fmt.Sprintf("invalid index %d", id), http.StatusInternalServerError)
	} else {
		element := petkind.PetKinds[id]
		fmt.Println(element)
		w.Header().Set("Content-Type", "application/json")
		js, err := json.Marshal(element)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(js)
	}
}

// GetLocation returns the full path of the config file based on the current executable location or using SERVICE_CONFIG_DIR env
func GetLocation(file string) string {
	if serviceConfigDirectory := os.Getenv("SERVICE_CONFIG_DIR"); serviceConfigDirectory != "" {
		fmt.Printf("Load configuration from %s\n", serviceConfigDirectory)
		return filepath.Join(serviceConfigDirectory, file)
	} else {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath := filepath.Dir(ex)
		return filepath.Join(exPath, file)
	}
}

func readiness_and_liveness(w http.ResponseWriter, r *http.Request) {
	span := NewServerSpan(r, "readiness_and_liveness")
	defer span.Finish()

	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func Start() {
	config := LoadConfiguration()

	http.HandleFunc("/petkind/v1/data", index)
	http.HandleFunc("/petkind/v1/data/", single)

	http.HandleFunc("/petkind/liveness", readiness_and_liveness)
	http.HandleFunc("/petkind/readiness", readiness_and_liveness)

	http.HandleFunc("/liveness", readiness_and_liveness)
	http.HandleFunc("/readiness", readiness_and_liveness)

	http.HandleFunc("/", index)

	rand.Seed(time.Now().UnixNano())
	shift = rand.Intn(100)

	fmt.Printf("******* Starting to the petkind service on port %s, mode %s\n", config.Service.Port, config.Service.Mode)
	fmt.Printf("******* Delay Period %d Amplitude %f shift %d \n", config.Service.Delay.Period, config.Service.Delay.Amplitude, shift)
	fmt.Printf("******* Frequency Error %d\n", config.Service.FrequencyError)
	log.Fatal(http.ListenAndServe(config.Service.Port, nil))
}
