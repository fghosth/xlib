/**
 * @Author: derek
 * @Description:
 * @File: remote_test.go
 * @Version: 1.0.0
 * @Date: 2022/5/15 06:27
 */

package xprometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"testing"
)

var (
	CallDurationsTimeBuckets = []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1500, 2000, 2500, 3000, 4000, 5000, 10000}
	labels                   = map[string]string{
		"host":    "localhost",
		"service": "Requesttest",
		"cluster": "saas-prd",
	}
	CallDurationsHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        "RequestTime_histogram_seconds_wechat",
		Help:        "RequestTime latency distributions. (ms)",
		Buckets:     CallDurationsTimeBuckets,
		ConstLabels: labels,
	}) //调用时间
	RequestFailOrder = prometheus.NewCounter(prometheus.CounterOpts{
		Name:        "Requestfail_order_total",
		Help:        "The total number of RequestFailOrder",
		ConstLabels: labels,
	})
	RequestTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "RequestTime_of_gauge",
		Help:        "RequestTime gauge data  (ms)",
		ConstLabels: labels,
	})
)

func TestMain(m *testing.M) {
	m.Run()
	os.Exit(0)
}
