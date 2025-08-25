package templates

type Context map[string]any

func (c Context) Add(key string, value any) {
	c[key] = value
}
