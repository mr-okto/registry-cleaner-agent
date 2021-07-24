package agent

type Config struct {
	BindAddr           string   `toml:"bind_addr"`
	CorsAllowedOrigins []string `toml:"cors_allowed_origins"`
	CorsAllowedHeaders []string `toml:"cors_allowed_headers"`
	CorsExposedHeaders []string `toml:"cors_exposed_headers"`
	BitCaskStoragePath string   `toml:"bitcask_storage_path"`
	GCIndexSchedule    string   `toml:"gc_index_schedule"`
	GCRemovalSchedule  string   `toml:"gc_removal_schedule"`
	ApiUrl             string   `toml:"registry_api_url"`
	ContainerName      string   `toml:"registry_container_name"`
	RegistryMountPoint string   `toml:"registry_mount_point"`
	RegistryConfig     string   `toml:"registry_config_path"`
}
