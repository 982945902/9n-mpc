/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"math"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// DefaultRetry is the recommended retry for a conflict where multiple clients
// are making changes to the same resource.
var AlwaysRetry = wait.Backoff{
	Steps:    math.MaxInt,
	Duration: 5 * time.Second,
	Factor:   1.0,
	Jitter:   0.0,
}

// OnError allows the caller to retry fn in case the error returned by fn is retriable
// according to the provided function. backoff defines the maximum retries and the wait
// interval between two retries.
func OnError(backoff wait.Backoff, retriable func(error) bool, fn func() error) error {
	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := fn()
		switch {
		case err == nil:
			return true, nil
		case retriable(err):
			lastErr = err
			return false, nil
		default:
			return false, err
		}
	})
	if err == wait.ErrWaitTimeout {
		err = lastErr
	}
	return err
}

// RetryOnConflict is used to make an update to a resource when you have to worry about
// conflicts caused by other code making unrelated updates to the resource at the same
// time. fn should fetch the resource to be modified, make appropriate changes to it, try
// to update it, and return (unmodified) the error from the update function. On a
// successful update, RetryOnConflict will return nil. If the update function returns a
// "Conflict" error, RetryOnConflict will wait some amount of time as described by
// backoff, and then try again. On a non-"Conflict" error, or if it retries too many times
// and gives up, RetryOnConflict will return an error to the caller.
//
//	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
//	    // Fetch the resource here; you need to refetch it on every try, since
//	    // if you got a conflict on the last update attempt then you need to get
//	    // the current version before making your own changes.
//	    pod, err := c.Pods("mynamespace").Get(name, metav1.GetOptions{})
//	    if err != nil {
//	        return err
//	    }
//
//	    // Make whatever updates to the resource are needed
//	    pod.Status.Phase = v1.PodFailed
//
//	    // Try to update
//	    _, err = c.Pods("mynamespace").UpdateStatus(pod)
//	    // You have to return err itself here (not wrapped inside another error)
//	    // so that RetryOnConflict can identify it correctly.
//	    return err
//	})
//	if err != nil {
//	    // May be conflict if max retries were hit, or may be something unrelated
//	    // like permissions or a network error
//	    return err
//	}
//	...
//
// TODO: Make Backoff an interface?
func RetryOnConflict(backoff wait.Backoff, fn func() error, ifretry func(err error) bool) error {
	return OnError(backoff, ifretry, fn)
}
