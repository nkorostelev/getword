package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sony/sonyflake"
)

var ApplicationDescription string = "Word Shuffle API"
var BuildVersion string = "0.0.0"

var Debug bool = false

type APICreateRequestStruct struct {
	Value string `json:"value"`
}

type APIReadRequestStruct struct {
}

type APIUpdateRequestStruct struct {
	ValueOld string `json:"value_old"`
	ValueNew string `json:"value_new"`
}

type APIDeleteRequestStruct struct {
	Value string `json:"value"`
}

type APICreateResponseStruct struct {
	Uid   uint64 `json:"uid"`
	Value string `json:"value"`
}

type APIReadResponseStruct struct {
	Uid   uint64 `json:"uid"`
	Value string `json:"value"`
}

type APIUpdateResponseStruct struct {
	Uid   uint64 `json:"uid"`
	Value string `json:"value"`
}

type APIDeleteResponseStruct struct {
	Uid   uint64 `json:"uid"`
	Value string `json:"value"`
}

type DatasetStruct struct {
	Words []string `json:"words"`
}

type ConfigurationStruct struct {
	Listen string `json:"listen"`
}

type handleSignalParamsStruct struct {
	httpServer http.Server
}

type MetricsStruct struct {
	Index    int32
	Warnings int32
	Errors   int32
	Create   int32
	Read     int32
	Update   int32
	Delete   int32
	Started  time.Time
}

var Configuration = ConfigurationStruct{}
var Dataset = DatasetStruct{}
var handleSignalParams = handleSignalParamsStruct{}

var MetricsNotifierPeriod int = 60
var Metrics = MetricsStruct{
	Index:    0,
	Warnings: 0,
	Errors:   0,
	Create:   0,
	Read:     0,
	Update:   0,
	Delete:   0,
	Started:  time.Now(),
}

var flake = sonyflake.NewSonyflake(sonyflake.Settings{})

func handleSignal() {

	log.Debug().Msg("Initialising signal handling function")

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	go func() {

		<-signalChannel

		err := handleSignalParams.httpServer.Shutdown(context.Background())

		if err != nil {
			log.Fatal().Err(err).Msgf("HTTP server Shutdown error")

		} else {
			log.Info().Msgf("HTTP server Shutdown complete")
		}

		log.Info().
			Int32("Index", Metrics.Index).
			Int32("Warnings", Metrics.Warnings).
			Int32("Errors", Metrics.Errors).
			Int32("Create", Metrics.Create).
			Int32("Read", Metrics.Read).
			Int32("Update", Metrics.Update).
			Int32("Delete", Metrics.Delete).
			Str("Started", Metrics.Started.Format("2006-01-02 15:04:05 MST")).
			Msgf("Counters")

		log.Warn().Msg("SIGINT")
		os.Exit(0)

	}()
}

func metricsNotifier() {
	go func() {
		for {
			time.Sleep(60 * time.Second)
			log.Info().
				Int32("Index", Metrics.Index).
				Int32("Warnings", Metrics.Warnings).
				Int32("Errors", Metrics.Errors).
				Int32("Create", Metrics.Create).
				Int32("Read", Metrics.Read).
				Int32("Update", Metrics.Update).
				Int32("Delete", Metrics.Delete).
				Str("Started", Metrics.Started.Format("2006-01-02 15:04:05 MST")).
				Msgf("Counters")
		}
	}()
}

func handlerIndex(rw http.ResponseWriter, req *http.Request) {
	_ = atomic.AddInt32(&Metrics.Index, 1)
	fmt.Fprintf(rw, "%s v%s\n", html.EscapeString(ApplicationDescription), html.EscapeString(BuildVersion))
}

func handlerMetrics(rw http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(rw, "# TYPE word_shuffle_req counter\n")
	fmt.Fprintf(rw, "# HELP Number of the requests to the API by type\n")
	fmt.Fprintf(rw, "word_shuffle_req{method=\"create\"} %v\n", Metrics.Create)
	fmt.Fprintf(rw, "word_shuffle_req{method=\"read\"} %v\n", Metrics.Read)
	fmt.Fprintf(rw, "word_shuffle_req{method=\"update\"} %v\n", Metrics.Update)
	fmt.Fprintf(rw, "word_shuffle_req{method=\"delete\"} %v\n", Metrics.Delete)

	fmt.Fprintf(rw, "# TYPE word_shuffle_errors counter\n")
	fmt.Fprintf(rw, "# HELP Number of the raised errors\n")
	fmt.Fprintf(rw, "word_shuffle_errors %v\n", Metrics.Errors)

	fmt.Fprintf(rw, "# TYPE word_shuffle_warnings counter\n")
	fmt.Fprintf(rw, "# HELP Number of the raised warnings\n")
	fmt.Fprintf(rw, "word_shuffle_warnings %v\n", Metrics.Warnings)

	fmt.Fprintf(rw, "# TYPE word_shuffle_index counter\n")
	fmt.Fprintf(rw, "# HELP Number of the requests to /\n")
	fmt.Fprintf(rw, "word_shuffle_index %v\n", Metrics.Index)

}

func handlerCreate(rw http.ResponseWriter, req *http.Request) {
	log.Warn().Msgf("Not implemented yet %s for %s", req.Method, req.RequestURI)
}

func handlerRead(rw http.ResponseWriter, req *http.Request) {

	var APIReadResponse APIReadResponseStruct

	_ = atomic.AddInt32(&Metrics.Read, 1)

	if req.Method != "GET" {
		log.Warn().Msgf("Ignoring unsupported http method %s", req.Method)
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	word := Dataset.Words[rand.Intn(len(Dataset.Words))]

	uid, err := flake.NextID()
	if err != nil {
		_ = atomic.AddInt32(&Metrics.Errors, 1)
		log.Err(err).Msgf("flake.NextID() failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	APIReadResponse.Uid = uid
	APIReadResponse.Value = word

	responseJSON, err := json.Marshal(APIReadResponse)
	if err != nil {
		log.Err(err).Msgf("APICreateResponseStruct marshall failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info().Msgf("Returning %s", string(responseJSON))
	fmt.Fprintf(rw, string(responseJSON))
}

func handlerUpdate(rw http.ResponseWriter, req *http.Request) {
	log.Warn().Msgf("Not implemented yet %s for %s", req.Method, req.RequestURI)
}

func handlerDelete(rw http.ResponseWriter, req *http.Request) {
	log.Warn().Msgf("Not implemented yet %s for %s", req.Method, req.RequestURI)
}

func main() {

	bindPtr := flag.String("bind", "127.0.0.1:8080", "Address and port to listen")
	datasetPathPtr := flag.String("dataset", "dataset.json", "Path to the JSON file with dataset")
	enableDebugPtr := flag.Bool("debug", false, "Enable verbose output")
	showVersionPtr := flag.Bool("version", false, "Show version")

	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Debug().Msg("Logger initialised")

	handleSignal()

	rand.Seed(time.Now().Unix())

	if *enableDebugPtr {
		log.Debug().Msg("Debug mode activated")
		Debug = true
	}

	if *showVersionPtr {
		fmt.Printf("%s\n", ApplicationDescription)
		fmt.Printf("Version: %s\n", BuildVersion)
		os.Exit(0)
	}

	if Debug {
		log.Debug().Msg("Periodical metrics reports activaded in debug mode")
		metricsNotifier()
	}

	datasetFile, err := os.Open(*datasetPathPtr)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error while opening the dataset file")
	}

	defer datasetFile.Close()
	log.Debug().Msgf("Opened dataset file %s", *datasetPathPtr)

	JSONDecoder := json.NewDecoder(datasetFile)

	err = JSONDecoder.Decode(&Dataset)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error while reading the dataset file")
	}

	log.Debug().Msgf("Successfully read the dataset file of '%v' words", len(Dataset.Words))

	listen_address, err := net.ResolveTCPAddr("tcp4", *bindPtr)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error while resolving bind address")
	}

	log.Info().Msgf("Listening on %s", listen_address.String())
	Configuration.Listen = listen_address.String()

	srv := &http.Server{
		Addr:         Configuration.Listen,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	handleSignalParams.httpServer = *srv

	http.HandleFunc("/", handlerIndex)
	http.HandleFunc("/metrics", handlerMetrics)
	http.HandleFunc("/create", handlerCreate)
	http.HandleFunc("/read", handlerRead)
	http.HandleFunc("/update", handlerUpdate)
	http.HandleFunc("/delete", handlerDelete)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msgf("HTTP server ListenAndServe error")
	}

}
