package eotel

import (
	"github.com/getsentry/sentry-go"
	"time"
)

func CaptureError(err error, tags map[string]string, extras map[string]interface{}) {
	if err == nil || !globalCfg.EnableSentry {
		return
	}
	sentry.WithScope(func(scope *sentry.Scope) {
		for k, v := range tags {
			scope.SetTag(k, v)
		}
		for k, v := range extras {
			scope.SetExtra(k, v)
		}
		sentry.CaptureException(err)
	})
	sentry.Flush(2 * time.Second)
}
