package utils

import (
	"fmt"
	"time"
)

func BackoffRetryMechanism[T any](numOfRetry int, fn func() (T, error)) (T, error) {
	if numOfRetry < 1 {
		numOfRetry = 5
	}

	var (
		res T
		err error
	)

	for i := range numOfRetry {
		res, err = fn()
		if err != nil {
			time.Sleep(time.Duration(i<<1) * time.Second)
			continue
		}
		return res, nil
	}

	return res, fmt.Errorf("maximum number of retries: %v", err)
}
