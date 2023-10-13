package constants

const (
	RecordSubcommand  = "record"
	ScriptSubcommand  = "script"
	ProfilingFileName = "perf.data"
	ScriptFileName    = "perf.script"
	CpuClockEvent     = "cpu-clock:"
	CyclesEvent       = "cycles:"
)

const (
	LabelAppName    = "app.kubernetes.io/name"
	AppNameNecoPerf = "necoperf-daemon"
)

const (
	NecoPerfMetricsPort       = 6541
	NecoPerfGrpcServerPort = 6543
	NecoperfGrpcPortName   = "necoperf-grpc"
)
