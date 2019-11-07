// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package locks

import (
	"github.com/daglabs/btcd/logger"
	"github.com/daglabs/btcd/util/panics"
)

var (
	log, _ = logger.Get(logger.SubsystemTags.UTIL)
	spawn  = panics.GoroutineWrapperFunc(log, logger.BackendLog)
)