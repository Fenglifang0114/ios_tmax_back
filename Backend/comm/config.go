package comm

const (
	SCALE_TYPE_UNKNOWN int = iota
	SCALE_TYPE_OLD_C51
	SCALE_TYPE_TMAX
)

var GcurScale = SCALE_TYPE_TMAX
