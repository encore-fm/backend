package events

func AsGroupIDs(xs []string) []GroupID {
	ids := make([]GroupID, len(xs))
	for _, x := range xs {
		ids = append(ids, GroupID(x))
	}
	return ids
}
