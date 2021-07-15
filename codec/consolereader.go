package codec

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	pbcodec "github.com/streamingfast/near-sf/pb/sf/near/codec/v1"
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
	currentBlock *pbcodec.Block
	currentTrace *pbcodec.TransactionTrace

	transactionTraces   []*pbcodec.TransactionTrace
	evmCallStackIndexes []int32

	blockStoreURL string
	stats         *parsingStats
}

func (c *ConsoleReader) Read() (out interface{}, err error) {
	return c.next(readBlock)
}

func (c ConsoleReader) ReadTransaction() (trace *pbcodec.TransactionTrace, err error) {
	out, err := c.next(readTransaction)
	if err != nil {
		return nil, err
	}

	return out.(*pbcodec.TransactionTrace), nil
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

		// Order conditions based (approximately) on those that appear more often
		switch {
		case strings.HasPrefix(line, "START_BLOCK"):
			err = ctx.readStartBlock(line)

		case strings.HasPrefix(line, "START_TRANSACTION"):
			err = ctx.readStartTransaction(line)

		case strings.HasPrefix(line, "END_TRANSACTION"):
			err = ctx.readEndTransaction(line)

		case strings.HasPrefix(line, "END_BLOCK"):
			return ctx.readEndBlock(line)
		case strings.HasPrefix(line, "APPLY_CHUNKS"):
		case strings.HasPrefix(line, "BEFORE_APPLY_CHUNKS"):
		case strings.HasPrefix(line, "CREATE_RECEIPT"):
		case strings.HasPrefix(line, "COMPLETED_LOCAL_RECEIPT"):
		case strings.HasPrefix(line, "COMPLETED_DELAYED_RECEIPT"):
		case strings.HasPrefix(line, "COMPLETED_SHARDED_RECEIPT"):
		default:
			return nil, fmt.Errorf("unsupported log line: %q", line)
		}

		if err != nil {
			chunks := strings.SplitN(line, " ", 2)
			return nil, fmt.Errorf("%s: %s (line %q)", chunks[0], err, line)
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
// DMLOG START_BLOCK <NUM> `<HASH>`
func (ctx *parseCtx) readStartBlock(line string) error {
	chunks, err := SplitInChunks(line, 3)
	if err != nil {
		return fmt.Errorf("split: %s", err)
	}

	blockNum, err := strconv.ParseUint(chunks[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid block num: %s", err)
	}

	if ctx.currentTrace != nil {
		return fmt.Errorf("found leftover transactionTrace when starting block %d", blockNum)
	}

	ctx.stats = newParsingStats(blockNum)
	ctx.currentBlock = &pbcodec.Block{
		Ver:    1,
		Hash:   []byte(chunks[1]),
		Number: blockNum,
		Size:   1, // SHOULD REMOVE THIS
	}

	return nil
}

// Formats
// DMLOG START_TRANSACTION `3QPB4KWh9sNVFWsRphryLd4ezmhuzAqUXgRAZPKKRFU7` ApplyState { block_index: 20, prev_block_hash: `6CAU5DSacXRNRWTCpK5b5YAmSX8jcoFVHqfsMF32L9Zc`, block_hash: `EQnM2hdJyVeNNEMip6dotf3vNXb2SMZpRJC5uipg54MY`, epoch_id: EpochId(`11111111111111111111111111111111`), epoch_height: 1, gas_price: 1000000000, block_timestamp: 1626386326477109000, gas_limit: Some(1000000000000000), random_seed: `FtjMPamwU1wzFNdxFdBwHsFAGaKk3Jwr5jBknjXhPsxV`, current_protocol_version: 46, config: RuntimeConfig { storage_amount_per_byte: 10000000000000000000, transaction_costs: RuntimeFeesConfig { action_receipt_creation_config: Fee { send_sir: 108059500000, send_not_sir: 108059500000, execution: 108059500000 }, data_receipt_creation_config: DataReceiptCreationConfig { base_cost: Fee { send_sir: 4697339419375, send_not_sir: 4697339419375, execution: 4697339419375 }, cost_per_byte: Fee { send_sir: 59357464, send_not_sir: 59357464, execution: 59357464 } }, action_creation_config: ActionCreationConfig { create_account_cost: Fee { send_sir: 99607375000, send_not_sir: 99607375000, execution: 99607375000 }, deploy_contract_cost: Fee { send_sir: 184765750000, send_not_sir: 184765750000, execution: 184765750000 }, deploy_contract_cost_per_byte: Fee { send_sir: 6812999, send_not_sir: 6812999, execution: 6812999 }, function_call_cost: Fee { send_sir: 2319861500000, send_not_sir: 2319861500000, execution: 2319861500000 }, function_call_cost_per_byte: Fee { send_sir: 2235934, send_not_sir: 2235934, execution: 2235934 }, transfer_cost: Fee { send_sir: 115123062500, send_not_sir: 115123062500, execution: 115123062500 }, stake_cost: Fee { send_sir: 141715687500, send_not_sir: 141715687500, execution: 102217625000 }, add_key_cost: AccessKeyCreationConfig { full_access_cost: Fee { send_sir: 101765125000, send_not_sir: 101765125000, execution: 101765125000 }, function_call_cost: Fee { send_sir: 102217625000, send_not_sir: 102217625000, execution: 102217625000 }, function_call_cost_per_byte: Fee { send_sir: 1925331, send_not_sir: 1925331, execution: 1925331 } }, delete_key_cost: Fee { send_sir: 94946625000, send_not_sir: 94946625000, execution: 94946625000 }, delete_account_cost: Fee { send_sir: 147489000000, send_not_sir: 147489000000, execution: 147489000000 } }, storage_usage_config: StorageUsageConfig { num_bytes_account: 100, num_extra_bytes_record: 40 }, burnt_gas_reward: Ratio { numer: 3, denom: 10 }, pessimistic_gas_price_inflation_ratio: Ratio { numer: 103, denom: 100 } }, wasm_config: VMConfig { ext_costs: ExtCostsConfig { base: 264768111, contract_compile_base: 35445963, contract_compile_bytes: 216750, read_memory_base: 2609863200, read_memory_byte: 3801333, write_memory_base: 2803794861, write_memory_byte: 2723772, read_register_base: 2517165186, read_register_byte: 98562, write_register_base: 2865522486, write_register_byte: 3801564, utf8_decoding_base: 3111779061, utf8_decoding_byte: 291580479, utf16_decoding_base: 3543313050, utf16_decoding_byte: 163577493, sha256_base: 4540970250, sha256_byte: 24117351, keccak256_base: 5879491275, keccak256_byte: 21471105, keccak512_base: 5811388236, keccak512_byte: 36649701, ripemd160_base: 853675086, ripemd160_block: 680107584, ecrecover_base: 3365369625000, log_base: 3543313050, log_byte: 13198791, storage_write_base: 64196736000, storage_write_key_byte: 70482867, storage_write_value_byte: 31018539, storage_write_evicted_byte: 32117307, storage_read_base: 56356845750, storage_read_key_byte: 30952533, storage_read_value_byte: 5611005, storage_remove_base: 53473030500, storage_remove_key_byte: 38220384, storage_remove_ret_value_byte: 11531556, storage_has_key_base: 54039896625, storage_has_key_byte: 30790845, storage_iter_create_prefix_base: 0, storage_iter_create_prefix_byte: 0, storage_iter_create_range_base: 0, storage_iter_create_from_byte: 0, storage_iter_create_to_byte: 0, storage_iter_next_base: 0, storage_iter_next_key_byte: 0, storage_iter_next_value_byte: 0, touching_trie_node: 16101955926, promise_and_base: 1465013400, promise_and_per_promise: 5452176, promise_return: 560152386, validator_stake_base: 911834726400, validator_total_stake_base: 911834726400 }, grow_mem_cost: 1, regular_op_cost: 3856371, limit_config: VMLimitConfig { max_gas_burnt: 200000000000000, max_gas_burnt_view: 200000000000000, max_stack_height: 16384, initial_memory_pages: 1024, max_memory_pages: 2048, registers_memory_limit: 1073741824, max_register_size: 104857600, max_number_registers: 100, max_number_logs: 100, max_total_log_length: 16384, max_total_prepaid_gas: 300000000000000, max_actions_per_receipt: 100, max_number_bytes_method_names: 2000, max_length_method_name: 256, max_arguments_length: 4194304, max_length_returned_data: 4194304, max_contract_size: 4194304, max_transaction_size: 4194304, max_length_storage_key: 4194304, max_length_storage_value: 4194304, max_promises_per_function_call_action: 1024, max_number_input_data_dependencies: 128 } }, account_creation_config: AccountCreationConfig { min_allowed_top_level_account_length: 0, registrar_account_id: "registrar" } }, cache: Some(Compiled contracts cache), is_new_chunk: true, profile: ERROR: No gas profiled
// DMLOG START_TRANSACTION <TRX_HASH>
func (ctx *parseCtx) readStartTransaction(line string) error {
	if ctx.currentTrace != nil {
		return fmt.Errorf("received when trx already begun")
	}

	chunks, err := SplitInChunks(line, 7)
	if err != nil {
		return fmt.Errorf("split: %s", err)
	}

	hash := chunks[0]
	ctx.currentTrace = &pbcodec.TransactionTrace{
		Receiver: []byte(hash), // change it to a string
		Hash:     []byte(hash),
	}

	return nil
}

// Formats
// DMLOG END_TRANSACTION 223182562500
func (ctx *parseCtx) readEndTransaction(line string) error {
	if ctx.currentTrace == nil {
		return fmt.Errorf("no matching BEGIN_APPLY_TRX")
	}

	trxTrace := ctx.currentTrace

	//chunks, err := SplitInBoundedChunks(line, 6)
	//if err != nil {
	//	return fmt.Errorf("split: %s", err)
	//}
	//
	//gasUsed := chunks[0]

	ctx.currentBlock.TransactionTraces = append(ctx.currentBlock.TransactionTraces, trxTrace)
	ctx.currentTrace = nil
	return nil
}

// Formats
// DMLOG END_BLOCK 20 `EQnM2hdJyVeNNEMip6dotf3vNXb2SMZpRJC5uipg54MY`
func (ctx *parseCtx) readEndBlock(line string) (*pbcodec.Block, error) {
	if ctx.currentBlock == nil {
		return nil, fmt.Errorf("no block started")
	}

	block := ctx.currentBlock
	ctx.transactionTraces = nil
	ctx.currentBlock = nil
	ctx.stats.log()
	return block, nil
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
