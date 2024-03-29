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
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"go.uber.org/zap/zapcore"
)

func Test_Logger(t *testing.T) {
	lg := (&Config{Console: struct {
		Enable  bool
		Encoder *zapcore.EncoderConfig
	}{Enable: true}}).Build()
	var dd = struct {
		A string
		B int
	}{"a", 2}
	lg.Write(logger.LvDebug, "fdaf", "k1", 1, "k2", dd)
	//zap.L().Debug("fdasfadsf", zap.String("fdsf", "fdaf"))
	//lg.Fatalf("fault test")
	logger.SetWriter(lg)
	logger.DebugField("fdafasdf", logger.String("test", "fdas"))
	//h := logger.WithFields(logger.String("plugins", "zap"))
	//h.DebugField("fdafdsaf", logger.String("k1", "k2"))
}
