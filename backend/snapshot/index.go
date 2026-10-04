package snapshot

// SubjectIndex groups all snapshot blocks by their exact subject name.
type SubjectIndex map[string][]Block

// NewSubjectIndex retains every candidate, including parked blocks, for offline
// matching. Subject alone is not an identity; callers must narrow candidates.
func NewSubjectIndex(blocks []Block) SubjectIndex {
	index := make(SubjectIndex)
	for _, block := range blocks {
		index[block.Subject] = append(index[block.Subject], block)
	}
	return index
}
