package i18n

type ToString interface {
	String() string
}

func ToStringSlice[T ToString](teamIDs []T) []string {
	ids := []string{}
	for _, id := range teamIDs {
		ids = append(ids, id.String())
	}
	return ids
}
