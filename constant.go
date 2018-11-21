package workerpool

type Status = int32

// status enumerations
const (
	Reserved Status = iota
	Unknown
	Runnable
	Stopped
)
