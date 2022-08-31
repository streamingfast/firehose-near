package nodemanager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ShinyTrinkets/overseer"
	"github.com/streamingfast/bstream"
	nodeManager "github.com/streamingfast/node-manager"
	logplugin "github.com/streamingfast/node-manager/log_plugin"
	"github.com/streamingfast/node-manager/superviser"
	"github.com/tidwall/gjson"
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
	return "near"
}

func NewSuperviser(
	binary string,
	isReader bool,
	arguments []string,
	dataDir string,
	headBlockUpdateFunc nodeManager.HeadBlockUpdater,
	debugFirehose bool,
	logToZap bool,
	appLogger *zap.Logger,
	nodelogger *zap.Logger,
) *Superviser {
	// Ensure process manager line buffer is large enough (50 MiB) for our Firehose instrumentation outputting lot's of text.
	overseer.DEFAULT_LINE_BUFFER_SIZE = 50 * 1024 * 1024

	supervisor := &Superviser{
		Superviser:          superviser.New(appLogger, binary, arguments),
		Logger:              appLogger,
		binary:              binary,
		arguments:           arguments,
		dataDir:             dataDir,
		headBlockUpdateFunc: headBlockUpdateFunc,
	}

	if isReader {
		supervisor.RegisterLogPlugin(logplugin.LogPluginFunc(supervisor.lastBlockSeenLogPlugin))
	} else {
		go supervisor.WatchLastBlock()
	}

	if logToZap {
		supervisor.RegisterLogPlugin(newToZapLogPlugin(debugFirehose, nodelogger))
	} else {
		supervisor.RegisterLogPlugin(logplugin.NewToConsoleLogPlugin(debugFirehose))
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

func (s *Superviser) sendRPCCommand(content string) ([]byte, error) {
	addr := "http://localhost:3030"

	bytesObj := []byte(content)
	reqBody := bytes.NewBuffer(bytesObj)

	resp, err := http.Post(addr, "application/json", reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (s *Superviser) getHead() (headNum uint64, headID string, headTime time.Time) {
	resp, err := s.sendRPCCommand(`{"jsonrpc":"2.0","id":"dontcare","method":"status","params":[]}`)
	if err != nil {
		return
	}
	headNum = gjson.GetBytes(resp, "result.sync_info.latest_block_height").Uint()
	headID = gjson.GetBytes(resp, "result.sync_info.latest_block_hash").String()
	headTime = gjson.GetBytes(resp, "result.sync_info.latest_block_time").Time()
	return
}

func (s *Superviser) WatchLastBlock() {
	for {
		if s.IsRunning() {
			headNum, headID, headTime := s.getHead()
			if headNum != 0 {
				blk := &bstream.Block{
					Id:        headID,
					Number:    headNum,
					Timestamp: headTime,
				}
				s.headBlockUpdateFunc(blk) // used by operator and metrics
				s.lastBlockSeen = headNum  // exported from Superviser as LastSeenBlockNum() for backups
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func (s *Superviser) lastBlockSeenLogPlugin(line string) {
	// FIRE BLOCK <HEIGHT> <HASH> <PROTO_HEX>
	if !strings.HasPrefix(line, "FIRE BLOCK") {
		return
	}

	parts := strings.SplitN(line[12:], " ", 2)
	if len(parts) != 2 {
		s.Logger.Error("invalid block line, will fail at parsing time later on", zap.String("line[0:64]", line[0:64]))
		return
	}

	blockNum, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		s.Logger.Error("unable to extract last block num", zap.String("line[0:64]", line[0:64]), zap.Error(err))
		return
	}

	s.lastBlockSeen = blockNum
}

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
