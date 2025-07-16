package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type BaseConfig struct {
	S3BasePath         string `yaml:"s3_base_path"`
	AWSAccessKeyID     string `yaml:"aws_access_key_id"`
	AWSSecretAccessKey string `yaml:"aws_secret_access_key"`
	AWSRegion          string `yaml:"aws_region"`
	S3Bucket           string `yaml:"s3_bucket"`
	WorkingDir         string `yaml:"working_dir"`
	MaxZipSize         int64  `yaml:"max_zip_size"` // in MB
}

type Config struct {
	BaseConfig `yaml:",inline"`
	Tasks      []Task `yaml:"tasks"`
}

type Task struct {
	ID             string   `yaml:"id"`
	BaseDir        string   `yaml:"dir"`
	SkipExtensions []string `yaml:"skip_extensions"`
	Password       string   `yaml:"password"`
	UseChecksum    bool     `yaml:"use_checksum"`
}

type TaskConfig struct {
	BaseConfig
	Task
}

func (c *Config) GetTask(taskId string) (*TaskConfig, error) {
	for _, task := range c.Tasks {
		if task.ID == taskId {
			return &TaskConfig{BaseConfig: c.BaseConfig, Task: task}, nil
		}
	}
	return nil, fmt.Errorf("task not found")
}

func GetConfig() Config {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		panic(err)
	}
	cfg.Validate()
	return cfg
}

func (c *BaseConfig) NewZipFileNameForTask(taskId string, index int) string {
	if _, err := os.Stat(c.WorkingDir); os.IsNotExist(err) {
		err := os.Mkdir(c.WorkingDir, 0755)
		if err != nil {
			panic(err)
		}
	}

	zipSuffix := time.Now().UTC().Format("2006-01-02_15_04_05")
	if index > 0 {
		zipSuffix = zipSuffix + "_" + fmt.Sprintf("%d", index)
	}

	zipName := strings.ReplaceAll(taskId, " ", "_") + "_" + zipSuffix + ".zip"

	return filepath.Join(c.WorkingDir, zipName)
}

func required(value string, name string) {
	if value == "" {
		Err(fmt.Sprintf("%s is required", name))
	}
}

func (c *Config) Validate() {
	if c.S3BasePath == "" {
		c.S3BasePath = ""
	}

	if strings.HasSuffix(c.S3BasePath, "/") {
		c.S3BasePath = c.S3BasePath[:len(c.S3BasePath)-1]
	}

	required(c.AWSAccessKeyID, "AWS access key id")
	required(c.AWSSecretAccessKey, "AWS secret access key")
	required(c.AWSRegion, "AWS region")
	required(c.S3Bucket, "S3 bucket")
	required(c.WorkingDir, "Working dir")

	if c.MaxZipSize <= 5 {
		Err("Max zip size must be greater than 5MB")
	}

	for _, task := range c.Tasks {
		task.Validate()
	}
}
func (t *Task) Validate() {
	required(t.ID, "Task id")
	required(t.BaseDir, fmt.Sprintf("Task - %s base dir", t.ID))
	if t.SkipExtensions == nil {
		t.SkipExtensions = []string{}
	}
	for _, ext := range t.SkipExtensions {
		if ext == "" {
			Err("Task skip extensions cannot be empty")
		}
	}

}
