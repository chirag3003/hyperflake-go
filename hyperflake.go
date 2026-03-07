package hyperflake

import (
	"fmt"
	"github.com/chirag3003/hyperflake-go/internal"
	"time"
)

// Bit layout of a 64-bit Hyperflake ID:
//
//	[63]    = sign bit      (1  bit)
//	[62:22] = timestamp     (41 bits) — milliseconds since epoch
//	[21:17] = datacenter ID (5  bits)
//	[16:12] = machine ID    (5  bits)
//	[11:0]  = sequence no.  (12 bits)
const (
	timestampShift    = 22
	datacenterIDShift = 17
	machineIDShift    = 12

	timestampMask    = 0x1FFFFFFFFFF // 41 bits
	datacenterIDMask = 0x1F          // 5  bits
	machineIDMask    = 0x1F          // 5  bits
	sequenceMask     = 0xFFF         // 12 bits
)

// Config holds the configuration for generating Hyperflake IDs.
type Config struct {
	epoch          int64 // Epoch timestamp in milliseconds
	datacenterID   int   // Datacenter ID (0–31)
	machineID      int   // Machine ID (0–31)
	sequenceNumber int   // Per-millisecond sequence counter
	signBit        int   // Sign bit (almost always 0)
	lastTimestamp  int64
}

// HyperFlakeID represents a decoded Hyperflake ID.
type HyperFlakeID struct {
	ID                  int64 // Original ID
	Signbit             int   // Sign bit
	DatacenterID        int   // Datacenter ID
	MachineID           int   // Machine ID
	SequenceNumber      int   // Sequence number
	TimestampSinceEpoch int64 // Milliseconds since custom epoch
	Timestamp           int64 // Unix timestamp in milliseconds
}

var DefaultEpoch = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
var defaultEpochMilli = DefaultEpoch.UnixMilli()

/*
NewHyperflakeConfig creates a new HyperflakeConfig with the given parameters.

Parameters:
  - datacenterID: Datacenter ID (0–31).
  - machineID:    Machine ID (0–31).
  - signBit:      Optional sign bit (default is 0).

Returns:
  - A pointer to the newly created Config.
*/
func NewHyperflakeConfig(datacenterID int, machineID int, signBit ...int) *Config {
	sBit := 0
	if len(signBit) > 0 {
		sBit = signBit[0]
	}
	return &Config{
		epoch:        defaultEpochMilli,
		datacenterID: datacenterID,
		machineID:    machineID,
		signBit:      sBit,
	}
}

/*
NewHyperflakeConfigWithEpoch creates a new HyperflakeConfig with a custom epoch.

Parameters:
  - datacenterID: Datacenter ID (0–31).
  - machineID:    Machine ID (0–31).
  - epoch:        Custom epoch timestamp in milliseconds.
  - signBit:      Optional sign bit (default is 0).

Returns:
  - A pointer to the newly created Config.
*/
func NewHyperflakeConfigWithEpoch(datacenterID int, machineID int, epoch int64, signBit ...int) *Config {
	config := NewHyperflakeConfig(datacenterID, machineID, signBit...)
	config.epoch = epoch
	return config
}

// GenerateHyperflakeID generates a new Hyperflake ID based on the current configuration.
func (config *Config) GenerateHyperflakeID() (int64, error) {
	timestamp := internal.GetCurrentTimestampSinceEpoch(config.epoch)

	if timestamp < config.lastTimestamp {
		return 0, fmt.Errorf("clock is moving backwards")
	}

	if timestamp == config.lastTimestamp {
		config.sequenceNumber++
	} else {
		config.sequenceNumber = 0
		config.lastTimestamp = timestamp
	}

	id := (int64(config.signBit) << 63) |
		(timestamp << timestampShift) |
		(int64(config.datacenterID) << datacenterIDShift) |
		(int64(config.machineID) << machineIDShift) |
		int64(config.sequenceNumber)

	return id, nil
}

// DecodeID decodes a given Hyperflake ID into its components.
func (config *Config) DecodeID(id int64) (*HyperFlakeID, error) {
	signBit := int((id >> 63) & 0x1)
	timestamp := (id >> timestampShift) & timestampMask
	datacenterID := int((id >> datacenterIDShift) & datacenterIDMask)
	machineID := int((id >> machineIDShift) & machineIDMask)
	sequenceNumber := int(id & sequenceMask)

	return &HyperFlakeID{
		ID:                  id,
		Signbit:             signBit,
		TimestampSinceEpoch: timestamp,
		Timestamp:           timestamp + config.epoch,
		DatacenterID:        datacenterID,
		MachineID:           machineID,
		SequenceNumber:      sequenceNumber,
	}, nil
}

// SetMachineID sets the machine ID in the configuration.
func (config *Config) SetMachineID(machineID int) {
	config.machineID = machineID
}

// SetDatacenterID sets the datacenter ID in the configuration.
func (config *Config) SetDatacenterID(datacenterID int) {
	config.datacenterID = datacenterID
}

// GetMachineID returns the machine ID from the configuration.
func (config *Config) GetMachineID() int {
	return config.machineID
}

// GetDatacenterID returns the datacenter ID from the configuration.
func (config *Config) GetDatacenterID() int {
	return config.datacenterID
}
