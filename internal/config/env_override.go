package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// ApplyEnvOverrides 使用 `env` 标签中的环境变量覆盖配置值。
func ApplyEnvOverrides(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	return applyEnv(reflect.ValueOf(cfg).Elem())
}

func applyEnv(value reflect.Value) error {
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		fieldType := value.Type().Field(index)

		if field.Kind() == reflect.Struct {
			if err := applyEnv(field); err != nil {
				return err
			}
		}

		envKey := fieldType.Tag.Get("env")
		if envKey == "" {
			continue
		}

		raw, ok := os.LookupEnv(envKey)
		if !ok {
			continue
		}

		if err := setFieldValue(field, raw); err != nil {
			return fmt.Errorf("%s: %w", envKey, err)
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, raw string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
		return nil
	case reflect.Bool:
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		field.SetBool(parsed)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(parsed)
		return nil
	case reflect.Slice:
		if field.Type().Elem().Kind() != reflect.String {
			return fmt.Errorf("unsupported slice type %s", field.Type().String())
		}
		parts := strings.Split(raw, ",")
		values := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				values = append(values, trimmed)
			}
		}
		field.Set(reflect.ValueOf(values))
		return nil
	default:
		return fmt.Errorf("unsupported field kind %s", field.Kind())
	}
}
