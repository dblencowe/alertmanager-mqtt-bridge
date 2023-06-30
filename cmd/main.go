package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/alertmanager/notify/webhook"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type options struct {
	HttpBindAddress string `short:"b" long:"bind" description:"Address to bind the HTTP control server to" default:"localhost:8031"`
}

type mqttMessage struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Status      string            `json:"status"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

var opts options
var mqttUrl *url.URL
var mqttClient mqtt.Client

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	mqttUrlStr, ok := os.LookupEnv("MQTT_URL")
	if !ok {
		log.Fatal("Error: Required MQTT_URL not present")
	}
	mqttUrl, err = url.Parse(mqttUrlStr)
	if err != nil {
		log.Fatal(err)
	}

	mqttOpts := mqtt.NewClientOptions().AddBroker(mqttUrl.String()).SetClientID(getProgramName()).SetUsername(mqttUrl.User.Username())
	password, isSet := mqttUrl.User.Password()
	if isSet {
		mqttOpts.SetPassword(password)
	}

	mqtt.ERROR = log.New(os.Stderr, "", 0)
	mqttClient = mqtt.NewClient(mqttOpts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error: Could not connect to MQTT: %s", token.Error())
	}

	http.HandleFunc("/", httpHandler)

	log.Fatal(http.ListenAndServe(opts.HttpBindAddress, nil))
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	var alert webhook.Message
	err := json.NewDecoder(r.Body).Decode(&alert)
	if err != nil {
		log.Printf("Could not decode alert: %v", err)
		http.Error(w, "Could not decode alert", http.StatusInternalServerError)
		return
	}

	for _, a := range alert.Alerts {
		msg := mqttMessage{
			Name:        a.Labels["alertname"],
			URL:         a.GeneratorURL,
			Status:      a.Status,
			StartsAt:    a.StartsAt,
			EndsAt:      a.EndsAt,
			Labels:      a.Labels,
			Annotations: a.Annotations,
		}
		msgJson, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error: marshalling the MQTT message failed: %s\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if token := mqttClient.Publish(mqttUrl.Path, 0, false, msgJson); token.Wait() && token.Error() != nil {
			log.Printf("Error: publishing the MQTT message failed: %s\n", token.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, http.StatusText(http.StatusCreated))
}

func getProgramName() string {
	path, err := os.Executable()

	if err != nil {
		log.Println("Warning: Could not determine program name; using 'unknown'.")
		return "unknown"
	}

	return filepath.Base(path)
}
