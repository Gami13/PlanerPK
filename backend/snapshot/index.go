package snapshot

// IndexBySubject retains every candidate, including parked blocks, for offline
// matching. Subject alone is not an identity; callers must narrow candidates.
func IndexBySubject(blocks []Block) map[string][]Block {
	index := make(map[string][]Block)
	for _, block := range blocks {
		index[block.Subject] = append(index[block.Subject], block)
	}
	return index
}
