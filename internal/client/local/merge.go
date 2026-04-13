// Package local содержит типы и функции для работы с локальным состоянием
// клиента GophKeeper.
//
// Пакет отвечает за:
//   - хранение локального состояния;
//   - сохранение и загрузку состояния из файла;
//   - объединение локальных данных с данными, полученными с сервера.
package local

import "time"

// ApplySyncedItems применяет к локальному состоянию список синхронизированных
// элементов, полученных с сервера.
//
// Если элемент уже существует в локальном состоянии, он заменяется новой
// версией. Если элемента еще нет, он добавляется в список.
//
// После успешного применения обновляется время последней синхронизации.
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
