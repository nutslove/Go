// URI: oci_metrics_exporter.go
// Description: OCI Metrics Exporter for Prometheus

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
)

var (
	listen_port string
)

// type namespace struct {
// 	metricname []string
// 	query      [][]string // queryはmetricnameの数だけある. 1つのmetricnameに対して複数のqueryがある場合があるためvalueをsliceにする
// 	metrichelp []string
// 	labels     []string
// }

var namespaces = map[string]map[string][]string{
	"oci_computeagent": {
		"metricname": {
			"oci_compute_instance_cpu_utilization",
			"oci_compute_instance_memory_utilization",
			"oci_compute_instance_diskbytes_read",
			"oci_compute_instance_diskbytes_written",
			"oci_compute_instance_diskiops_read",
			"oci_compute_instance_diskiops_written",
			"oci_compute_instance_loadaverage",
			"oci_compute_instance_networkbytes_in",
			"oci_compute_instance_networkbytes_out",
		},
		"queries": {
			"CpuUtilization[1m].avg()",
			"MemoryUtilization[1m].avg()",
			"DiskBytesRead[1m].rate()",
			"DiskBytesWritten[1m].rate()",
			"DiskIopsRead[1m].rate()",
			"DiskIopsWritten[1m].rate()",
			"LoadAverage[1m].avg()",
			"NetworksBytesIn[1m].rate()",
			"NetworksBytesOut[1m].rate()",
		},
		"metrichelp": {
			"OCI Compute Instance CPU Utilization(percent)",
			"OCI Compute Instance Memory Utilization(percent)",
			"OCI Compute Instance Disk Read Throughput(bytes)",
			"OCI Compute Instance Disk Written Throughput(bytes)",
			"OCI Compute Instance Disk Read IOPS(count)",
			"OCI Compute Instance Disk Written IOPS(count)",
			"OCI Compute Instance LoadAverage(process)",
			"OCI Compute Instance Network Get Throughput(bytes)",
			"OCI Compute Instance Network Send Throughput(bytes)",
		},
		"labels": {
			"resourceId",
			"availabilityDomain",
			"region",
			"resourceDisplayName",
			"faultDomain",
		},
	},
	"oci_blockstore": {
		"metricname": {
			"oci_blockstore_volume_read_throughput",
			"oci_blockstore_volume_write_throughput",
			"oci_blockstore_volume_read_operations",
			"oci_blockstore_volume_write_operations",
			"oci_blockstore_volume_throttled_operations",
		},
		"queries": {
			"VolumeReadThroughput[1m].avg()",
			"VolumeWriteThroughput[1m].avg()",
			"VolumeReadOps[1m].avg()",
			"VolumeWriteOps[1m].avg()",
			"VolumeThrottledIOs[1m].avg()",
		},
		"metrichelp": {
			"OCI Blockstore Volume Read throughput(bytes)",
			"OCI Blockstore Volume Write throughput(bytes)",
			"OCI Blockstore Volume Read operations(count)",
			"OCI Blockstore Volume Write operations(count)",
			"OCI Blockstore Volume Throttled I/Os(count)",
		},
		"labels": {
			"attachmentId",
			"resourceId",
		},
	},
}

// OCIリソースのnamespaceごとにgaugevec変数を用意する
// リソースごとにLabel数が異なり、1つのgaugevec変数ではgaugevec.With(prometheus.Labels{}処理で以下のようなpanicが起きる
// panic: inconsistent label cardinality: expected 5 label values but got 2 in prometheus.Labels{"attachmentId":"ocid1.instance.oc1.ap-tokyo-1.anxhiljruuqvp3iccmp7d4fcjgk5x5y2yfa523pxxawch57blxpmks3pa64a", "resourceId":"ocid1.bootvolume.oc1.ap-tokyo-1.abxhiljrsbigp3d2kv7khwi47urzljbvfbjeaxoimwpfa2viqxqc7vtacy5q"}

var oci_computeagent_gaugevec = []*prometheus.GaugeVec{}
var oci_blockstore_gaugevec = []*prometheus.GaugeVec{}

func init() {
	// Metrics have to be registered to be exposed:
	for ns, v := range namespaces {
		switch ns {
		case "oci_computeagent":
			initializeGauegeVec(ns, v, &oci_computeagent_gaugevec)
		case "oci_blockstore":
			initializeGauegeVec(ns, v, &oci_blockstore_gaugevec)
		default:
			fmt.Println("You should create %v gaugevec var first.", ns)
		}
	}
}

func initializeGauegeVec(namespace string, metrics map[string][]string, gaugeVecList *[]*prometheus.GaugeVec) {
	for i, _ := range metrics["metricname"] {
		metric := GetGagueVec(metrics["metricname"][i], metrics["metrichelp"][i], metrics["labels"])
		prometheus.MustRegister(metric)
		*gaugeVecList = append(*gaugeVecList, metric)
	}
}

func GetGagueVec(metricname string, metrichelp string, labels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: metricname,
		Help: metrichelp,
	}, labels,
	)
}

func GetMetric(namespace string, query string, gaugevec *prometheus.GaugeVec) {
	// Send the request using the service client
	client, err := monitoring.NewMonitoringClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	req := monitoring.SummarizeMetricsDataRequest{
		// CompartmentIdInSubtree: common.Bool(true),
		// OpcRequestId: common.String("2FAIPMGNBUMX9TB067XB<unique_ID>"),
		SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
			// Resolution: common.String("EXAMPLE-resolution-Value"),
			// ResourceGroup: common.String("EXAMPLE-resourceGroup-Value"),
			StartTime: &common.SDKTime{Time: time.Now().Add(-time.Minute * 2)},
			EndTime:   &common.SDKTime{Time: time.Now()},
			Namespace: common.String(namespace),
			Query:     common.String(query)},
		CompartmentId: common.String("ocid1.tenancy.oc1..aaaaaaaapscp6dkyqn52z2j5u2qrz3yvhl7lw225maur75ptlollu4npaeza")}

	// fmt.Println("req: ",req)

	resp, err := client.SummarizeMetricsData(context.Background(), req)
	// fmt.Println("resp: ", resp) // メトリクスのデータを確認するために出力・確認用

	helpers.FatalIfError(err)

	for i, _ := range resp.Items {
		// fmt.Println("NAMESPACE: ",*resp.Items[i].Namespace)
		switch namespace := *resp.Items[i].Namespace; namespace {
		case "oci_computeagent":
			resourcedisplayname := resp.Items[i].Dimensions["resourceDisplayName"]
			region := resp.Items[i].Dimensions["region"]
			availabilitydomain := resp.Items[i].Dimensions["availabilityDomain"]
			faultdomain := resp.Items[i].Dimensions["faultDomain"]
			resourceid := resp.Items[i].Dimensions["resourceId"]
			metric := *resp.Items[i].AggregatedDatapoints[len(resp.Items[i].AggregatedDatapoints)-1].Value

			gaugevec.With(prometheus.Labels{
				"resourceId":          resourceid,
				"availabilityDomain":  availabilitydomain,
				"region":              region,
				"resourceDisplayName": resourcedisplayname,
				"faultDomain":         faultdomain,
			}).Set(metric)
		case "oci_blockstore":
			attachmentId := resp.Items[i].Dimensions["attachmentId"]
			resourceId := resp.Items[i].Dimensions["resourceId"]
			metric := *resp.Items[i].AggregatedDatapoints[len(resp.Items[i].AggregatedDatapoints)-1].Value

			gaugevec.With(prometheus.Labels{
				"attachmentId": attachmentId,
				"resourceId":   resourceId,
			}).Set(metric)
		}
	}
}

func GetMetrics() {
	for {
		for ns, v := range namespaces {
			// SummarizeMetricsData関数(API)に10回/secのRateLimitがあるためNameSpaceごとに1秒間sleepする
			// 1つのnamespaceのmetrics種類が10個を超えたら1つのnamespaceの中でもsleepを入れるなどしてRateLimitを超えないようにする必要がある
			time.Sleep(1 * time.Second)

			switch ns {
			case "oci_computeagent":
				for i, _ := range v["metricname"] {
					GetMetric(ns, v["queries"][i], oci_computeagent_gaugevec[i])
				}
			case "oci_blockstore":
				for i, _ := range v["metricname"] {
					GetMetric(ns, v["queries"][i], oci_blockstore_gaugevec[i])
				}
			}
		}
		time.Sleep(60 * time.Second)
	}
}

func main() {
	oci_metrics_exporter := &cli.App{
		Name:  "OCI Metrics Exporter",
		Usage: "OCI Metrics Exporter for Prometheus",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "listen-port",
				Aliases:     []string{"p"},
				Usage:       "listen port",
				Value:       "8080",
				Destination: &listen_port,
				EnvVars:     []string{"LISTEN_PORT"},
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Println("Start OCI Metrics Exporter")
			go GetMetrics()
			return nil
		},
	}

	if err := oci_metrics_exporter.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	// go GetMetrics()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+listen_port, nil))
}
