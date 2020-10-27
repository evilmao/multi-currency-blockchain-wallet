package moneroutil

// This is rewriting in go from monero's C code
func TreeHashCnt(count int) int {
	if count < 3 {
		panic("count must be >= 3")
	}

	if count > 0x10000000 {
		panic("count must be <= 0x10000000")
	}

	pow := uint64(2)
	for pow < uint64(count) {
		pow <<= 1
	}

	return int(pow >> 1)
}

func mergeHashes(h1, h2 Hash) []byte {
	mergedHashes := make([]byte, HashLength*2)
	copy(mergedHashes[0:HashLength], h1[:])
	copy(mergedHashes[HashLength:], h2[:])
	return mergedHashes
}

func TreeHash(hashes []Hash) Hash {
	switch len(hashes) {
	case 0:
		return NullHash
	case 1:
		return hashes[0]
	case 2:
		mergedHashes := mergeHashes(hashes[0], hashes[1])
		return Keccak256(mergedHashes)
	}

	cnt := TreeHashCnt(len(hashes))
	temp := make([]Hash, cnt)
	copy(temp[0:2*cnt-len(hashes)], hashes)

	var i, j int
	for i, j = 2*cnt-len(hashes), 2*cnt-len(hashes); j < cnt; i, j = i+2, j+1 {
		temp[j] = Keccak256(mergeHashes(hashes[i], hashes[i+1]))
	}

	for cnt > 2 {
		cnt >>= 1
		for i, j = 0, 0; j < cnt; i, j = i+2, j+1 {
			temp[j] = Keccak256(mergeHashes(temp[i], temp[i+1]))
		}
	}

	return Keccak256(mergeHashes(temp[0], temp[1]))
}
