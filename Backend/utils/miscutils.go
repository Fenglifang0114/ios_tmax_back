package utils

import (
	mcrc "github.com/BertoldVdb/go-misc/multicrc"
)

// 计算Crc32MPEG2
func Crc32MPEG2(input []byte) uint32 {
	crc := mcrc.NewCRC(mcrc.Crc32MPEG2)
	crc.Reset().AddBytes(input)
	return crc.Result32()
}
