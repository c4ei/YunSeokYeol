package standalone

import (
	"github.com/c4ei/c4exd/infrastructure/logger"
	"github.com/c4ei/c4exd/util/panics"
)

var log = logger.RegisterSubSystem("NTAR")
var spawn = panics.GoroutineWrapperFunc(log)
