package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/wblakecaldwell/profiler"
)

var (
	extraInfo map[string]interface{}
	mu        = &sync.RWMutex{}
	running   bool
)

func RunProfiler(binding string, port string) {
	extraInfo = make(map[string]interface{})
	profiler.AddMemoryProfilingHandlers()
	profiler.RegisterExtraServiceInfoRetriever(extraServiceInfo)

	listen := fmt.Sprintf("%s:%s", binding, port)
	log.Infof("Profiler is running on %s:%s/profiler/info.html", binding, port)
	http.ListenAndServe(listen, nil)
}

func UpdateExtra(key string, value interface{}) {
	if running {
		mu.Lock()
		extraInfo[key] = value
		mu.Unlock()
	}
}

func Int(key string) int {
	if value, ok := extraServiceInfo()[key].(int); ok {
		return value
	}
	return 0
}

func extraServiceInfo() map[string]interface{} {
	mu.RLock()
	defer mu.RUnlock()

	copy := map[string]interface{}{}
	for k, v := range extraInfo {
		copy[k] = v
	}
	return copy
}
