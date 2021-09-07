package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ory/viper"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

func LoadTo(flags *pflag.FlagSet, s interface{}) error {
	if s == nil {
		return fmt.Errorf("nil struct configuration on LoadTo")
	}

	// This way, we can set config from (in order):
	// File: app.storage.dsn=pg (yaml tab)
	// ENV: APP_STORAGE_DSN=pg
	// Flag: --app.storage.dsn=pg or --app.storage.dsn pg
	v := viper.New()
	v.AutomaticEnv()

	// to support env var: convert . to _
	// for example: app.storage.dsn will using key APP_STORAGE_DSN in env variable
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := v.BindPFlags(flags)
	if err != nil {
		return fmt.Errorf("error binding flag viper: %w", err)
	}

	return v.Unmarshal(s)
}

func Setup(cmd *cobra.Command, _ []string, s interface{}) (*zap.Logger, error) {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			MessageKey:     "msg",
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			LineEnding:     zapcore.DefaultLineEnding,
			LevelKey:       "level",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
		}),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), // pipe to multiple writer
		zapcore.DebugLevel,
	)

	log := zap.New(core)

	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return log, fmt.Errorf("error get config filename: %w", err)
	}

	if configFile != "" {
		fileContent, err := ioutil.ReadFile(configFile)
		if err != nil {
			return log, fmt.Errorf("error read file config %s: %w", configFile, err)
		}

		dec := yaml.NewDecoder(bytes.NewReader(fileContent))
		dec.SetStrict(false)
		err = dec.Decode(s) // not pointer because when calling setup must be pointer
		return log, err
	}

	err = LoadTo(cmd.Flags(), s)
	return log, err
}
