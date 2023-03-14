// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zap

import (
	"os"
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	logger.RegisterConstructor("zap", func() logger.Logger {
		cfg := &Config{}
		if err := config.Get("zap").Scan(cfg); err != nil {
			logger.FatalFiled("fault to load zap config", logger.Err(err))
		}
		return cfg.Build()
	})
}

var (
	mu              sync.Mutex
	fileWriteSyncer zapcore.WriteSyncer
)

func getWriteSyncer(cfg *Config) zapcore.WriteSyncer {
	mu.Lock()
	defer mu.Unlock()
	if fileWriteSyncer == nil {
		fileWriteSyncer = newFileSyncer(&cfg.File.FileConfig)
	}
	return fileWriteSyncer
}

type Logger struct {
	cfg *Config
	lg  *zap.Logger
	*zap.SugaredLogger
	lv *zap.AtomicLevel
}

func (lg *Logger) Clone() logger.Logger {
	lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	lv.SetLevel(lg.Level())
	return newLogger(&lv, lg.cfg)
}

var _ logger.Logger = (*Logger)(nil)

func (lg *Logger) SetLevel(lv logger.Level) {
	switch lv {
	case logger.LvDebug:
		lg.lv.SetLevel(zap.DebugLevel)
	case logger.LvInfo:
		lg.lv.SetLevel(zap.InfoLevel)
	case logger.LvWarn:
		lg.lv.SetLevel(zap.WarnLevel)
	case logger.LvError:
		lg.lv.SetLevel(zap.ErrorLevel)
	case logger.LvFault:
		lg.lv.SetLevel(zap.FatalLevel)
	}
}

func (lg *Logger) Enable(lv logger.Level) bool {
	switch lv {
	case logger.LvDebug:
		return lg.lv.Enabled(zap.DebugLevel)
	case logger.LvInfo:
		return lg.lv.Enabled(zap.InfoLevel)
	case logger.LvWarn:
		return lg.lv.Enabled(zap.WarnLevel)
	case logger.LvError:
		return lg.lv.Enabled(zap.ErrorLevel)
	case logger.LvFault:
		return lg.lv.Enabled(zap.FatalLevel)
	}
	return false
}

func (lg *Logger) GetLevel() logger.Level {
	switch lg.lv.Level() {
	case zap.DebugLevel:
		return logger.LvDebug
	case zap.InfoLevel:
		return logger.LvInfo
	case zap.WarnLevel:
		return logger.LvWarn
	case zap.ErrorLevel:
		return logger.LvError
	case zap.FatalLevel:
		return logger.LvFault
	}
	return logger.LvDebug
}

func (lg *Logger) ZapLogger() *zap.Logger {
	return lg.lg
}

func newLogger(lv *zap.AtomicLevel, config *Config) *Logger {
	zapOptions := make([]zap.Option, 0)
	zapOptions = append(zapOptions, zap.AddStacktrace(zap.PanicLevel))
	if config.AddCaller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	}
	cores := make([]zapcore.Core, 0, 1)
	isErr := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel && lv.Level() <= zapcore.ErrorLevel
	})
	isNotErr := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lv.Level() <= lvl
	})

	if config.Console.Enable {
		var wsOut, wsErr = zapcore.Lock(os.Stdout), zapcore.Lock(os.Stderr)
		var encoder = zapcore.NewConsoleEncoder(*config.Console.Encoder)
		cores = append(cores,
			zapcore.NewCore(encoder, wsErr, isErr),
			zapcore.NewCore(encoder, wsOut, isNotErr),
		)
	}
	if config.File.Enable {
		ws := zapcore.AddSync(getWriteSyncer(config))
		encoder := zapcore.NewJSONEncoder(*config.File.Encoder)
		cores = append(cores, zapcore.NewCore(encoder, ws, lv))
	}
	lg := zap.New(zapcore.NewTee(cores...), zapOptions...)
	return &Logger{
		cfg:           config,
		lg:            lg,
		SugaredLogger: lg.Sugar(),
		lv:            lv,
	}
}

func NewLogger(config *Config) *Logger {
	lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if err := lv.UnmarshalText([]byte(config.Level)); err != nil {
		panic(err)
	}
	return newLogger(&lv, config)
}
