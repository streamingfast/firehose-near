package cli

import (
	"github.com/dfuse-io/dlauncher/launcher"
	"github.com/dfuse-io/logging"
	"go.uber.org/zap"
)

var userLog = launcher.UserLog
var zlog *zap.Logger

func init() {
	logging.Register("github.com/dfuse-io/dfuse-ethereum/cmd/dfuseeth", &zlog)
}
