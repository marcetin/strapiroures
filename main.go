package main

import (
	"flag"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/handlers"
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
	log.Fatal().Err(http.ListenAndServe(":"+c.appport, handlers.CORS()(handlers.CompressHandler(InterceptHandler(r, DefaultErrorHandler)))))
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

// CommonMiddleware --Set content-type
func CommonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header")
		next.ServeHTTP(w, r)
	})
}

type interceptResponseWriter struct {
	http.ResponseWriter
	errH func(http.ResponseWriter, int)
}

func (w *interceptResponseWriter) WriteHeader(status int) {
	if status >= http.StatusBadRequest {
		w.errH(w.ResponseWriter, status)
		w.errH = nil
	} else {
		w.ResponseWriter.WriteHeader(status)
	}
}

type ErrorHandler func(http.ResponseWriter, int)

func (w *interceptResponseWriter) Write(p []byte) (n int, err error) {
	if w.errH == nil {
		return len(p), nil
	}
	return w.ResponseWriter.Write(p)
}

func DefaultErrorHandler(w http.ResponseWriter, status int) {
	//t := template.Must(template.ParseFiles("errors/error.html"))
	//w.Header().Set("Content-Type", "text/html")
	//t.Execute(w, map[string]interface{}{"status": status})
	w.Header().Set("Content-Type", "text/html")
	//tpl.TemplateHandler(cfg.Path).ExecuteTemplate(w, "error_gohtml", map[string]interface{}{"status": status})
}

func InterceptHandler(next http.Handler, errH ErrorHandler) http.Handler {
	if errH == nil {
		errH = DefaultErrorHandler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&interceptResponseWriter{w, errH}, r)
	})
}
