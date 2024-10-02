package internal

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
)

type StatusLine struct {
	resources []Resource

	resourceSizes          []int64
	totalBytes             int64
	numBytesDownloaded     int64
	numResourcesDownloaded int
	spinI                  int
	indexCh                chan int
	startTime              time.Time
	sizingErr              error
	ctx                    context.Context
}

var spinChars = [5]string{"-", "\\", "|", "/", "-"}

// NewStatusLine creates, initializes, and starts a new StatusLine.
func NewStatusLine(resources []Resource, ctx context.Context) (*StatusLine, error) {
	st := StatusLine{}
	st.resources = resources
	st.indexCh = make(chan int)
	st.ctx = ctx
	st.getResourcesSizes(resources, 10000)
	st.start()
	return &st, st.sizingErr
}

// Increment informs the StatusLine that a resource (at index i in resource list) has finished downloading.
func (st *StatusLine) Increment(i int) {
	st.indexCh <- i
}

// start begins the goroutine and loop that will update/print the status line.
func (st *StatusLine) start() {
	st.startTime = time.Now()

	go func() {
		st.spinI = 0

		for {

			// Block until value is inserted into indexCh (>= 0 when resource finishes downloading, -1 every 50 milliseconds to keep timer and spinner updating).
			var i int
			select {
			case i = <-st.indexCh:
			case <-time.After(50 * time.Millisecond):
				i = -1
			case <-st.ctx.Done():
				return
			}
			if i != -1 {
				st.numBytesDownloaded += st.resourceSizes[i]
				st.numResourcesDownloaded++
			}

			// Update/rotate spinner.
			st.spinI += 1
			if st.spinI == len(spinChars) {
				st.spinI = 0
			}

			line := st.composeStatusString()
			fmt.Print(line)

			if st.numResourcesDownloaded == len(st.resources) {
				fmt.Println()
				return
			}

		}
	}()

}

func (st *StatusLine) getResourcesSizes(resources []Resource, timeoutMs int) {
	fmt.Print("\rFetching resource sizes...")
	st.resourceSizes = make([]int64, len(resources))
	for i := 0; i < len(st.resourceSizes); i++ {
		st.resourceSizes[i] = 0
	}

	st.totalBytes = 0
	for i, r := range resources {
		resource := r
		httpClient := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}
		resp, err := httpClient.Head(resource.Urls[0])
		if err != nil {
			st.sizingErr = err
			return
		}
		st.totalBytes += resp.ContentLength
		st.resourceSizes[i] = resp.ContentLength
	}
}

func (st *StatusLine) composeStatusString() string {
	var spinner string
	if st.numResourcesDownloaded < len(st.resources) {
		spinner = spinChars[st.spinI]
	} else {
		spinner = "✔"
	}

	barStr := "["
	barLength := 20
	if st.totalBytes < 20 {
		barLength = int(st.totalBytes)
	}
	squareSize := st.totalBytes / int64(barLength)
	for i := 0; i < int(st.numBytesDownloaded/squareSize); i += 1 {
		barStr += "█"
	}
	if st.numResourcesDownloaded < len(st.resources) {
		barStr += " "
	}
	for i := int(st.numBytesDownloaded/squareSize) + 1; i < barLength; i += 1 {
		barStr += " "
	}
	barStr += "]"

	completeStr := strconv.Itoa(st.numResourcesDownloaded) + "/" + strconv.Itoa(len(st.resources)) + " Resources"

	var byteStr string
	if st.sizingErr == nil {
		byteStr = humanize.Bytes(uint64(st.numBytesDownloaded)) + " / " + humanize.Bytes(uint64(st.totalBytes))
	} else {
		byteStr = "<issue_fetching_resource_sizes>"
	}

	elapsedStr := strconv.Itoa(int(time.Since(st.startTime).Round(time.Second).Seconds())) + "s Elapsed"

	pad := "          "
	line := "\r" + spinner + barStr + pad + completeStr + pad + byteStr + pad + elapsedStr // "\r" lets us clear the line.

	return line
}
