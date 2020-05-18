package chezmoi

// recursiveMerge recursively merges maps in source into dest.
func recursiveMerge(dest, source map[string]interface{}) {
	for key, sourceValue := range source {
		destValue, ok := dest[key]
		if !ok {
			dest[key] = sourceValue
			continue
		}
		destMap, ok := destValue.(map[string]interface{})
		if !ok || destMap == nil {
			dest[key] = sourceValue
			continue
		}
		sourceMap, ok := sourceValue.(map[string]interface{})
		if !ok {
			dest[key] = sourceValue
			continue
		}
		recursiveMerge(destMap, sourceMap)
	}
}
