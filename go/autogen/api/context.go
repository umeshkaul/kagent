package api

type ChatCompletionContextConfig struct{}

func (c *ChatCompletionContextConfig) ToConfig() (map[string]interface{}, error) {
	return toConfig(c)
}

func (c *ChatCompletionContextConfig) FromConfig(config map[string]interface{}) error {
	return fromConfig(c, config)
}
