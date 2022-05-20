/**
 * @Author: derek
 * @Description:
 * @File: remoteWrite_test.go
 * @Version: 1.0.0
 * @Date: 2022/5/16 14:25
 */

package xprometheus

import (
	"fmt"
	"github.com/k0kubun/pp"
	"log"
	"testing"
	"time"
)

func TestGetGaugeMetrics(t *testing.T) {
	now := time.Now().Add(-2 * time.Minute)
	n := 9.8
	for i := 0; i < 10; i++ {
		n = n + 6.3
		RequestTime.Set(n * 10)
	}
	res := GetGaugeMetrics("RequestTime", "prtest_job", now, RequestTime)
	pp.Println(res.Label, res.Sample)
}

func TestGetCountMetrics(t *testing.T) {
	now := time.Now().Add(-2 * time.Minute)
	n := 9.8
	for i := 0; i < 10; i++ {
		n = n + 6.3
		RequestFailOrder.Add(n * 10)
	}
	res := GetCountMetrics("Requestfail_order", "prtest_job", now, RequestFailOrder)
	pp.Println(res.Label, res.Sample)
}

func TestGetHistogramMetrics(t *testing.T) {
	now := time.Now().Add(-2 * time.Minute)
	n := 9.8
	for i := 0; i < 10; i++ {
		n = n + 6.3
		CallDurationsHistogram.Observe(n * 10)
	}
	res := GetHistogramMetrics("RequestTime_histogram_seconds_wechat", "prtest_job", now, CallDurationsHistogram)
	pp.Println(res.SampleSum, res.SampleCount)
	for _, v := range res.Bucket {
		fmt.Println(v.Label, v.Sample)
	}
}

func TestRemoteWrite(t *testing.T) {
	now := time.Now().Add(-2 * time.Minute)
	n := 9.8
	for i := 0; i < 10; i++ {
		n = n + 6.3
		RequestTime.Set(n * 10)
		RequestFailOrder.Add(n * 10)
		CallDurationsHistogram.Observe(n * 10)
	}
	countRes := GetCountMetrics("Requestfail_order_total", "prtest_job", now, RequestFailOrder)
	gaugeRes := GetGaugeMetrics("RequestTime_of_gauge", "prtest_job", now, RequestTime)
	Histogrames := GetHistogramMetrics("RequestTime_histogram_seconds_wechat", "prtest_job", now, CallDurationsHistogram)
	var data []Hdata
	data = append(data, countRes, gaugeRes, Histogrames.SampleCount, Histogrames.SampleSum)
	data = append(data, Histogrames.Bucket...)
	opt := RemoteWriteOpt{
		Username: "XXXXXXXXXXX",
		Password: "XXXXXXXXXXXXXX",
		URL:      "https://localhost:9090/api/v3/write",
	}
	err := RemoteWrite(data, opt)
	if err != nil {
		log.Println("RemoteWrite(data, opt)", err)
	}
}
