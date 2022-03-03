package tmplfunc

func DerefBool(b *bool) bool {
	return b != nil && *b
}

func DerefInt(i *int, nilValue int) int {
	if i == nil {
		return nilValue
	}

	return *i
}

func DerefFloat32(i *float32, nilValue float32) float32 {
	if i == nil {
		return nilValue
	}

	return *i
}

func DerefString(s *string, nilValue string) string {
	if s == nil {
		return nilValue
	}

	return *s
}
