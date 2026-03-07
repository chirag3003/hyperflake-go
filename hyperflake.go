package hyperflake

import (
	"fmt"
	"github.com/chirag3003/hyperflake-go/internal"
	"time"
)

// Config holds the configuration for generating Hyperflake IDs.
type Config struct {
	epoch              int64  // Epoch timestamp in milliseconds
	datacenterIDBits   int    // Number of bits for the datacenter ID
	machineIDBits      int    // Number of bits for the machine ID
	sequenceNumberBits int    // Number of bits for the sequence number
	SignBit            int    // Sign bit
	datacenterIDBinary string // Binary representation of the datacenter ID
	machineIDBinary    string // Binary representation of the machine ID
	lastTimestamp      int64
}

// HyperFlakeID represents a decoded Hyperflake ID.
type HyperFlakeID struct {
	ID                  int64 // Original ID
	Signbit             int   // Sign bit
	DatacenterID        int   // Datacenter ID
	MachineID           int   // Machine ID
	SequenceNumber      int   // Sequence number
	TimestampSinceEpoch int64 // TimestampSinceEpoch since epoch
	Timestamp           int64 // Timestamp
}

var DefaultEpoch = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
var defaultEpochMilli = DefaultEpoch.UnixMilli()

/*
NewHyperflakeConfig creates a new HyperflakeConfig with the given parameters.

Parameters:
  - datacenterIDBits:
    Number of bits for the datacenter ID.
  - machineIDBits:
    Number of bits for the machine ID.
  - signBit:
    Optional sign bit (default is 0).

Returns:
- A pointer to the newly created HyperflakeConfig.
*/
func NewHyperflakeConfig(datacenterIDBits int, machineIDBits int, signBit ...int) *Config {
	sBit := 0
	if len(signBit) > 0 {
		sBit = signBit[0]
	}
	return &Config{
		epoch:              defaultEpochMilli,
		datacenterIDBits:   datacenterIDBits,
		machineIDBits:      machineIDBits,
		sequenceNumberBits: 0,
		SignBit:            sBit,
		datacenterIDBinary: internal.IntToBinaryString(datacenterIDBits, 5),
		machineIDBinary:    internal.IntToBinaryString(machineIDBits, 5),
	}
}

/*
NewHyperflakeConfigWithEpoch creates a new HyperflakeConfig with the given parameters and a custom epoch.

Parameters:
- datacenterIDBits: Number of bits for the datacenter ID.
- machineIDBits: Number of bits for the machine ID.
- epoch: Custom epoch timestamp in milliseconds.
- signBit: Optional sign bit (default is 0).

Returns:
- A pointer to the newly created HyperflakeConfig.
*/
func NewHyperflakeConfigWithEpoch(datacenterIDBits int, machineIDBits int, epoch int64, signBit ...int) *Config {
	config := NewHyperflakeConfig(datacenterIDBits, machineIDBits, signBit...)
	config.epoch = epoch
	return config
}

// GenerateHyperflakeID generates a new Hyperflake ID based on the current configuration.
func (config *Config) GenerateHyperflakeID() (int64, error) {
	timestamp := internal.GetCurrentTimestampSinceEpoch(config.epoch)
	if timestamp < config.lastTimestamp {
		return 0, fmt.Errorf("clock is moving backwards")
	}
	signBitBinary := internal.IntToBinaryString(config.SignBit, 1)
	timestampBinary := internal.IntToBinaryString(int(timestamp), 41)
	sequenceNumber := 0
	if timestamp == config.lastTimestamp {
		sequenceNumber = config.sequenceNumberBits
		config.sequenceNumberBits++
	} else {
		config.sequenceNumberBits = 1
		config.lastTimestamp = timestamp
	}
	sequenceNumberBinary := internal.IntToBinaryString(sequenceNumber, 12)

	hyperflakeBinary := internal.BuildString(64,
		signBitBinary,
		timestampBinary,
		config.datacenterIDBinary,
		config.machineIDBinary,
		sequenceNumberBinary,
	)
	hyperflakeID, err := internal.BinaryStringToInt(hyperflakeBinary)
	return hyperflakeID, err
}

// DecodeID decodes a given Hyperflake ID into its components.
func (config *Config) DecodeID(id int64) (*HyperFlakeID, error) {
	// Convert the ID to a 64-bit binary string
	hyperflakeBinary := internal.IntToBinaryString(int(id), 64)

	// Extract and convert the sign bit
	signBitBinary := hyperflakeBinary[:1]
	signBit, err := internal.BinaryStringToInt(signBitBinary)
	if err != nil {
		return nil, err
	}

	// Extract and convert the timestamp
	timestampBinary := hyperflakeBinary[1:42]
	timestamp, err := internal.BinaryStringToInt(timestampBinary)
	if err != nil {
		return nil, err
	}

	// Extract and convert the datacenter ID
	datacenterIDBinary := hyperflakeBinary[42:47]
	datacenterID, err := internal.BinaryStringToInt(datacenterIDBinary)
	if err != nil {
		return nil, err
	}

	// Extract and convert the machine ID
	machineIDBinary := hyperflakeBinary[47:52]
	machineID, err := internal.BinaryStringToInt(machineIDBinary)
	if err != nil {
		return nil, err
	}

	// Extract and convert the sequence number
	sequenceNumberBinary := hyperflakeBinary[52:64]
	sequenceNumber, err := internal.BinaryStringToInt(sequenceNumberBinary)
	if err != nil {
		return nil, err
	}

	// Create and return the HyperFlakeID struct
	hyperflakeID := &HyperFlakeID{
		ID:                  id,
		Signbit:             int(signBit),
		TimestampSinceEpoch: timestamp,
		Timestamp:           timestamp + config.epoch,
		DatacenterID:        int(datacenterID),
		MachineID:           int(machineID),
		SequenceNumber:      int(sequenceNumber),
	}
	return hyperflakeID, nil
}

// SetMachineID sets the machine ID in the configuration.
func (config *Config) SetMachineID(machineID int) {
	config.machineIDBits = machineID
	config.machineIDBinary = internal.IntToBinaryString(machineID, 5)
}

// SetDatacenterID sets the datacenter ID in the configuration.
func (config *Config) SetDatacenterID(datacenterID int) {
	config.datacenterIDBits = datacenterID
	config.datacenterIDBinary = internal.IntToBinaryString(datacenterID, 5)
}

// GetMachineID returns the machine ID from the configuration.
func (config *Config) GetMachineID() int {
	return config.machineIDBits
}

// GetDatacenterID returns the datacenter ID from the configuration.
func (config *Config) GetDatacenterID() int {
	return config.datacenterIDBits
}
