package internal

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
)

type StatusLine struct {
	resources              *[]Resource
	resourceSizes          []int64
	totalBytes             int64
	numBytesDownloaded     int64
	numResourcesDownloaded int
	spinI                  int
	isRunning              atomic.Bool
	startTime              time.Time
	sizingErr              error
	ctx                    context.Context
	mtx                    sync.RWMutex
}

const spinChars = "-\\|/"
const tickMs = 50
const timeoutMs = 1000

// NewStatusLine creates and initializes a new StatusLine.
func NewStatusLine(ctx context.Context, resources *[]Resource) *StatusLine {
	st := StatusLine{}
	st.resources = resources
	st.ctx = ctx
	return &st
}

// Increment informs the StatusLine that a resource (at index i in resource list) has finished downloading.
// The SL is printed and Stopped if all resources are downloaded.
func (st *StatusLine) Increment(i int) {
	st.mtx.Lock()
	st.numBytesDownloaded += st.resourceSizes[i]
	st.numResourcesDownloaded++
	st.mtx.Unlock()

	fmt.Print(st.GetStatusString() + " ")
	if st.numResourcesDownloaded == len(*st.resources) {
		fmt.Println()
		st.Stop()
		return
	}
}

// Start begins the goroutine and loop that will update/print the status line.
// Pass true to force SL to update (spinner and second counter) every 50ms.
func (st *StatusLine) Start(doTick bool) {
	st.isRunning.Store(true)
	st.startTime = time.Now()
	go func() {
		fmt.Print(st.GetStatusString())
		for {
			select {
			case <-st.ctx.Done():
				st.Stop()
				fmt.Println()
				return
			case <-time.After(tickMs * time.Millisecond):
				if !doTick {
					continue
				}
			}

			st.mtx.Lock()
			st.spinI = (st.spinI + 1) % len(spinChars)
			st.mtx.Unlock()

			fmt.Print(st.GetStatusString() + " ")
		}
	}()

}

func (st *StatusLine) Stop() {
	st.isRunning.Store(false)
}

// initResourceSizes fetches the size, in bytes, of each resource.
func (st *StatusLine) InitResourcesSizes() error {
	log.Debug().Msg("Fetching resource sizes")
	st.resourceSizes = make([]int64, len(*st.resources))
	for i := 0; i < len(st.resourceSizes); i++ {
		st.resourceSizes[i] = 0
	}

	st.totalBytes = 0
	for i, r := range *st.resources {
		resource := r
		headers := http.Header{}
		err := requests.
			URL(resource.Urls[0]).
			Transport(noCompressionTransport).
			Head().
			CopyHeaders(headers).
			CheckStatus(http.StatusOK).
			Fetch(context.Background())
		if err != nil {
			log.Debug().Msg("Error fetching resource sizes")
			return err
		}
		ContentLength, err := strconv.Atoi(headers.Get("Content-Length"))
		if err != nil {
			log.Debug().Msg("Error fetching resource sizes")
			return err
		}
		st.totalBytes += int64(ContentLength)
		st.resourceSizes[i] = int64(ContentLength)
	}

	return nil
}

// GetStatusString composes and returns the status line string for printing.
func (st *StatusLine) GetStatusString() string {
	st.mtx.RLock()
	defer st.mtx.RUnlock()

	var spinner string
	if st.numResourcesDownloaded < len(*st.resources) {
		spinner = string(spinChars[st.spinI])
	} else {
		spinner = "✔"
	}

	barStr := "["
	if st.sizingErr == nil && st.totalBytes > 0 {
		barLength := 20
		if st.totalBytes < 20 {
			barLength = int(st.totalBytes)
		}
		squareSize := st.totalBytes / int64(barLength)
		for i := 0; i < int(st.numBytesDownloaded/squareSize); i += 1 {
			barStr += "█"
		}
		for i := int(st.numBytesDownloaded / squareSize); i < barLength; i += 1 {
			barStr += " "
		}
	}
	barStr += "]"

	completeStr := strconv.Itoa(st.numResourcesDownloaded) + "/" + strconv.Itoa(len(*st.resources)) + " Resources"
	byteStr := humanize.Bytes(uint64(st.numBytesDownloaded)) + " / " + humanize.Bytes(uint64(st.totalBytes))
	elapsedStr := strconv.Itoa(int(time.Since(st.startTime).Round(time.Second).Seconds())) + "s elapsed"

	pad := "          "
	var line string
	if st.sizingErr == nil {
		line = "\r" + spinner + barStr + pad + completeStr + pad + byteStr + pad + elapsedStr // "\r" lets us clear the line.
	} else {
		line = "\r" + spinner + "[]" + pad + completeStr + pad + elapsedStr
	}
	return line
}
