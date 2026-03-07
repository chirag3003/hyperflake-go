package hyperflake

import (
	"sync"
	"testing"
)

// TestDecodeID verifies that known IDs decode to the expected components.
func TestDecodeID(t *testing.T) {
	hf := NewHyperflakeConfig(0, 0)

	testCases := []struct {
		name                   string
		id                     int64
		expectedSignbit        int
		expectedTimestamp      int64
		expectedDatacenterID   int
		expectedMachineID      int
		expectedSequenceNumber int
	}{
		{
			name:                   "zero datacenter and machine ID",
			id:                     3282575599297626112,
			expectedSignbit:        0,
			expectedTimestamp:      1729311810178,
			expectedDatacenterID:   0,
			expectedMachineID:      0,
			expectedSequenceNumber: 0,
		},
		{
			name:                   "non-zero datacenter, machine ID and sequence number",
			id:                     3273974649684016911,
			expectedSignbit:        0,
			expectedTimestamp:      1727261183992,
			expectedDatacenterID:   6,
			expectedMachineID:      11,
			expectedSequenceNumber: 3855,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoded, err := hf.DecodeID(tc.id)
			if err != nil {
				t.Fatalf("DecodeID(%d) returned error: %v", tc.id, err)
			}
			if decoded.Signbit != tc.expectedSignbit {
				t.Errorf("Signbit = %d; want %d", decoded.Signbit, tc.expectedSignbit)
			}
			if decoded.Timestamp != tc.expectedTimestamp {
				t.Errorf("Timestamp = %d; want %d", decoded.Timestamp, tc.expectedTimestamp)
			}
			if decoded.DatacenterID != tc.expectedDatacenterID {
				t.Errorf("DatacenterID = %d; want %d", decoded.DatacenterID, tc.expectedDatacenterID)
			}
			if decoded.MachineID != tc.expectedMachineID {
				t.Errorf("MachineID = %d; want %d", decoded.MachineID, tc.expectedMachineID)
			}
			if decoded.SequenceNumber != tc.expectedSequenceNumber {
				t.Errorf("SequenceNumber = %d; want %d", decoded.SequenceNumber, tc.expectedSequenceNumber)
			}
		})
	}
}

// TestGenerateAndDecode verifies that a generated ID round-trips correctly through DecodeID.
func TestGenerateAndDecode(t *testing.T) {
	const datacenterID = 12
	const machineID = 7

	config := NewHyperflakeConfig(datacenterID, machineID)
	id, err := config.GenerateHyperflakeID()
	if err != nil {
		t.Fatalf("GenerateHyperflakeID() error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("GenerateHyperflakeID() returned non-positive ID: %d", id)
	}

	decoded, err := config.DecodeID(id)
	if err != nil {
		t.Fatalf("DecodeID(%d) error: %v", id, err)
	}
	if decoded.DatacenterID != datacenterID {
		t.Errorf("DatacenterID = %d; want %d", decoded.DatacenterID, datacenterID)
	}
	if decoded.MachineID != machineID {
		t.Errorf("MachineID = %d; want %d", decoded.MachineID, machineID)
	}
	if decoded.ID != id {
		t.Errorf("ID = %d; want %d", decoded.ID, id)
	}
}

// TestConcurrency verifies that GenerateHyperflakeID is safe for concurrent use
// and produces no duplicate IDs across goroutines.
func TestConcurrency(t *testing.T) {
	const goroutines = 100
	const idsPerGoroutine = 100

	config := NewHyperflakeConfig(1, 1)
	ids := make(chan int64, goroutines*idsPerGoroutine)

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := config.GenerateHyperflakeID()
				if err != nil {
					t.Errorf("GenerateHyperflakeID() error: %v", err)
					return
				}
				ids <- id
			}
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[int64]struct{}, goroutines*idsPerGoroutine)
	for id := range ids {
		if _, exists := seen[id]; exists {
			t.Errorf("duplicate ID generated: %d", id)
		}
		seen[id] = struct{}{}
	}
}

// TestClockBackwards verifies that GenerateHyperflakeID returns an error
// if the clock appears to move backwards.
func TestClockBackwards(t *testing.T) {
	config := NewHyperflakeConfig(0, 0)
	// Simulate a future lastTimestamp to trigger the clock-backwards guard.
	config.lastTimestamp = 1<<41 - 1 // max possible timestamp

	_, err := config.GenerateHyperflakeID()
	if err == nil {
		t.Fatal("expected clock-backwards error, got nil")
	}
}

// TestCustomEpoch verifies that NewHyperflakeConfigWithEpoch uses the provided epoch
// and that the decoded timestamp reflects it correctly.
func TestCustomEpoch(t *testing.T) {
	customEpoch := int64(1_000_000_000_000) // arbitrary epoch in ms

	config := NewHyperflakeConfigWithEpoch(1, 1, customEpoch)
	id, err := config.GenerateHyperflakeID()
	if err != nil {
		t.Fatalf("GenerateHyperflakeID() error: %v", err)
	}

	decoded, err := config.DecodeID(id)
	if err != nil {
		t.Fatalf("DecodeID(%d) error: %v", id, err)
	}

	// Timestamp field should equal TimestampSinceEpoch + customEpoch.
	if decoded.Timestamp != decoded.TimestampSinceEpoch+customEpoch {
		t.Errorf("Timestamp = %d; want TimestampSinceEpoch(%d) + epoch(%d) = %d",
			decoded.Timestamp, decoded.TimestampSinceEpoch, customEpoch, decoded.TimestampSinceEpoch+customEpoch)
	}
}

// TestSettersGetters verifies Set/Get methods for datacenter and machine IDs.
func TestSettersGetters(t *testing.T) {
	config := NewHyperflakeConfig(0, 0)

	config.SetDatacenterID(15)
	if got := config.GetDatacenterID(); got != 15 {
		t.Errorf("GetDatacenterID() = %d; want 15", got)
	}

	config.SetMachineID(31)
	if got := config.GetMachineID(); got != 31 {
		t.Errorf("GetMachineID() = %d; want 31", got)
	}

	// Verify the updated IDs are encoded in generated IDs.
	id, err := config.GenerateHyperflakeID()
	if err != nil {
		t.Fatalf("GenerateHyperflakeID() error: %v", err)
	}
	decoded, err := config.DecodeID(id)
	if err != nil {
		t.Fatalf("DecodeID error: %v", err)
	}
	if decoded.DatacenterID != 15 {
		t.Errorf("encoded DatacenterID = %d; want 15", decoded.DatacenterID)
	}
	if decoded.MachineID != 31 {
		t.Errorf("encoded MachineID = %d; want 31", decoded.MachineID)
	}
}

func BenchmarkGenerateHyperflakeID(b *testing.B) {
	config := NewHyperflakeConfig(3, 7)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = config.GenerateHyperflakeID()
	}
}

func BenchmarkDecodeID(b *testing.B) {
	config := NewHyperflakeConfig(3, 7)
	id, _ := config.GenerateHyperflakeID()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = config.DecodeID(id)
	}
}
