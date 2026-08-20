package logger

import (
	"github.com/fatih/color"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func LogInfo(format string, args ...interface{}) {
	klog.Infof(color.CyanString(format, args...))
}

func LogError(format string, args ...interface{}) {
	klog.Errorf(color.RedString(format, args...))
}

func LogFatal(format string, args ...interface{}) {
	klog.Fatalf(color.RedString(format, args...))
}

func LogSuccess(format string, args ...interface{}) {
	klog.Infof(color.GreenString(format, args...))
}

func LogWarning(format string, args ...interface{}) {
	klog.Warningf(color.YellowString(format, args...))
}

func PhaseColor(phase v1.PodPhase) string {
	switch phase {
	case v1.PodPending:
		return color.YellowString(string(phase))
	case v1.PodRunning:
		return color.GreenString(string(phase))
	case v1.PodSucceeded:
		return color.GreenString(string(phase))
	case v1.PodFailed:
		return color.RedString(string(phase))
	default:
		return string(phase)
	}
}
