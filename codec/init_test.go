package codec

import (
	firenear "github.com/streamingfast/firehose-near"
	"github.com/streamingfast/logging"
)

func init() {
	logging.InstantiateLoggers()
	firenear.TestingInitBstream()
}
