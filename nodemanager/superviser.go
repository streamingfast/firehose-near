package nodemanager

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/ShinyTrinkets/overseer"
	nodeManager "github.com/dfuse-io/node-manager"
	logplugin "github.com/dfuse-io/node-manager/log_plugin"
	"github.com/dfuse-io/node-manager/metrics"
	"github.com/dfuse-io/node-manager/superviser"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Superviser struct {
	*superviser.Superviser

	//backupMutex         sync.Mutex
	infoMutex           sync.Mutex
	binary              string
	arguments           []string
	dataDir             string
	lastBlockSeen       uint64
	serverId            string
	headBlockUpdateFunc nodeManager.HeadBlockUpdater
	Logger              *zap.Logger
}

func (s *Superviser) GetName() string {
	return "neard"
}

func NewSuperviser(
	binary string,
	arguments []string,
	dataDir string,
	headBlockUpdateFunc nodeManager.HeadBlockUpdater,
	debugDeepMind bool,
	logToZap bool,
	appLogger *zap.Logger,
	nodelogger *zap.Logger,
) *Superviser {
	// Ensure process manager line buffer is large enough (50 MiB) for our Deep Mind instrumentation outputting lot's of text.
	overseer.DEFAULT_LINE_BUFFER_SIZE = 50 * 1024 * 1024

	supervisor := &Superviser{
		Superviser:          superviser.New(appLogger, binary, arguments),
		Logger:              appLogger,
		binary:              binary,
		arguments:           arguments,
		dataDir:             dataDir,
		headBlockUpdateFunc: headBlockUpdateFunc,
	}

	supervisor.RegisterLogPlugin(logplugin.LogPluginFunc(supervisor.lastBlockSeenLogPlugin))

	if logToZap {
		supervisor.RegisterLogPlugin(newToZapLogPlugin(debugDeepMind, nodelogger))
	} else {
		supervisor.RegisterLogPlugin(logplugin.NewToConsoleLogPlugin(debugDeepMind))
	}

	appLogger.Info("created near superviser", zap.Object("superviser", supervisor))
	return supervisor
}

func (s *Superviser) setServerId(serverId string) error {
	ipAddr := getIPAddress()
	if ipAddr == "" {
		return fmt.Errorf("cannot find local IP address")
	}

	s.infoMutex.Lock()
	defer s.infoMutex.Unlock()
	s.serverId = fmt.Sprintf(`${1}@%s:30303`, ipAddr)
	return nil
}

func (s *Superviser) GetCommand() string {
	return s.binary + " " + strings.Join(s.arguments, " ")
}

func (s *Superviser) IsRunning() bool {
	isRunning := s.Superviser.IsRunning()
	isRunningMetricsValue := float64(0)
	if isRunning {
		isRunningMetricsValue = float64(1)
	}

	metrics.NodeosCurrentStatus.SetFloat64(isRunningMetricsValue)

	return isRunning
}

func (s *Superviser) LastSeenBlockNum() uint64 {
	return s.lastBlockSeen
}

func (s *Superviser) ServerID() (string, error) {
	return s.serverId, nil
}

func (s *Superviser) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("binary", s.binary)
	enc.AddArray("arguments", stringArray(s.arguments))
	enc.AddString("data_dir", s.dataDir)
	enc.AddUint64("last_block_seen", s.lastBlockSeen)
	enc.AddString("server_id", s.serverId)

	return nil
}

func (s *Superviser) lastBlockSeenLogPlugin(line string) {
	if !strings.HasPrefix(line, "DMLOG FINALIZE_BLOCK") {
		return
	}

	line = strings.TrimSpace(strings.TrimPrefix(line, "DMLOG FINALIZE_BLOCK"))

	blockNum, err := strconv.ParseUint(line, 10, 64)
	if err != nil {
		s.Logger.Error("unable to extract last block num", zap.String("line", line), zap.Error(err))
		return
	}

	//metrics.SetHeadBlockNumber(blockNum)
	s.lastBlockSeen = blockNum
}

// AddPeer sends a command through IPC socket to connect geth to the given peer

//func (s *Superviser) sendGethCommand(cmd string) (string, error) {
//	c, err := net.Dial("unix", s.ipcFilePath)
//	if err != nil {
//		return "", err
//	}
//	defer c.Close()
//
//	_, err = c.Write([]byte(cmd))
//	if err != nil {
//		return "", err
//	}
//
//	resp, err := readString(c)
//	return resp, err
//}

func getIPAddress() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.IsGlobalUnicast() {
				return ip.String()
			}
		}
	}
	return ""
}
