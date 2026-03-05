package service

import "github.com/google/uuid"

func unique(IDs []uuid.UUID) []uuid.UUID {
	uniqueIDs := make(map[uuid.UUID]struct{})
	for i := range IDs {
		uniqueIDs[IDs[i]] = struct{}{}
	}

	result := make([]uuid.UUID, 0, len(uniqueIDs))
	for id := range uniqueIDs {
		result = append(result, id)
	}

	return result
}
