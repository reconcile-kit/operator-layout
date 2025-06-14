package app

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

type Config struct {
	ShardID     string `validate:"required" label:"--shard-id"`
	StorageURL  string `validate:"required" label:"--storage-url"`
	InformerURL string `validate:"required" label:"--informer-url"`
	LogLevel    int    `label:"--log-level"`
}

func ParseConfig() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ShardID, "shard-id", "", "Shard ID")
	flag.StringVar(&cfg.StorageURL, "storage-url", "", "storage URL")
	flag.StringVar(&cfg.InformerURL, "informer-url", "", "informer broker URL")
	flag.IntVar(&cfg.LogLevel, "log-level", 6, "Log level")

	flag.Parse()

	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(f reflect.StructField) string {
		if lbl := f.Tag.Get("label"); lbl != "" {
			return lbl
		}
		return f.Name
	})
	err := validate.Struct(cfg)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			var validationResult []string
			for _, e := range validationErrors {
				validationResult = append(validationResult, fmt.Sprintf("%s is %s", e.Field(), e.Tag()))
			}
			return nil, fmt.Errorf("invalid arguments: %s", strings.Join(validationResult, ", "))
		}
		return nil, err
	}

	return &cfg, nil
}
