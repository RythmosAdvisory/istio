// Copyright 2020 Istio Authors
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

package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"

	meshconfig "istio.io/api/mesh/v1alpha1"

	"istio.io/istio/pkg/config/mesh"
)

func TestGetMeshConfig(t *testing.T) {
	overrides := `
defaultConfig:
  discoveryAddress: foo:123
  proxyMetadata:
    SOME: setting
  drainDuration: 1s`
	overridesExpected := func() meshconfig.ProxyConfig {
		m := mesh.DefaultProxyConfig()
		m.DiscoveryAddress = "foo:123"
		m.ProxyMetadata = map[string]string{"SOME": "setting"}
		m.DrainDuration = types.DurationProto(time.Second)
		return m
	}()
	cases := []struct {
		name        string
		annotation  string
		environment string
		file        string
		expect      meshconfig.ProxyConfig
	}{
		{
			name:   "Defaults",
			expect: mesh.DefaultProxyConfig(),
		},
		{
			name: "Annotation Override",
			annotation: `discoveryAddress: foo:123
proxyMetadata:
  SOME: setting
drainDuration: 1s`,
			expect: overridesExpected,
		},
		{
			name:   "File Override",
			file:   overrides,
			expect: overridesExpected,
		},
		{
			name:        "Environment Override",
			environment: overrides,
			expect:      overridesExpected,
		},
		{
			// Hopefully no one actually has all three of these set in a real system, but we will still
			// test them all together.
			name: "Multiple Override",
			// Order is file < env < annotation
			file: `
defaultConfig:
  discoveryAddress: file:123
  proxyMetadata:
    SOME: setting
  drainDuration: 1s`,
			environment: `
defaultConfig:
  discoveryAddress: environment:123
  proxyMetadata:
    OTHER: option`,
			annotation: `
discoveryAddress: annotation:123
proxyMetadata:
  ANNOTATION: something
drainDuration: 5s
`,
			expect: func() meshconfig.ProxyConfig {
				m := mesh.DefaultProxyConfig()
				m.DiscoveryAddress = "annotation:123"
				m.ProxyMetadata = map[string]string{"ANNOTATION": "something"}
				m.DrainDuration = types.DurationProto(5 * time.Second)
				return m
			}(),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			meshConfig = tt.environment
			got, err := getMeshConfig(tt.file, tt.annotation)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(*got.DefaultConfig, tt.expect) {
				t.Fatalf("got \n%v expected \n%v", *got.DefaultConfig, tt.expect)
			}
		})
	}
}
