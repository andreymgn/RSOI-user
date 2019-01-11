// Copyright (c) 2017 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracer

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

// NewRouter creates a new TracedRouter.
func NewRouter(tracer opentracing.Tracer) *TracedRouter {
	return &TracedRouter{
		Mux:    mux.NewRouter(),
		tracer: tracer,
	}
}

// TracedRouter is a wrapper around mux.Router that instruments handlers for tracing.
type TracedRouter struct {
	Mux    *mux.Router
	tracer opentracing.Tracer
}

// Handle implements mux.Router#Handle
func (tm *TracedRouter) Handle(pattern string, handler http.Handler) {
	middleware := nethttp.Middleware(
		tm.tracer,
		handler,
		nethttp.OperationNameFunc(func(r *http.Request) string {
			return "HTTP " + r.Method + " " + pattern
		}))
	tm.Mux.Handle(pattern, middleware)
}

// ServeHTTP implements http.ServeMux#ServeHTTP
func (tm *TracedRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tm.Mux.ServeHTTP(w, r)
}
