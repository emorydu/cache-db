package driver

type Logger interface {
	Fatal(string, ...any)
	Error(string, ...any)
	Warn(string, ...any)
	Info(string, ...any)
	Debug(string, ...any)
	Trace(string, ...any)
}
