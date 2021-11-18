package config

import (
	"encoding/json"
	"errors"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
)

type Config struct {
	Identity     Identity
	Addresses    Addresses
	Datastore    Datastore
	Discovery    Discovery
	AccountLevel AccountLevel
}

const (
	DefaultPathName   = ".pando"
	DefaultPathRoot   = "~/" + DefaultPathName
	DefaultConfigFile = "config"
	EnvDir            = "PANDO_PATH"
)

var (
	ErrInitialized    = errors.New("configuration file already exists")
	ErrNotInitialized = errors.New("not initialized")
)

// Filename returns the configuration file path given a configuration root
// directory. If the configuration root directory is empty, use the default one
func Filename(configRoot string) (string, error) {
	return Path(configRoot, DefaultConfigFile)
}

// Marshal configuration with JSON
func Marshal(value interface{}) ([]byte, error) {
	// need to prettyprint, hence MarshalIndent, instead of Encoder
	return json.MarshalIndent(value, "", "  ")
}

// Path returns the config file path relative to the configuration root. If an
// empty string is provided for `configRoot`, the default root is used.
func Path(configRoot, configFile string) (string, error) {
	if configRoot == "" {
		var err error
		configRoot, err = PathRoot()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(configRoot, configFile), nil
}

// PathRoot returns the default configuration root directory
func PathRoot() (string, error) {
	dir := os.Getenv(EnvDir)
	if dir != "" {
		return dir, nil
	}
	return homedir.Expand(DefaultPathRoot)
}

// Load reads the json-serialized config at the specified path
func Load(filePath string) (*Config, error) {
	var err error
	if filePath == "" {
		filePath, err = Filename("")
		if err != nil {
			return nil, err
		}
	}

	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrNotInitialized
		}
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, err
}

// Save writes the json-serialized config to the specified path
func (c *Config) Save(filePath string) error {
	var err error
	if filePath == "" {
		filePath, err = Filename("")
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	buf, err := Marshal(c)
	if err != nil {
		return err
	}
	_, err = f.Write(buf)
	return err
}
