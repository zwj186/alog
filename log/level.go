package log

// LogLevel 日志级别
type LogLevel byte

const (
	DEBUG LogLevel = iota + 1
	INFO
	WARN
	ERROR
	FATAL
)

func (l LogLevel) ToString() string {
	var v string
	switch l {
	case DEBUG:
		v = "D" //"Debug"
	case INFO:
		v = "I" //"Info"
	case WARN:
		v = "W" //"Warn"
	case ERROR:
		v = "E" //"Error"
	case FATAL:
		v = "F" //"Fatal"
	}
	return v
}

func (l LogLevel) ToLowerString() string {
	var v string
	switch l {
	case DEBUG:
		v = "debug"
	case INFO:
		v = "info"
	case WARN:
		v = "warn"
	case ERROR:
		v = "error"
	case FATAL:
		v = "fatal"
	}
	return v
}

func (l LogLevel) ToUpperString() string {
	var v string
	switch l {
	case DEBUG:
		v = "DEBUG"
	case INFO:
		v = "INFO"
	case WARN:
		v = "WARN"
	case ERROR:
		v = "ERROR"
	case FATAL:
		v = "FATAL"
	}
	return v
}