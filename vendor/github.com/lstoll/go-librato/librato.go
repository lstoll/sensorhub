// Go client for Librato Metrics
//
// <https://github.com/rcrowley/go-librato>
package librato

type Metrics interface {
	Close()
	GetCounter(name string) chan int64
	GetCustomCounter(name string) chan map[string]int64
	GetCustomGauge(name string) chan map[string]interface{}
	GetGauge(name string) chan interface{}
	NewCounter(name string) chan int64
	NewCustomCounter(name string) chan map[string]int64
	NewCustomGauge(name string) chan map[string]interface{}
	NewGauge(name string) chan interface{}
	Wait()
}

func handle(i interface{}, bodyMetric tmetric) bool {
	var intobj map[string]int64
	var ifobj map[string]interface{}
	var ok bool
	switch ch := i.(type) {
	case chan interface{}:
		bodyMetric["value"], ok = <-ch
	case chan int64:
		bodyMetric["value"], ok = <-ch
	case chan map[string]interface{}:
		ifobj, ok = <-ch
		for k, v := range ifobj {
			bodyMetric[k] = v
		}
	case chan map[string]int64:
		intobj, ok = <-ch
		for k, v := range intobj {
			bodyMetric[k] = v
		}
	}

	return ok
}

// models http://dev.librato.com/v1/post/metrics (3) Array format (JSON only)
type tbody map[string]tibody
type tibody []tmetric
type tmetric map[string]interface{}
