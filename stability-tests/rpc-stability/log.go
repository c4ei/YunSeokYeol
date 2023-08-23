package main

import (
	"github.com/c4ei/c4exd/infrastructure/logger"
	"github.com/c4ei/c4exd/util/panics"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("JSTT")
	spawn      = panics.GoroutineWrapperFunc(log)
)
