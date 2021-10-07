package codec

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	pbcodec "github.com/streamingfast/sf-near/pb/sf/near/codec/v1"
	"go.uber.org/zap"
)

// ConsoleReader is what reads the `geth` output directly. It builds
// up some LogEntry objects. See `LogReader to read those entries .
type ConsoleReader struct {
	lines chan string
	close func()

	ctx  *parseCtx
	done chan interface{}
}

func NewConsoleReader(lines chan string) (*ConsoleReader, error) {
	l := &ConsoleReader{
		lines: lines,
		close: func() {},
		ctx:   &parseCtx{},
		done:  make(chan interface{}),
	}
	return l, nil
}

//todo: WTF?
func (l *ConsoleReader) Done() <-chan interface{} {
	return l.done
}

func (c *ConsoleReader) Close() {
	c.close()
}

type parsingStats struct {
	startAt  time.Time
	blockNum uint64
	data     map[string]int
}

func newParsingStats(block uint64) *parsingStats {
	return &parsingStats{
		startAt:  time.Now(),
		blockNum: block,
		data:     map[string]int{},
	}
}

func (s *parsingStats) log() {
	zlog.Info("mindreader block stats",
		zap.Uint64("block_num", s.blockNum),
		zap.Int64("duration", int64(time.Since(s.startAt))),
		zap.Reflect("stats", s.data),
	)
}

func (s *parsingStats) inc(key string) {
	if s == nil {
		return
	}
	k := strings.ToLower(key)
	value := s.data[k]
	value++
	s.data[k] = value
}

type parseCtx struct {
	stats *parsingStats
}

func (c *ConsoleReader) Read() (out interface{}, err error) {
	return c.next(readBlock)
}

const (
	readBlock       = 1
	readTransaction = 2
)

func (c *ConsoleReader) next(readType int) (out interface{}, err error) {
	ctx := c.ctx

	zlog.Debug("next", zap.Int("read_type", readType))

	for line := range c.lines {
		if !strings.HasPrefix(line, "DMLOG ") {
			continue
		}

		line = line[6:]

		switch {
		case strings.HasPrefix(line, "BLOCK"):
			out, err = ctx.readBlock(line)
		default:
			if traceEnabled {
				zlog.Debug("skipping unknown deep mind log line", zap.String("line", line))
			}

			continue
		}

		if err != nil {
			chunks := strings.SplitN(line, " ", 2)
			return nil, fmt.Errorf("%s: %s (line %q)", chunks[0], err, line)
		}

		if out != nil {
			return out, nil
		}
	}

	zlog.Info("lines channel has been closed")
	return nil, io.EOF
}

func (c *ConsoleReader) ProcessData(reader io.Reader) error {
	scanner := c.buildScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		c.lines <- line
	}

	if scanner.Err() == nil {
		close(c.lines)
		return io.EOF
	}

	return scanner.Err()
}

func (c *ConsoleReader) buildScanner(reader io.Reader) *bufio.Scanner {
	buf := make([]byte, 50*1024*1024)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(buf, 50*1024*1024)

	return scanner
}

// Formats
// DMLOG BLOCK <NUM> <HASH> <PROTO_HEX>
func (ctx *parseCtx) readBlock(line string) (*pbcodec.Block, error) {
	chunks, err := SplitInChunks(line, 4)
	if err != nil {
		return nil, fmt.Errorf("split: %s", err)
	}

	blockNum, err := strconv.ParseUint(chunks[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid block num: %w", err)
	}

	// We skip block hash for now

	protoBytes, err := hex.DecodeString(chunks[2])
	if err != nil {
		return nil, fmt.Errorf("invalid block bytes: %w", err)
	}

	block := &pbcodec.Block{}
	if err := proto.Unmarshal(protoBytes, block); err != nil {
		return nil, fmt.Errorf("invalid block: %w", err)
	}

	newParsingStats(blockNum).log()

	return block, err
}

// splitInChunks split the line in `count` chunks and returns the slice `chunks[1:count]` (so exclusive end), but verifies
// that there are only exactly `count` chunks, and nothing more.
func SplitInChunks(line string, count int) ([]string, error) {
	chunks := strings.SplitN(line, " ", -1)
	if len(chunks) != count {
		return nil, fmt.Errorf("%d fields required but found %d fields for line %q", count, len(chunks), line)
	}

	return chunks[1:count], nil
}

// splitInBoundedChunks split the line in `count` chunks and returns the slice `chunks[1:count]` (so exclusive end),
// but will accumulate all trailing chunks within the last (for free-form strings, or JSON objects)
func SplitInBoundedChunks(line string, count int) ([]string, error) {
	chunks := strings.SplitN(line, " ", count)
	if len(chunks) != count {
		return nil, fmt.Errorf("%d fields required but found %d fields for line %q", count, len(chunks), line)
	}

	return chunks[1:count], nil
}
