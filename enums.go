package xz

type Action uint

const (
	Run Action = iota
	SyncFlush
	FullFlush
	Finish
)

type Errno uint

var _ error = Errno(0)

const (
	Ok Errno = iota
	StreamEnd
	NoCheck
	UnsupportedCheck
	GetCheck
	MemError
	MemlimitError
	FormatError
	OptionsError
	DataError
	BufError
	ProgError
)

var errorMsg = [...]string{
	"Operation completed successfully",
	"End of stream was reached",
	"Input stream has no integrity check",
	"Cannot calculate the integrity check",
	"Integrity check type is now available",
	"Cannot allocate memory",
	"Memory usage limit was reached",
	"File format not recognized",
	"Invalid or unsupported options",
	"Data is corrupt",
	"No progress is possible",
	"Programming error",
}

func (e Errno) Error() string {
	return errorMsg[e]
}
