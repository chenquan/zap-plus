/*
 *    Copyright 2022 chenquan
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package config

import "gopkg.in/natefinch/lumberjack.v2"

type (
	// Config represents a configuration.
	Config struct {
		Trace
		// Level is the log level. Possible values: debug, info, warn, error
		Level string `yaml:"level" validate:"eq=debug|eq=info|eq=warn|eq=error"`
		// Format is the log format. Possible values: json, text
		Format string `yaml:"format" validate:"eq=text|eq=json"`
		Mode   string `yaml:"mode" validate:"eq=file|eq=console|eq="`
		lumberjack.Logger
	}
	// Trace represents a tracing configuration.
	Trace struct {
		Name     string  `yaml:"name"`
		Endpoint string  `yaml:"endpoint"`
		Sampler  float64 `yaml:"sampler"`
		Batcher  string  `yaml:"batcher"  validate:"eq=jaeger|eq=zipkin|eq="` // jaeger|zipkin
	}
)
