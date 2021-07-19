package agent

type Config struct {
	BindAddr           string `toml:"bind_addr"`
	BitCaskStoragePath string `toml:"bitcask_storage_path"`
	GCSchedule         string `toml:"gc_schedule"`
	ApiUrl             string `toml:"registry_api_url"`
	ContainerName      string `toml:"registry_container_name"`
	RegistryRootDir    string `toml:"registry_root_dir"`
	RegistryConfig     string `toml:"registry_config"`
}
