package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gopkg.in/Knetic/govaluate.v2"
)

var data = `
ID:
  - out: id
PASSWORD:
  - out: password
PASSKEY:
  - out: password
indoortempf:
  - out: indoortemp_fahrenheit
    collector: Gauge
  - out: indoortemp_celsius
    collector: Gauge
    expression: ([in]-32)/1.8
    dodecimalplace: true
    decimalplace: 1
tempinf:
  - out: indoortemp_fahrenheit
    collector: Gauge
  - out: indoortemp_celsius
    collector: Gauge
    expression: ([in]-32)/1.8
    dodecimalplace: true
    decimalplace: 1
tempf:
  - out: temp_fahrenheit
    collector: Gauge
  - out: temp_celsius
    collector: Gauge
    expression: ([in]-32)/1.8
    dodecimalplace: true
    decimalplace: 1
dewptf:
  - out: dewpt_fahrenheit
    collector: Gauge
  - out: dewpt_celsius
    collector: Gauge
    expression: ([in]-32)/1.8
    dodecimalplace: true
    decimalplace: 1
windchillf:
  - out: windchill_fahrenheit
    collector: Gauge
  - out: windchill_celsius
    collector: Gauge
    expression: ([in]-32)/1.8
    dodecimalplace: true
    decimalplace: 1
humidity:
  - out: humidity
    collector: Gauge
humidityin:
  - out: indoorhumidity
    collector: Gauge
indoorhumidity:
  - out: indoorhumidity
    collector: Gauge
windspeedmph:
  - out: windspeed_mph
    collector: Gauge
  - out: windspeed_kmh
    collector: Gauge
    expression: ([in] * 1.609344)
    dodecimalplace: true
    decimalplace: 1
  - out: windspeed_ms
    collector: Gauge
    expression: "[in] * (1609.344 / 3600)"
    dodecimalplace: true
    decimalplace: 1
windgustmph:
  - out: windgust_mph
    collector: Gauge
  - out: windgust_kmh
    collector: Gauge
    expression: ([in] * 1.609344)
    dodecimalplace: true
    decimalplace: 1
  - out: windgust_ms
    collector: Gauge
    expression: "[in] * (1609.344 / 3600)"
    dodecimalplace: true
    decimalplace: 1
maxdailygust:
  - out: maxdailygust_mph
    collector: Gauge
  - out: maxdailygust_kmh
    collector: Gauge
    expression: ([in] * 1.609344)
    dodecimalplace: true
    decimalplace: 1
  - out: maxdailygust_ms
    collector: Gauge
    expression: "[in] * (1609.344 / 3600)"
    dodecimalplace: true
    decimalplace: 1
winddir:
  - out: winddir
    collector: Gauge
    expression: "([in] + 29)% 360"
    dodecimalplace: true
    decimalplace: 1
absbaromin:
  - out: absbarom_in
    collector: Gauge
  - out: absbarom_hpa
    collector: Gauge
    expression: "[in] * 33.863886666667"
    dodecimalplace: true
    decimalplace: 1
baromabsin:
  - out: absbarom_in
    collector: Gauge
  - out: absbarom_hpa
    collector: Gauge
    expression: "[in] * 33.863886666667"
    dodecimalplace: true
    decimalplace: 1
baromin:
  - out: barom_in
    collector: Gauge
  - out: barom_hpa
    collector: Gauge
    expression: "[in] * 33.863886666667"
    dodecimalplace: true
    decimalplace: 1
baromrelin:
  - out: barom_in
    collector: Gauge
  - out: barom_hpa
    collector: Gauge
    expression: "[in] * 33.863886666667"
    dodecimalplace: true
    decimalplace: 1
rainin:
  - out: rain_in
    collector: Gauge
  - out: rain_mm
    collector: Gauge
    expression: "[in] * 25.4"
    dodecimalplace: true
rainratein:
  - out: rain_in
    collector: Gauge
  - out: rain_mm
    collector: Gauge
    expression: "[in] * 25.4"
    dodecimalplace: true
hourlyrainin:
  - out: hourlyrain_in
    collector: Gauge
  - out: hourlyrain_mm
    collector: Gauge
    expression: "[in] * 25.4"
    dodecimalplace: true
dailyrainin:
  - out: dailyrain_in
    collector: Gauge
  - out: dailyrain_mm
    collector: Gauge
    expression: "[in] * 25.4"
    dodecimalplace: true
weeklyrainin:
  - out: weeklyrain_in
    collector: Gauge
  - out: weeklyrain_mm
    collector: Gauge
    expression: "[in] * 25.4"
    dodecimalplace: true
monthlyrainin:
  - out: monthlyrain_in
    collector: Gauge
  - out: monthlyrain_mm
    collector: Gauge
    expression: "[in] * 25.4"
    dodecimalplace: true
yearlyrainin:
  - out: yearlyrain_in
    collector: Gauge
  - out: yearlyrain_mm
    collector: Gauge
    expression: "[in] * 25.4"
    dodecimalplace: true
totalrainin:
  - out: totalrain_in
    collector: Gauge
  - out: totalrain_mm
    collector: Gauge
    expression: "[in] * 25.4"
    dodecimalplace: true
eventrainin:
  - out: eventrain_in
    collector: Gauge
  - out: eventrain_mm
    collector: Gauge
    expression: "[in] * 25.4"
    dodecimalplace: true
solarradiation:
  - out: solarradiation
    collector: Gauge
UV:
  - out: uv
    collector: Gauge
uv:
  - out: uv
    collector: Gauge
dateutc:
  - out: dateutc_info
    collector: Time
action:
  - out: action_info
softwaretype:
  - out: softwaretype_info
rtfreq:
  - out: rtfreq_info
freq:
  - out: freq_info
model:
  - out: model_info
stationtype:
  - out: stationtype_info
realtime:
  - out: realtime_info
wh65batt:
  - out: wh65batt
    collector: Gauge
`

type Configs map[string][]OutputConfig

type OutputConfig struct {
	Out            string
	Collector      string
	Expression     string
	Dodecimalplace bool
	Decimalplace   int8
}

type OutCollectorInstance struct {
	Type      string
	Collector prometheus.Collector
}

var configs Configs

var outCollectors map[string]OutCollectorInstance = make(map[string]OutCollectorInstance)

func main() {
	err := yaml.Unmarshal([]byte(data), &configs)

	if err != nil {
		fmt.Println(err)
	}

	http.Handle("/import", http.HandlerFunc(importData))
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9000", nil)
}

func importData(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if len(query) > 0 {
		for variable, value := range query {
			processVariableAndValue(variable, value)
		}
	} else {
		r.ParseForm()
		for variable, value := range r.Form {
			processVariableAndValue(variable, value)
		}
	}

	w.WriteHeader(200)
	// w.Write([]byte(strings.Join(filters, ",")))
}

func processVariableAndValue(variable string, value []string) {
	c := configs[variable]
	if c == nil {
		fmt.Println("unknown variable ", variable, " with values ", value, " wont process")
		return
	}

	for _, singleConfig := range c {
		processConfigWithVariableAndValue(singleConfig, variable, value)
	}
}

func processConfigWithVariableAndValue(config OutputConfig, variable string, value []string) {
	out := getOrCreateCollector(config)
	if out.Type == "Gauge" {
		gauge, _ := out.Collector.(prometheus.Gauge)
		f := getConvertedValue(config.Expression, config.Dodecimalplace, config.Decimalplace, value[0])
		gauge.Set(f)
		return
	} else if out.Type == "Counter" {
		counter, _ := out.Collector.(prometheus.Counter)
		f := getConvertedValue(config.Expression, config.Dodecimalplace, config.Decimalplace, value[0])
		counter.Add(f)
		return
	} else if out.Type == "Time" {
		gauge, _ := out.Collector.(prometheus.Gauge)
		f := getTimeValue(value[0])
		gauge.Set(f)
		return
	} else if out.Type == "GaugeVec" {
		gaugev, _ := out.Collector.(*prometheus.GaugeVec)
		gaugev.WithLabelValues(value[0]).Set(1)
		return
	}
	fmt.Println("Reach unprocessed ", variable)
}

func getTimeValue(v string) (f float64) {
	t, _ := time.Parse("2006-01-02 15:04:05", strings.Replace(v, "+", " ", -1))
	// from: https://github.com/prometheus/client_golang/blob/2261d5cda14eb2adc5897b56996248705f9bb840/prometheus/gauge.go#L98
	return float64(t.UnixNano()) / 1e9
}

func getConvertedValue(exp string, do bool, dec int8, v string) (f float64) {
	f, _ = strconv.ParseFloat(v, 64)
	if exp == "" {
		return roundTo(f, do, dec)
	}

	expression, _ := govaluate.NewEvaluableExpression(exp)
	parameters := make(map[string]interface{}, 8)
	parameters["in"] = f
	result, _ := expression.Evaluate(parameters)

	return roundTo(result.(float64), do, dec)
}

func roundTo(v float64, do bool, dec int8) (f float64) {
	if !do {
		return v
	}
	if dec == 0 {
		return math.Floor(v)
	}
	e := (float64)(dec * 10)
	return math.Floor(v*e) / e
}

func getOrCreateCollector(config OutputConfig) (out OutCollectorInstance) {
	v, ok := outCollectors[config.Out]
	if !ok {
		return initializeCollector(config)
	}
	return v
}

func initializeCollector(config OutputConfig) (out OutCollectorInstance) {
	if config.Collector == "Gauge" {
		gauge := promauto.NewGauge(prometheus.GaugeOpts{
			Name: config.Out,
			Help: config.Out,
		})
		outCollectors[config.Out] = OutCollectorInstance{config.Collector, gauge}
		return outCollectors[config.Out]
	} else if config.Collector == "Counter" {
		counter := promauto.NewCounter(prometheus.CounterOpts{
			Name: config.Out,
			Help: config.Out,
		})
		outCollectors[config.Out] = OutCollectorInstance{config.Collector, counter}
		return outCollectors[config.Out]
	} else if config.Collector == "Time" {
		gauge := promauto.NewGauge(prometheus.GaugeOpts{
			Name: config.Out,
			Help: config.Out,
		})
		outCollectors[config.Out] = OutCollectorInstance{config.Collector, gauge}
		return outCollectors[config.Out]
	}
	counter := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: config.Out,
		Help: config.Out,
	},
		[]string{config.Out})
	outCollectors[config.Out] = OutCollectorInstance{config.Collector, counter}
	return outCollectors[config.Out]
}
