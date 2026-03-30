package utils

func SplitText(s string, chunkSize int) []string {
	runes := []rune(s)
	var chunks []string

	for len(runes) > 0 {
		if len(runes) > chunkSize {
			chunks = append(chunks, string(runes[:chunkSize]))
			runes = runes[chunkSize:]
		} else {
			chunks = append(chunks, string(runes))
			break
		}
	}

	return chunks
}
