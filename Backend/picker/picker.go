package picker

import (
	"tmaxsrv/comm"
)

type State int

const (
	NOT_FOUND State = iota
	NOT_ENOUGH_DATA
	FOUND
)

type PickerFunc func(inData []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint, pack comm.Packet)

func GetPickerFn(stype comm.ScaleCat) PickerFunc {
	switch stype {
	case comm.SCALE_T2200:
		return PickerFnT2200
	case comm.SCALE_TMAX:
		return pickerFnTmax
	case comm.SCALE_TMAX_PASSTH:
		return pickerFnTmaxPassth
	}

	return nil
}
