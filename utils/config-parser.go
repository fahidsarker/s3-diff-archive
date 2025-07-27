package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	nTypes "s3-diff-archive/types"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// var SUPPORTED_STORAGE_CLASSES = []string{"STANDARD", "INTELLIGENT_TIERING", "STANDARD_IA", "ONEZONE_IA", "GLACIER", "DEEP_ARCHIVE"}

type Secrets struct {
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	S3Bucket           string
}

type BaseConfig struct {
	Secrets
	NotifyScript string `yaml:"notify_script"`
	MaxZipSize   int64  `yaml:"max_zip_size"` // in MB
	S3BasePath   string `yaml:"s3_base_path"`
	WorkingDir   string `yaml:"working_dir"`
	LogsDir      string `yaml:"logs_dir"`
}

type Config struct {
	BaseConfig `yaml:",inline"`
	Tasks      []Task `yaml:"tasks"`
}

type Task struct {
	ID                 string   `yaml:"id"`
	Dir                string   `yaml:"dir"`
	Excludes           []string `yaml:"exclude"`
	StorageClassString string   `yaml:"storage_class"`
	UseChecksum        bool     `yaml:"use_checksum"`
	Password           string   `yaml:"encryption_key"`
	StorageClass       types.StorageClass
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

func getSecretsFromEnv(envPath string) Secrets {
	err := godotenv.Load(envPath)
	if err != nil {
		panic("Error loading .env file or no .env file found")
	}
	return Secrets{
		AWSAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          os.Getenv("AWS_REGION"),
		S3Bucket:           os.Getenv("S3_BUCKET"),
	}
}

func GetConfig(path string, envPath string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	cfg.Secrets = getSecretsFromEnv(envPath)
	if err != nil {
		panic(err)
	}
	cfg.Validate()
	return &cfg
}

func (c *BaseConfig) NewZipFileNameForTask(taskId string, index int, extras ...string) string {
	if _, err := os.Stat(c.WorkingDir); os.IsNotExist(err) {
		err := os.Mkdir(c.WorkingDir, 0755)
		if err != nil {
			panic(err)
		}
	}

	zipSuffix := time.Now().UTC().Format("2006_01_02_15_04_05")
	if index > 0 {
		zipSuffix = zipSuffix + "_" + fmt.Sprintf("%d", index)
	}

	zipName := strings.ReplaceAll(taskId, " ", "_") + "_" + zipSuffix + strings.Join(extras, "_") + ".zip"

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

	c.S3BasePath = strings.TrimSuffix(c.S3BasePath, "/")

	required(c.AWSAccessKeyID, "AWS access key id")
	required(c.AWSSecretAccessKey, "AWS secret access key")
	required(c.AWSRegion, "AWS region")
	required(c.S3Bucket, "S3 bucket")
	required(c.WorkingDir, "Working dir")

	if c.MaxZipSize <= 5 {
		Err("Max zip size must be greater than 5MB")
	}

	for i := range c.Tasks {
		c.Tasks[i].validate()
	}

}
func (t *Task) validate() {
	required(t.ID, "Task id")
	required(t.Dir, fmt.Sprintf("Task - %s base dir", t.ID))
	if t.Excludes == nil {
		t.Excludes = []string{}
	}
	for _, ext := range t.Excludes {
		if ext == "" {
			Err("Task Exclude Regex cannot be empty")
		}
	}

	if t.StorageClassString == "" {
		// println("Empty storage class")
		t.StorageClass = types.StorageClassDeepArchive
	} else {
		invalid := true
		var supported []string
		for _, sc := range types.StorageClassDeepArchive.Values() {
			if sc == types.StorageClass(t.StorageClassString) {
				t.StorageClass = sc
				invalid = false
				break
			}
			supported = append(supported, string(sc))
		}
		if invalid {
			Err(fmt.Sprintf("Invalid storage class: %s. Supported storage classes: %s", t.StorageClassString, strings.Join(supported, ", ")))
		}
	}

}

func (t *TaskConfig) CreateS3Config(storageCls types.StorageClass) *nTypes.S3Config {
	return &nTypes.S3Config{
		AccessKeyID:     t.AWSAccessKeyID,
		SecretAccessKey: t.AWSSecretAccessKey,
		S3BasePath:      fmt.Sprintf("%s/%s", t.S3BasePath, t.ID),
		StorageClass:    storageCls,
		Region:          t.AWSRegion,
		S3Bucket:        t.S3Bucket,
	}
}
