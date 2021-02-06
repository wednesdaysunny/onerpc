package conf

import (
	"log"
	"net/http"
	"os"
)

func StartPprof() {
	if os.Getenv("OPEN_PPROF") == "1" {
		go func() {
			log.Println(http.ListenAndServe(":6060", nil))
		}()
	}
}
