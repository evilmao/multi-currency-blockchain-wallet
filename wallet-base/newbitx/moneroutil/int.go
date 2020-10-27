package moneroutil

import "unsafe"

func bytesToUint32(buf []byte) uint32 {
	return *(*uint32)(unsafe.Pointer(&buf[0]))
}

// rename it to not mix with Uint64ToBytes!
func uint32ToBytes(val uint32) []byte {
	bytes := make([]byte, 4)
	*(*uint32)(unsafe.Pointer(&bytes[0])) = val
	return bytes
}