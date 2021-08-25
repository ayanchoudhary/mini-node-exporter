package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
)

var node_load = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "node_load",
	},
	[]string{"duration"},
)

var node_uptime = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "node_uptime",
	},
)

func init() {
	prometheus.Register(node_load)
	prometheus.Register(node_uptime)
}

func collect_node_load() {
	for {
		load, _ := load.Avg()
		node_load.WithLabelValues("1m").Set(load.Load1)
		node_load.WithLabelValues("5m").Set(load.Load5)
		node_load.WithLabelValues("15m").Set(load.Load15)
	}
}

func collect_node_uptime() {
	for {
		uptime, _ := host.Uptime()
		node_uptime.Set(float64(uptime))
	}
}

func main() {

	go collect_node_load()
	go collect_node_uptime()

	router := InitRoutes()

	// Start serving the application
	fmt.Println("mini-node-exporter running on :23333")
	http.ListenAndServe(":23333", router)
}

type Load struct {
	OneMin     float64 `json:"1m"`
	FiveMin    float64 `json:"5m"`
	FifteenMin float64 `json:"15m"`
}

//InitRoutes exports the router
func InitRoutes() *chi.Mux {
	r := chi.NewRouter()
	cors := cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Origin", "Authorization", "Access-Control-Allow-Origin"},
		ExposedHeaders:   []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	})

	// middlewares
	r.Use(cors.Handler)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mini-node-exporter"))
	})

	r.Route("/info", func(r chi.Router) {
		r.Get("/hostname", func(w http.ResponseWriter, r *http.Request) {
			hostname, err := os.Hostname()
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
			}
			w.Write([]byte(hostname))
		})

		r.Get("/uptime", func(w http.ResponseWriter, r *http.Request) {
			uptime, err := host.Uptime()
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
			}
			w.Write([]byte(strconv.FormatUint(uptime, 10)))
		})

		r.Get("/load", func(w http.ResponseWriter, r *http.Request) {
			load, err := load.Avg()
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
			}
			applicationLoad := Load{
				load.Load1,
				load.Load5,
				load.Load15,
			}

			response, _ := json.Marshal(applicationLoad)
			w.Write([]byte(response))
		})
	})

	r.Handle("/metrics", promhttp.Handler())
	return r
}
