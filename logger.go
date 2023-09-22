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
	logger.RegisterWriterBuilder("zap", func() logger.Writer {
		cfg := &Config{}
		if err := config.Get("zap").Scan(cfg); err != nil {
			logger.FatalField("fault to load zap config", logger.Err(err))
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
	cfg   *Config
	sugar *zap.SugaredLogger
	lv    *zap.AtomicLevel
}

func (lg *Logger) Write(lv logger.Level, msg string, kvs ...interface{}) {
	switch lv {
	case logger.LvDebug:
		lg.sugar.Debugw(msg, kvs...)
	case logger.LvInfo:
		lg.sugar.Infow(msg, kvs...)
	case logger.LvWarn:
		lg.sugar.Warnw(msg, kvs...)
	case logger.LvError:
		lg.sugar.Errorw(msg, kvs...)
	case logger.LvFault:
		lg.sugar.Fatalw(msg, kvs...)
	}
}

var _ logger.Writer = (*Logger)(nil)

func newLogger(lv *zap.AtomicLevel, cfg *Config) *Logger {
	zapOptions := make([]zap.Option, 0)
	zapOptions = append(zapOptions, zap.AddStacktrace(zap.PanicLevel))
	if cfg.AddCaller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(cfg.CallerSkip))
	}
	cores := make([]zapcore.Core, 0, 1)
	isErr := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel && lv.Level() <= zapcore.ErrorLevel
	})
	isNotErr := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lv.Level() <= lvl
	})

	if cfg.Console.Enable {
		var wsOut, wsErr = zapcore.Lock(os.Stdout), zapcore.Lock(os.Stderr)
		var encoder = zapcore.NewConsoleEncoder(*cfg.Console.Encoder)
		cores = append(cores,
			zapcore.NewCore(encoder, wsErr, isErr),
			zapcore.NewCore(encoder, wsOut, isNotErr),
		)
	}
	if cfg.File.Enable {
		ws := zapcore.AddSync(getWriteSyncer(cfg))
		encoder := zapcore.NewJSONEncoder(*cfg.File.Encoder)
		cores = append(cores, zapcore.NewCore(encoder, ws, lv))
	}
	lg := zap.New(zapcore.NewTee(cores...), zapOptions...)
	l := &Logger{
		cfg:   cfg,
		sugar: lg.Sugar(),
		lv:    lv,
	}
	return l
}

func NewLogger(cfg *Config) *Logger {
	lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if err := lv.UnmarshalText([]byte(cfg.Level)); err != nil {
		panic(err)
	}
	lg := newLogger(&lv, cfg)
	return lg
}
