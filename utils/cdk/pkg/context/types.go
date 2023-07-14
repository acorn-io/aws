package context

type AwsConfig struct {
	Account string
	Region  string
}

type PluginProvider interface {
	Render(*CdkContext) (map[string]any, error)
}
