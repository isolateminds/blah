package containers

import (
	"context"
	"strings"
	"sync"

	"github.com/dannyvidal/blah/internal/color"
)

//Ensures that the context is canceled if its not canceled yet
func ensureCTXCanceled(ctx context.Context, cancel context.CancelFunc) {
	if ctx.Err() == nil {
		cancel()
	}
}

//Ensures that
//       wg.Done() //gets called.
//Fatally exits if an error is not nil because that means it was not handled properly in the callback
//or it can indicate something went wrong within the callback
func exit(wg *sync.WaitGroup, err error) int {
	if err != nil {
		color.PrintFatal(err)
	}
	wg.Done()
	return 0
}

//parses the NeedContainerRemoveError and returns the container ID
func GetIDFromNeedContainerRemoveError(err error) string {
	if !IsErrNeedContainerRemove(err) {
		color.PrintFatal(err)
	}
	return strings.Split(err.Error(), "\"")[3]
}
