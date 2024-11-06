package cmd

import "tmaxsrv/comm"

const (
	PRN_FMT_FLASH_ADDR_T2200   = 0x1801e000
	PRN_FMT_SIZE_T2200         = 2048
	PRN_FMT_FLASH_ADDR_TMAX    = 0x18005000
	PRN_FMT_SIZE_TMAX          = 2048
	SERIAL_FMT_FLASH_ADDR_TMAX = 0x18007000
	SERIAL_FMT_SIZE_TMAX       = 2048
	PLU_ROM_ADDR_TMAX          = 0x2006A000
	PLU_ROM_SIZE_TMAX          = 4096
)

const (
	CMD_TIMEOUT_IMMEDIATE          int = -1
	CMD_TIMEOUT_VERY_SHORT_200_MS  int = 1000
	CMD_TIMEOUT_SHORT_1500_MS      int = 1500
	CMD_TIMEOUT_MEDIUM_2000_MS     int = 500 //Test 20241031
	CMD_TIMEOUT_MED_LONG_4000_MS   int = 4000
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

func GetSerialFmtAddrNSize(scaleCat comm.ScaleCat) (int, int) {
	switch scaleCat {
	case comm.SCALE_TMAX:
		return SERIAL_FMT_FLASH_ADDR_TMAX, SERIAL_FMT_SIZE_TMAX
	}
	return -1, -1
}

func GetPluRomAddrNSize(scaleCat comm.ScaleCat) (int, int) {
	switch scaleCat {
	case comm.SCALE_TMAX:
		return PLU_ROM_ADDR_TMAX, PLU_ROM_SIZE_TMAX
	}
	return -1, -1
}
