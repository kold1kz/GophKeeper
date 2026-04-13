package local

import "time"

func ApplySyncedItems(state *ClientState, items []LocalItem, serverTime time.Time) {
	index := make(map[string]int, len(state.Items))
	for i, item := range state.Items {
		index[item.ID] = i
	}

	for _, item := range items {
		if pos, ok := index[item.ID]; ok {
			state.Items[pos] = item
			continue
		}
		state.Items = append(state.Items, item)
	}

	state.LastSyncAt = &serverTime
}
