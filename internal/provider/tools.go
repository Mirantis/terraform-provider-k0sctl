package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/k0sproject/rig"
	"github.com/sirupsen/logrus"
)

// AllLoggingToTFLog turns on passing of ruslog and log to tflog.
func AllLoggingToTFLog(ctx context.Context) {
	logrus.AddHook(logrusTFLogHandler{ctx: ctx})
	logrus.SetLevel(logrus.TraceLevel) // trace all log levels, as we don't know what to catch yet.

	rig.SetLogger(rigTFLogLogger{ctx: ctx})

}

// logRusTFLogHandler a tflog handler which integrates logrus so that logrus output gets handled natively.
type logrusTFLogHandler struct {
	ctx context.Context
}

// Receive a logrus event.
func (lh logrusTFLogHandler) Fire(e *logrus.Entry) error {
	go func(event *logrus.Entry) {
		//lh.ctx = context.Background()
		//ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		//defer cancel()
		logrusTFLogFire(lh.ctx, event)
	}(e)

	return nil
}

// Levels that this logrus hook will handle.
func (lh logrusTFLogHandler) Levels() []logrus.Level {
	return logrus.AllLevels
}

func logrusTFLogFire(ctx context.Context, e *logrus.Entry) {
	mes := e.Message
	addFields := map[string]interface{}{
		"pipe": "logrusTFLogFire",
	}

	switch e.Level {
	case logrus.DebugLevel:
		tflog.Debug(ctx, mes, addFields)
	case logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel:
		tflog.Error(ctx, mes, addFields)
	case logrus.InfoLevel:
		tflog.Info(ctx, mes, addFields)
	case logrus.WarnLevel:
		tflog.Warn(ctx, mes, addFields)
	}
}

// rigTFLogLogger Logger that converts k0sProject logging to tflog.
// @NOTE we re-use the logrus levels for convenience - but this has nothing to do with logrus.
type rigTFLogLogger struct {
	ctx context.Context
}

func (l rigTFLogLogger) Tracef(msg string, values ...interface{}) {
	rigLoggerTFLogFire(logrus.TraceLevel, msg, l.ctx, values...)
}
func (l rigTFLogLogger) Debugf(msg string, values ...interface{}) {
	rigLoggerTFLogFire(logrus.DebugLevel, msg, l.ctx, values...)
}
func (l rigTFLogLogger) Infof(msg string, values ...interface{}) {
	rigLoggerTFLogFire(logrus.InfoLevel, msg, l.ctx, values...)
}
func (l rigTFLogLogger) Warnf(msg string, values ...interface{}) {
	rigLoggerTFLogFire(logrus.WarnLevel, msg, l.ctx, values...)
}
func (l rigTFLogLogger) Errorf(msg string, values ...interface{}) {
	rigLoggerTFLogFire(logrus.ErrorLevel, msg, l.ctx, values...)
}

// rigLoggerTFLogFire Take a k0sProject.Rig log entry, and fire a tflog entry.
func rigLoggerTFLogFire(level logrus.Level, entry string, ctx context.Context, values ...interface{}) {
	go func(msg string) {
		//ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		//defer cancel()

		addFields := map[string]interface{}{
			"pipe": "rigTFLogLogger",
		}

		switch level {
		case logrus.DebugLevel:
			tflog.Debug(ctx, msg, addFields)
		case logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel:
			tflog.Error(ctx, msg, addFields)
		case logrus.InfoLevel:
			tflog.Info(ctx, msg, addFields)
		case logrus.WarnLevel:
			tflog.Warn(ctx, msg, addFields)
		}

	}(fmt.Sprintf(entry, values...))
}

// this decodes some strings in the file that are base64 encoded.
func helperStringBase64Decode(val string) string {
	valDecodedBytes, _ := base64.StdEncoding.DecodeString(val)
	return string(valDecodedBytes)
}
