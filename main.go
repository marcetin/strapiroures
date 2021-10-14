package main

import (
	"flag"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type conf struct {
	src, srcport, appport string
}

func main() {
	// Get cmd line parameters
	src := flag.String("src", "http://127.0.0.1", "Source")
	srcport := flag.String("srcport", "11111", "Source Port")
	appport := flag.String("appport", "8080", "Application Port")
	loglevel := flag.String("loglevel", "debug", "Logging level (debug, info, warn, error)")
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Default level for this example is info, unless debug flag is present
	switch *loglevel {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	c := &conf{
		src:     *src,
		srcport: *srcport,
		appport: *appport,
	}

	r := mux.NewRouter()
	r.HandleFunc("/{col}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		c.get(&w, vars["col"])
	})
	r.HandleFunc("/{col}/{slug}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		c.get(&w, vars["col"]+"?slug="+vars["slug"])
	})
	r.StrictSlash(true)
	log.Print("Starting up on port " + c.appport)
	log.Fatal().Err(http.ListenAndServe(":"+c.appport, r))
}

func ErrorLog(err error) {
	if err != nil {
		log.Print("Error", err)
	}
}

func (c *conf) get(w *http.ResponseWriter, path string) {
	respc, err := http.Get(c.src + ":" + c.srcport + "/" + path)
	ErrorLog(err)
	defer respc.Body.Close()
	mapBody, err := ioutil.ReadAll(respc.Body)
	if mapBody != nil {
		(*w).Write(mapBody)
	}
}
