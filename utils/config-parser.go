package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TaskName           string `yaml:"task_name"`
	AWSAccessKeyID     string `yaml:"aws_access_key_id"`
	AWSSecretAccessKey string `yaml:"aws_secret_access_key"`
	AWSRegion          string `yaml:"aws_region"`
	S3Bucket           string `yaml:"s3_bucket"`
	WorkingDir         string `yaml:"working_dir"`
	MaxZipSize         int64  `yaml:"max_zip_size"` // in MB
	Tasks              []Task `yaml:"tasks"`
}

type Task struct {
	ID             string   `yaml:"id"`
	BaseDir        string   `yaml:"dir"`
	SkipExtensions []string `yaml:"skip_extensions"`
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

	return cfg
}
