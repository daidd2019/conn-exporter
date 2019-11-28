package collector

import (
	"sync"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"os/exec"
	"fmt"
	"strings"
)

type Metrics struct {
	metrics map[string]*prometheus.Desc
	apps    []*PortData
	mutex   sync.Mutex
}

type PortData struct {
	app string
	port string
}
func NewPortData(app, port string) *PortData {
	return &PortData{app, port}
}

type MetricsData struct {
	app string
	host string
	value float64
}
func NewMetricsData(app, host string , value float64) *MetricsData{
	return &MetricsData{app, host, value}
}


func newGlobalMetric(namespace string, metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(namespace+"_"+metricName, docString, labels, nil)
}


func NewMetrics(namespace string, ports []*PortData) *Metrics {

	metricsdef := make(map[string]*prometheus.Desc)

	for _, v := range ports {
		metricsdef[v.app] = newGlobalMetric(namespace, v.app + "_connections_metric", v.port + " connections", []string{"host"})
	}
	return &Metrics{
		metrics:  metricsdef,
		apps: ports ,
		}
}

func (c *Metrics) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m
	}
}

func (c *Metrics) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()  // 加锁
	defer c.mutex.Unlock()

	for _, v := range c.GetConnectionsData() {
		ch <-prometheus.MustNewConstMetric(c.metrics[v.app], prometheus.GaugeValue, v.value, v.host)
    }


}

func (c *Metrics) GetConnectionsData() []*MetricsData {
	result := ExecCommand("netstat -nat | grep ESTABLISHED")

	var data []*MetricsData
	
	for _, v := range strings.Split(result, "\n") {
		if strings.Contains(v, "ESTABLISHED") {
			lines := strings.Fields(v)
			for _, vd := range c.apps {
				if vd.port == lines[3] {
					delIndex := strings.Index(lines[4], ":")
					ip := lines[4][:delIndex]
					flag := true
					for _, vvd := range data {
						if vvd.app == vd.app && vvd.host == ip {
							vvd.value = vvd.value + 1
							flag = false
						}
					}
					if flag {
						data = append(data, NewMetricsData(vd.app, ip, 1))
					}
				}
			}
		}
	}

	return data 
}


func ExecCommand(strCommand string) string {
	cmd := exec.Command("/bin/bash", "-c", strCommand)
    stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		fmt.Println("Execute failed when Start:" + err.Error())
		return ""
	}

	out_bytes, _ := ioutil.ReadAll(stdout)
	stdout.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Println("Execute failed when Wait:" + err.Error())
		return ""
	}
	return string(out_bytes)
}

