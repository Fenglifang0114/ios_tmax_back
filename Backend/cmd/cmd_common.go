package cmd

import "tmaxsrv/comm"

const (
	PRN_FMT_FLASH_ADDR_T2200 = 0x0801e000
	PRN_FMT_SIZE_T2200       = 2048
	PRN_FMT_FLASH_ADDR_TMAX  = 0x0801e000
	PRN_FMT_SIZE_TMAX        = 2048
)

const (
	CMD_TIMEOUT_IMMEDIATE          int = -1
	CMD_TIMEOUT_SHORT_100_MS       int = 100
	CMD_TIMEOUT_MEDIUM_2000_MS     int = 2000
	CMD_TIMEOUT_LONG_20000_MS      int = 20000
	CMD_TIMEOUT_VERY_LONG_60000_MS int = 60000
	CMD_TIMEOUT_NEVER_999999999_MS int = 999999999
)

func GetPrnFmtAddrNSize(scaleCat comm.ScaleCat, orderNo int) (int, int) {
	switch scaleCat {
	case comm.SCALE_T2200:
		return PRN_FMT_FLASH_ADDR_T2200 + (orderNo-1)*PRN_FMT_SIZE_T2200, PRN_FMT_SIZE_T2200
	case comm.SCALE_TMAX:
		return PRN_FMT_FLASH_ADDR_TMAX + (orderNo-1)*PRN_FMT_SIZE_TMAX, PRN_FMT_SIZE_TMAX
	}

	return -1, -1
}
