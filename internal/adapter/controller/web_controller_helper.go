package controller

// configOrDefault は config が nil でなければ toConfig(config) を呼び、
// nil なら defaultCfg を返す。
func configOrDefault[T any, C any](config *T, toConfig func(*T) C, defaultCfg C) C {
	if config != nil {
		return toConfig(config)
	}
	return defaultCfg
}
