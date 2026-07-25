// Package dataset provides record-oriented I/O against z/OS data sets.
//
// # Status: declared, not scoped
//
// ADR 0003 declares this family and deliberately does not design it. Record I/O
// on z/OS is where the impedance mismatch with Go is largest: a data set has a
// record format (F, FB, V, VB, U), a logical record length and a block size, and
// none of that maps onto io.Reader's "give me some bytes".
//
// The interfaces below are the minimum that is structurally certain — records
// are units, not byte streams — and the concrete implementation waits on scope
// from the owner and on the Phase 2 reading of go-recordio.
package dataset

import "io"

// RecordFormat describes the physical organisation of records in a data set.
type RecordFormat uint8

const (
	// RECFM_F is fixed-length records.
	RECFM_F RecordFormat = iota
	// RECFM_FB is fixed-length blocked.
	RECFM_FB
	// RECFM_V is variable-length records, each with a record descriptor word.
	RECFM_V
	// RECFM_VB is variable-length blocked.
	RECFM_VB
	// RECFM_U is undefined-length records.
	RECFM_U
)

// Attributes describes a data set's record structure.
type Attributes struct {
	Format    RecordFormat
	LRecl     uint16
	BlkSize   uint16
	Allocated bool
}

// RecordReader reads whole records. It is deliberately not an io.Reader: a
// record boundary is meaningful on z/OS and a byte-stream interface would erase
// it.
type RecordReader interface {
	// ReadRecord returns the next record. The returned slice is valid only
	// until the next call. It returns io.EOF at end of data.
	ReadRecord() ([]byte, error)

	// Attributes reports the data set's record structure.
	Attributes() Attributes

	io.Closer
}

// RecordWriter writes whole records.
type RecordWriter interface {
	// WriteRecord writes one record. Records longer than LRecl are an error
	// rather than being silently truncated or split.
	WriteRecord(rec []byte) error

	io.Closer
}
