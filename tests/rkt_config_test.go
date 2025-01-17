// Copyright 2016 The rkt Authors
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

// +build host coreos src kvm

package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/rkt/rkt/tests/testutils"
)

type Basic struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type Oauth struct {
	Token string `json:"token"`
}

type Credentials struct {
	*Basic
	*Oauth
}

type Auth struct {
	Type        string      `json:"type,omitempty"`
	Registries  []string    `json:"registries,omitempty"`
	Domains     []string    `json:"domains,omitempty"`
	Credentials Credentials `json:"credentials"`
}

type Paths struct {
	Data         string `json:"data,omitempty"`
	Stage1Images string `json:"stage1-images,omitempty"`
}

type Stage1 struct {
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
	Location string `json:"location,omitempty"`
}

type cfg struct {
	*Auth
	*Paths
	*Stage1

	RktVersion string `json:"rktVersion"`
	RktKind    string `json:"rktKind"`
}

type stageCfg struct {
	Stage0 []cfg `json:"stage0"`
}

func (s stageCfg) search(x cfg) int {
	var c cfg
	var i int

	for i, c = range s.Stage0 {
		if reflect.DeepEqual(c, x) {
			return i
		}
	}

	return -1
}

func TestConfig(t *testing.T) {
	for i, tt := range []struct {
		configFunc func(*testutils.RktRunCtx)
		expected   []cfg
	}{
		{ // one auth domain
			configFunc: func(ctx *testutils.RktRunCtx) {
				writeConfig(t, authDir(ctx.LocalDir()), "basic.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "auth",
						Auth: &Auth{
							Type:    "basic",
							Domains: []string{"coreos.com"},
							Credentials: Credentials{
								Basic: &Basic{"user", "userPassword"},
							},
						},
					},
				))
			},
			expected: []cfg{
				{
					RktVersion: "v1",
					RktKind:    "auth",
					Auth: &Auth{
						Type:    "basic",
						Domains: []string{"coreos.com"},
						Credentials: Credentials{
							Basic: &Basic{"user", "userPassword"},
						},
					},
				},
			},
		},

		{ // one docker auth registry
			configFunc: func(ctx *testutils.RktRunCtx) {
				writeConfig(t, authDir(ctx.LocalDir()), "docker.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "dockerAuth",
						Auth: &Auth{
							Registries: []string{"quay.io"},
							Credentials: Credentials{
								Basic: &Basic{"docker", "dockerPassword"},
							},
						},
					},
				))
			},
			expected: []cfg{
				{
					RktVersion: "v1",
					RktKind:    "dockerAuth",
					Auth: &Auth{
						Registries: []string{"quay.io"},
						Credentials: Credentials{
							Basic: &Basic{"docker", "dockerPassword"},
						},
					},
				},
			},
		},

		{ // one paths entry
			configFunc: func(ctx *testutils.RktRunCtx) {
				writeConfig(t, pathsDir(ctx.LocalDir()), "paths.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "paths",
						Paths: &Paths{
							Data:         "/home/me/rkt/data",
							Stage1Images: "/home/me/rkt/stage1-images",
						},
					},
				))
			},
			expected: []cfg{
				{
					RktVersion: "v1",
					RktKind:    "paths",
					Paths: &Paths{
						Data:         "/home/me/rkt/data",
						Stage1Images: "/home/me/rkt/stage1-images",
					},
				},
			},
		},

		{ // overwrite paths entry in user dir
			configFunc: func(ctx *testutils.RktRunCtx) {
				writeConfig(t, pathsDir(ctx.SystemDir()), "paths.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "paths",
						Paths: &Paths{
							Data:         "/usr/lib/rkt/data",
							Stage1Images: "/usr/lib/rkt/stage1-images",
						},
					},
				))
				writeConfig(t, pathsDir(ctx.UserDir()), "paths.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "paths",
						Paths: &Paths{
							Data: "/home/me/rkt/data",
						},
					},
				))
			},
			expected: []cfg{
				{
					RktVersion: "v1",
					RktKind:    "paths",
					Paths: &Paths{
						Data:         "/home/me/rkt/data",
						Stage1Images: "/usr/lib/rkt/stage1-images",
					},
				},
			},
		},

		{ // one stage1 entry
			configFunc: func(ctx *testutils.RktRunCtx) {
				writeConfig(t, stage1Dir(ctx.LocalDir()), "stage1.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "stage1",
						Stage1: &Stage1{
							Name:     "example.com/rkt/stage1",
							Version:  "1.2.3",
							Location: "http://example.com/download/stage1.aci",
						},
					},
				))
			},
			expected: []cfg{
				{
					RktVersion: "v1",
					RktKind:    "stage1",
					Stage1: &Stage1{
						Name:     "example.com/rkt/stage1",
						Version:  "1.2.3",
						Location: "http://example.com/download/stage1.aci",
					},
				},
			},
		},

		{ // overwrite stage1 entry
			configFunc: func(ctx *testutils.RktRunCtx) {
				writeConfig(t, stage1Dir(ctx.LocalDir()), "stage1.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "stage1",
						Stage1: &Stage1{
							Name:     "example.com/rkt/stage1",
							Version:  "1.2.3",
							Location: "http://example.com/download/stage1.aci",
						},
					},
				))
				writeConfig(t, stage1Dir(ctx.UserDir()), "stage1.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "stage1",
						Stage1: &Stage1{
							Location: "http://localhost:8080/stage1.aci",
						},
					},
				))
			},
			expected: []cfg{
				{
					RktVersion: "v1",
					RktKind:    "stage1",
					Stage1: &Stage1{
						Name:     "example.com/rkt/stage1",
						Version:  "1.2.3",
						Location: "http://localhost:8080/stage1.aci",
					},
				},
			},
		},

		{ // two auth domains, two docker registries
			configFunc: func(ctx *testutils.RktRunCtx) {
				writeConfig(t, authDir(ctx.UserDir()), "basic.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "auth",
						Auth: &Auth{
							Type:    "basic",
							Domains: []string{"tectonic.com", "coreos.com"},
							Credentials: Credentials{
								Basic: &Basic{"user", "userPassword"},
							},
						},
					},
				))
				writeConfig(t, authDir(ctx.UserDir()), "docker.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "dockerAuth",
						Auth: &Auth{
							Registries: []string{"quay.io", "gcr.io"},
							Credentials: Credentials{
								Basic: &Basic{"docker", "dockerPassword"},
							},
						},
					},
				))
			},
			expected: []cfg{
				{
					RktVersion: "v1",
					RktKind:    "auth",
					Auth: &Auth{
						Type:    "basic",
						Domains: []string{"tectonic.com"},
						Credentials: Credentials{
							Basic: &Basic{"user", "userPassword"},
						},
					},
				},
				{
					RktVersion: "v1",
					RktKind:    "auth",
					Auth: &Auth{
						Type:    "basic",
						Domains: []string{"coreos.com"},
						Credentials: Credentials{
							Basic: &Basic{"user", "userPassword"},
						},
					},
				},
				{
					RktVersion: "v1",
					RktKind:    "dockerAuth",
					Auth: &Auth{
						Registries: []string{"quay.io"},
						Credentials: Credentials{
							Basic: &Basic{"docker", "dockerPassword"},
						},
					},
				},
				{
					RktVersion: "v1",
					RktKind:    "dockerAuth",
					Auth: &Auth{
						Registries: []string{"gcr.io"},
						Credentials: Credentials{
							Basic: &Basic{"docker", "dockerPassword"},
						},
					},
				},
			},
		},

		{ // overwrite one auth domain, one dockerAuth registry in user dir
			configFunc: func(ctx *testutils.RktRunCtx) {
				writeConfig(t, authDir(ctx.SystemDir()), "basic.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "auth",
						Auth: &Auth{
							Type:    "basic",
							Domains: []string{"tectonic.com", "coreos.com"},
							Credentials: Credentials{
								Basic: &Basic{"user", "userPassword"},
							},
						},
					},
				))

				writeConfig(t, authDir(ctx.UserDir()), "basic.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "auth",
						Auth: &Auth{
							Type:    "oauth",
							Domains: []string{"tectonic.com"},
							Credentials: Credentials{
								Oauth: &Oauth{"someToken"},
							},
						},
					},
				))

				writeConfig(t, authDir(ctx.SystemDir()), "docker.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "dockerAuth",
						Auth: &Auth{
							Registries: []string{"quay.io", "gcr.io"},
							Credentials: Credentials{
								Basic: &Basic{"docker", "dockerPassword"},
							},
						},
					},
				))

				writeConfig(t, authDir(ctx.UserDir()), "docker.json", mustMarshalJSON(
					cfg{
						RktVersion: "v1",
						RktKind:    "dockerAuth",
						Auth: &Auth{
							Registries: []string{"gcr.io"},
							Credentials: Credentials{
								Basic: &Basic{"over", "written"},
							},
						},
					},
				))
			},
			expected: []cfg{
				{
					RktVersion: "v1",
					RktKind:    "auth",
					Auth: &Auth{
						Type:    "oauth",
						Domains: []string{"tectonic.com"},
						Credentials: Credentials{
							Oauth: &Oauth{"someToken"},
						},
					},
				},
				{
					RktVersion: "v1",
					RktKind:    "auth",
					Auth: &Auth{
						Type:    "basic",
						Domains: []string{"coreos.com"},
						Credentials: Credentials{
							Basic: &Basic{"user", "userPassword"},
						},
					},
				},
				{
					RktVersion: "v1",
					RktKind:    "dockerAuth",
					Auth: &Auth{
						Registries: []string{"quay.io"},
						Credentials: Credentials{
							Basic: &Basic{"docker", "dockerPassword"},
						},
					},
				},
				{
					RktVersion: "v1",
					RktKind:    "dockerAuth",
					Auth: &Auth{
						Registries: []string{"gcr.io"},
						Credentials: Credentials{
							Basic: &Basic{"over", "written"},
						},
					},
				},
			},
		},
	} {
		func() {
			ctx := testutils.NewRktRunCtx()
			defer ctx.Cleanup()

			tt.configFunc(ctx)

			rktCmd := ctx.Cmd() + " --debug --insecure-options=image --pretty-print=false config"
			nobodyUid, _ := testutils.GetUnprivilegedUidGid()
			out, status := runRkt(t, rktCmd, nobodyUid, 0)

			if status != 0 {
				panic(fmt.Errorf("test %d: expected exit status code 0, got %d", i, status))
			}

			var got stageCfg
			if err := json.Unmarshal([]byte(out), &got); err != nil {
				panic(err)
			}

			for ii, e := range tt.expected {
				if got.search(e) < 0 {
					t.Errorf("test %d, expected to find config %d but didn't", i, ii)
				}
			}
		}()
	}
}

func mustMarshalJSON(data interface{}) string {
	buf, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return string(buf)
}
