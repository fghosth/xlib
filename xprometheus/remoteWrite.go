/**
 * @Author: derek
 * @Description:
 * @File: remoteWrite.go
 * @Version: 1.0.0
 * @Date: 2022/5/16 13:59
 */

package xprometheus

import (
	"context"
	"encoding/base64"
	"github.com/castai/promwrite"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"log"
	"strconv"
	"time"
)

var (
	CategoryCount = "count"
	CategoryGauge = "gauge"
)

type HMetric struct {
	Bucket      []Hdata
	SampleSum   Hdata
	SampleCount Hdata
}
type Hdata struct {
	Name   string
	Label  []promwrite.Label
	Sample promwrite.Sample
}

type RemoteWriteOpt struct {
	Username string
	Password string
	URL      string
}

// RemoteWrite
// @Description: 远程写如prometheus
// @param data
// @param opt
// @return err
func RemoteWrite(data []Hdata, opt RemoteWriteOpt) (err error) {
	auth := []byte(opt.Username + ":" + opt.Password)
	basicAuth := base64.StdEncoding.EncodeToString(auth)
	header := map[string]string{"Authorization": "Basic " + basicAuth}
	client := promwrite.NewClient(opt.URL)
	var pdata []promwrite.TimeSeries
	// metrics
	for _, v := range data {
		pdata = append(pdata, promwrite.TimeSeries{
			Labels: v.Label,
			Sample: v.Sample,
		})
	}
	resp, err := client.Write(context.Background(), &promwrite.WriteRequest{
		TimeSeries: pdata,
	}, promwrite.WriteHeaders(header))
	log.Println(resp)
	return
}

// getCountAndGaugeMetrics
// @Description: 得到metric count 和gauge 类型
// @param name
// @param job
// @param t
// @param dto
// @param category string "count" and "gauge"
// @return metics
func getCountAndGaugeMetrics(name string, job string, t time.Time, w *dto.Metric, category string) (metics Hdata) {
	var labelMap []promwrite.Label
	var val float64
	var title string
	switch category {
	case CategoryCount:
		val = w.Counter.GetValue()
		title = name + "_total"
	case CategoryGauge:
		val = w.Gauge.GetValue()
		title = name + "_Gauge"
	}
	//job and name label
	labelMap = append(labelMap, []promwrite.Label{
		{
			Name:  "job",
			Value: job,
		},
		{
			Name:  "__name__",
			Value: title,
		},
	}...)
	for _, v := range w.Label {
		labelMap = append(labelMap, promwrite.Label{
			Name:  v.GetName(),
			Value: v.GetValue(),
		})
	}

	// metrics
	metics = Hdata{
		Name:  name,
		Label: labelMap,
		Sample: promwrite.Sample{
			Time:  t,
			Value: val,
		},
	}
	return
}

// GetGaugeMetrics
// @Description: 获取Count remote write 数据
// @param name metric name
// @param job job name
// @param t time
// @param histogram 数据
// @return metics
func GetGaugeMetrics(name string, job string, t time.Time, counter prometheus.Gauge) (metics Hdata) {
	w := &dto.Metric{}
	err := counter.Write(w)
	if err != nil {
		log.Println(err)
	}
	metics = getCountAndGaugeMetrics(name, job, t, w, CategoryGauge)
	return
}

// GetCountMetrics
// @Description: 获取Count remote write 数据
// @param name metric name
// @param job job name
// @param t time
// @param histogram 数据
// @return metics
func GetCountMetrics(name string, job string, t time.Time, counter prometheus.Counter) (metics Hdata) {
	w := &dto.Metric{}
	err := counter.Write(w)
	if err != nil {
		log.Println(err)
	}
	metics = getCountAndGaugeMetrics(name, job, t, w, CategoryCount)
	return
}

// GetHistogramMetrics
// @Description: 获取Histogram remote write 数据
// @param name metric name
// @param job job name
// @param t time
// @param histogram 数据
// @return metics
func GetHistogramMetrics(name string, job string, t time.Time, histogram prometheus.Histogram) (metics HMetric) {
	w := &dto.Metric{}
	err := histogram.Write(w)
	if err != nil {
		log.Println(err)
	}
	var labelMap []promwrite.Label
	//job label
	labelMap = append(labelMap, promwrite.Label{
		Name:  "job",
		Value: job,
	})
	for _, v := range w.Label {
		labelMap = append(labelMap, promwrite.Label{
			Name:  v.GetName(),
			Value: v.GetValue(),
		})
	}
	//now := time.Now().Add(-2 * time.Minute)
	// sum metric
	metics.SampleSum = Hdata{
		Name: name + "_sum",
		Label: append([]promwrite.Label{
			{
				Name:  "__name__",
				Value: name + "_sum",
			},
		}, labelMap...),
		Sample: promwrite.Sample{
			Time:  t,
			Value: w.GetHistogram().GetSampleSum(),
		},
	}
	// count metric
	metics.SampleCount = Hdata{
		Name: name + "_count",
		Label: append([]promwrite.Label{
			{
				Name:  "__name__",
				Value: name + "_count",
			},
		}, labelMap...),
		Sample: promwrite.Sample{
			Time:  t,
			Value: float64(w.GetHistogram().GetSampleCount()),
		},
	}
	// bucket metrics
	for _, v := range w.Histogram.Bucket {
		l := []promwrite.Label{
			{
				Name:  "le",
				Value: strconv.FormatFloat(*v.UpperBound, 'f', 2, 64),
			},
			{
				Name:  "__name__",
				Value: name + "_bucket",
			},
		}
		hdata := Hdata{
			Name:  name + "_bucket",
			Label: append(l, labelMap...),
			Sample: promwrite.Sample{
				Time:  t,
				Value: float64(v.GetCumulativeCount()),
			},
		}
		//pp.Println(hdata)
		metics.Bucket = append(metics.Bucket, hdata)
	}
	return
}
