package lx

// SetCached sets a function result to a constant value
// For example, "version()": "1.2".
func (px *PX) SetCached(key string, value any) {
	px.mu.Lock()
	defer px.mu.Unlock()
	px.cached[key] = value
}
