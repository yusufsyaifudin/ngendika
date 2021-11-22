package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

func Setup(configFile string, s interface{}) (*zap.Logger, error) {
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

	fileContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		return log, fmt.Errorf("error read file config %s: %w", configFile, err)
	}

	dec := yaml.NewDecoder(bytes.NewReader(fileContent))
	dec.SetStrict(false)
	err = dec.Decode(s) // not pointer because when calling setup must be pointer
	return log, err
}
