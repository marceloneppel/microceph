package ceph

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ConfigWriter is an interface for writing config files
type ConfigWriter interface {
	WriteConfig(any) error
}

// Config struct
type Config struct {
	configTemplate *template.Template
	configDir      string
	configFile     string
}

// GetPath returns the path to the config file
func (c *Config) GetPath() string {
	return filepath.Join(c.configDir, c.configFile)
}

// WriteConfig writes the configuration file given a data bag and a filemode
func (c *Config) WriteConfig(data map[string]any, mode int) error {
	fd, err := os.OpenFile(c.GetPath(), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(mode))
	if err != nil {
		return fmt.Errorf("Couldn't write %s: %w", c.configFile, err)
	}
	defer fd.Close()

	err = c.configTemplate.Execute(fd, data)
	if err != nil {
		return fmt.Errorf("Couldn't render %s: %w", c.configFile, err)
	}
	return nil
}

// newCephConfig creates a new ceph.conf
func newCephConfig(configDir string) *Config {
	return &Config{
		configTemplate: template.Must(template.New("cephConf").Parse(`# # Generated by MicroCeph, DO NOT EDIT.
[global]
run dir = {{.runDir}}
fsid = {{.fsid}}
mon host = {{.monitors}}
public_network = {{.pubNet}}
auth allow insecure global id reclaim = false
ms bind ipv4 = {{.ipv4}}
ms bind ipv6 = {{.ipv6}}

[client]
{{if .isCache}}rbd_cache = {{.isCache}}{{end}}
{{if .cacheSize}}rbd_cache_size = {{.cacheSize}}{{end}}
{{if .isCacheWritethrough}}rbd_cache_writethrough_until_flush = {{.isCacheWritethrough}}{{end}}
{{if .cacheMaxDirty}}rbd_cache_max_dirty = {{.cacheMaxDirty}}{{end}}
{{if .cacheTargetDirty}}rbd_cache_target_dirty = {{.cacheTargetDirty}}{{end}}
`)),
		configFile: "ceph.conf",
		configDir:  configDir,
	}
}

// newCephKeyring creates a new Ceph keyring config
func newCephKeyring(configDir, configFile string) *Config {
	return &Config{
		configTemplate: template.Must(template.New("cephKeyring").Parse(`# Generated by MicroCeph, DO NOT EDIT.
[{{.name}}]
	key = {{.key}}
`)),
		configDir:  configDir,
		configFile: configFile,
	}
}

// newRadosGWConfig creates a new radosgw config file
func newRadosGWConfig(configDir string) *Config {
	return &Config{
		configTemplate: template.Must(template.New("radosgwConfig").Parse(`# Generated by MicroCeph, DO NOT EDIT.
[global]
mon host = {{.monitors}}
run dir = {{.runDir}}
auth allow insecure global id reclaim = false

[client.radosgw.gateway]
rgw init timeout = 1200
rgw frontends = beast{{if or (ne .rgwPort 0) (not .sslCertificatePath) (not .sslPrivateKeyPath)}} port={{.rgwPort}}{{end}}{{if and .sslCertificatePath .sslPrivateKeyPath}} ssl_port={{.sslPort}} ssl_certificate={{.sslCertificatePath}} ssl_private_key={{.sslPrivateKeyPath}}{{end}}
`)),
		configFile: "radosgw.conf",
		configDir:  configDir,
	}
}
