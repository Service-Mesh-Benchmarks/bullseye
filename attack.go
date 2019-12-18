package main

import (
  "os"
  "time"
  "strconv"
  "net/http"
  vegeta "github.com/AkshatM/vegeta/lib"
  "github.com/dchest/uniuri"
)

var possible_headers = [8]string{"user-agent", "server", "x-client-trace-id", "x-envoy-downstream-service-cluster", "x-envoy-downstream-service-node", "x-envoy-external-address", "x-envoy-force-trace", "x-envoy-internal"}

// This maps a number to a list of headers. The header *values* are generated
// dynamically. The Nth header key is populated by chekcing if the Nth digit from
// the right is set to 1 in `profile`. e.g. 11111111 and 11110001 is a valid `profile`
func generateHeaders(profile string) http.Header {

	if len(profile) != 8 {
		panic("Profile must be exactly eight digits long")
	}

	var headers http.Header

	for index, header := range possible_headers {
		if string(profile[index]) == "1" {
			headers.Add(header, uniuri.New())
		}
	}
	return headers
}

// Use Vegeta to hit Envoy with desired rate, duration and headers.
// This generates latency metrics in HDR format (http://hdrhistogram.github.io/HdrHistogram/plotFiles.html)
// to stdout.
func main() {

  url := os.Args[1]
  profile := os.Args[2]

  frequency, err := strconv.Atoi(os.Args[3])
  if err != nil {
	  panic(err)
  }

  length, err := strconv.Atoi(os.Args[4])
  if err != nil {
	  panic(err)
  }

  rate := vegeta.Rate{Freq: frequency, Per: time.Second}
  duration := time.Duration(length) * time.Second
  targeter := vegeta.NewStaticTargeter(vegeta.Target{
    Method: "GET",
    URL:    url,
    BodyContainsTimestamp: true,
    Header: generateHeaders(profile),
  })
  attacker := vegeta.NewAttacker()

  var metrics vegeta.Metrics
  for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
    metrics.Add(res)
  }
  metrics.Close()

  reporter := vegeta.NewHDRHistogramPlotReporter(&metrics)
  reporter.Report(os.Stdout)
}
