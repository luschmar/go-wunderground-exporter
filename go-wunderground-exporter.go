package main

import (
    "fmt"
    "net/http"
    "strconv"
    "math"

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
indoortempf:
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
indoorhumidity:
  - out: indoorhumidity
    collector: Gauge
humidity:
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
winddir:
  - out: winddir
    collector: Gauge
absbaromin:
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
rainin:
  - out: rain_in
    collector: Gauge
  - out: rain_mm
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
solarradiation:
  - out: solarradiation
    collector: Gauge
UV:
  - out: uv
    collector: Gauge
dateutc:
  - out: dateutc_info
`

type Configs map[string] []OutputConfig

type OutputConfig struct {
    Out string
    Collector string
    Expression string
    Dodecimalplace bool
    Decimalplace int8
}

var configs Configs


var outCollectors map[string] prometheus.Collector = make(map[string]prometheus.Collector)

func main() {
    err := yaml.Unmarshal([]byte(data), &configs)

    if(err != nil) {
        fmt.Println(err)
    }

    http.Handle("/import", http.HandlerFunc(importData))
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":9000", nil)
}

func importData(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()

    for variable, value := range query {
        processVariableAndValue(variable, value)
    }

    w.WriteHeader(200)
    // w.Write([]byte(strings.Join(filters, ",")))
}

func processVariableAndValue(variable string, value []string) {
    c := configs[variable]
    if(c == nil) {
        fmt.Println("unknown variable ", variable, "wont process")
        return
    }

    for _, singleConfig := range c {
        processConfigWithVariableAndValue(singleConfig, variable, value)
    }
}

func processConfigWithVariableAndValue(config OutputConfig, variable string, value []string) {
    collector := getOrCreateCollector(config)
    gauge, ok := collector.(prometheus.Gauge)
    if(ok) {
        f := getConvertedValue(config.Expression, config.Dodecimalplace, config.Decimalplace, value[0])
        gauge.Set(f)
        return
    }
    counter, ok := collector.(prometheus.Counter)
    if(ok) {
        f := getConvertedValue(config.Expression, config.Dodecimalplace, config.Decimalplace, value[0])
        counter.Add(f)
        return
    }
    gaugev, ok := collector.(*prometheus.GaugeVec)
    if(ok) {
        gaugev.WithLabelValues(value[0]).Set(1)
        return
    }
    fmt.Println("Reach unprocessed ", variable)
}

func getConvertedValue(exp string, do bool, dec int8, v string) (f float64){
    f, _ = strconv.ParseFloat(v, 64)
    if(exp == "" ) {
        return roundTo(f, do, dec)
    }

    expression, _ := govaluate.NewEvaluableExpression(exp);
    parameters := make(map[string]interface{}, 8)
    parameters["in"] = f;
    result, _ := expression.Evaluate(parameters);

    return roundTo(result.(float64), do, dec)
}

func roundTo(v float64, do bool, dec int8) (f float64) {
    if(!do) {
        return v
    }
    if(dec == 0) {
        return math.Floor(v)
    }
    e := (float64)(dec*10)
    return math.Floor(v*e)/e
}

func getOrCreateCollector(config OutputConfig) (c prometheus.Collector) {
    c = outCollectors[config.Out]
    if(c == nil) {
        return initializeCollector(config)
    }

    return c
}

func initializeCollector(config OutputConfig) (c prometheus.Collector){
    if(config.Collector == "Gauge") {
        gauge :=  promauto.NewGauge(prometheus.GaugeOpts{
                Name: config.Out,
                Help: config.Out,
        })
        outCollectors[config.Out] = gauge
        return gauge
    } else if(config.Collector == "Counter") {
        counter :=  promauto.NewCounter(prometheus.CounterOpts{
                Name: config.Out,
                Help: config.Out,
        })
        outCollectors[config.Out] = counter
        return counter
    } 
    counter :=  promauto.NewGaugeVec(prometheus.GaugeOpts{
            Name: config.Out,
            Help: config.Out,
    },
    []string{config.Out})
    outCollectors[config.Out] = counter
    return counter
}
