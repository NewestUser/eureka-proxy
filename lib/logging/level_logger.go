package logging

import (
	"bytes"
	"fmt"

	"github.com/ArthurHlt/gominlog"
)

type levelLogger struct {
	trace bool
	isOn  bool
	log   *gominlog.MinLog
}

func NewLevelLogger(trace, isOn bool) Logger {
	logger := gominlog.NewClassicMinLogWithPackageName("logger")

	return &levelLogger{log: logger, trace: trace, isOn: isOn}
}

func (l *levelLogger) Separator() {
	if l.trace || l.isOn {
		l.log.Debug("----------------------------------------------------------------------------")
	}
}

func (l *levelLogger) Trace(vals ...interface{}) {
	if l.trace && l.isOn {

		l.log.Debug(addSpaces(vals...))
	}
}

func (l *levelLogger) TraceF(format string, vals ...interface{}) string {
	if l.trace && l.isOn {
		msg := fmt.Sprintf(format, vals...)
		l.log.Info(msg)
		return msg
	}

	return ""
}

func (l *levelLogger) Info(vals ...interface{}) {
	if l.isOn {
		l.log.Info(addSpaces(vals...))
	}
}

func (l *levelLogger) InfoF(format string, vals ...interface{}) string {
	if l.isOn {
		msg := fmt.Sprintf(format, vals...)
		l.log.Info(msg)
		return msg
	}

	return ""
}

func (l *levelLogger) Err(vals ...interface{}) {
	l.log.Error(addSpaces(vals...))
}

func (l *levelLogger) ErrF(format string, vals ...interface{}) error {
	err := fmt.Errorf(format, vals...)
	l.log.Error(err.Error())

	return err
}

func addSpaces(vals ...interface{}) string {

	buff := &bytes.Buffer{}

	for argNum, arg := range vals {
		if argNum > 0 {
			buff.WriteByte(' ')
		}

		sprintf := fmt.Sprintf("%v", arg)
		buff.WriteString(sprintf)
	}

	return buff.String()
}
