package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/leetcode-golang-classroom/golang-new-relic-go-routine-sample/internal/config"
	"github.com/newrelic/go-agent/v3/newrelic"
)

var (
	randomer *rand.Rand
)

// init set initial values for variables used in the function
func init() {
	randomer = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Workshop > You may track errors using the Transaction.NoticeError method.
// The easiest way to get started with NoticeError is to use errors based on Go's standard error interface.
// https://github.com/newrelic/go-agent/blob/master/GUIDE.md#error-reporting
func noticeErrorWithAttributes(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "noticing an error!")

	txn := newrelic.FromContext(r.Context())
	txn.NoticeError(newrelic.Error{
		Message: "uh oh. something went very wrong",
		Class:   "errors are aggregated by class",
		Attributes: map[string]interface{}{
			"important_number": 97232,
			"relevant_string":  "classError",
		},
	})
	println("Oops, there is an error!")
}

func async(w http.ResponseWriter, r *http.Request) {
	// To access the transaction in your handler, use the newrelic.FromContext API.
	txn := newrelic.FromContext(r.Context())
	// This WaitGroup is used to wait for all the goroutines to finish.
	wg := &sync.WaitGroup{}
	println("goRoutines created!")

	for i := 1; i < 9; i++ {
		wg.Add(1)
		i := i

		// Workshop > trace asynchronous applications
		// The Transaction.NewGoroutine() allows transactions to create segments in multiple goroutines.
		// https://docs.newrelic.com/docs/apm/agents/go-agent/features/trace-asynchronous-applications
		go func(txn *newrelic.Transaction) {
			defer wg.Done()
			defer txn.StartSegment("goroutine" + strconv.Itoa(i)).End()
			println("goRoutine " + strconv.Itoa(i))

			randomDelay := randomer.Intn(500)
			time.Sleep(time.Duration(randomDelay) * time.Millisecond)
		}(txn.NewGoroutine())
	}

	// Workshop > Ensure the WaitGroup is done
	segment := txn.StartSegment("WaitGroup")
	wg.Wait()
	segment.End()
	w.Write([]byte("goRoutines success!"))
}

func main() {
	nrApp, nrErr := newrelic.NewApplication(
		newrelic.ConfigAppName(config.AppConfig.AppName),
		newrelic.ConfigLicense(config.AppConfig.NewRelicLicenseKey),
		// newrelic.ConfigDebugLogger(os.Stdout),
	)

	if nrErr != nil {
		log.Fatal(nrErr)
	}

	if waitErr := nrApp.WaitForConnection(5 * time.Second); waitErr != nil {
		log.Fatalf("nrApp.WaitForConnection failed %v", waitErr)
	}

	// add route for async call
	http.HandleFunc(newrelic.WrapHandleFunc(nrApp, "/error", noticeErrorWithAttributes))
	http.HandleFunc(newrelic.WrapHandleFunc(nrApp, "/async", async))
	err := http.ListenAndServe(fmt.Sprintf(":%v", config.AppConfig.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
	// Wait for shut down to ensure data gets flushed
	nrApp.Shutdown(5 * time.Second)
}
