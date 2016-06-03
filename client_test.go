// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpClient(t *testing.T) {
	cl, err := NewHttpClient(
		httpsClientConfig{"client.crt", "client.key", "server.crt", true},
	)
	assert.NotNil(t, cl)
	// assert.NotNil(t, cl.Transport.TLSClientConfig)

	// no https config, we should obtain a httpClient
	cl, err = NewHttpClient(httpsClientConfig{})
	assert.NotNil(t, cl)
	// assert.Nil(t, cl.Transport.TLSClientConfig)

	// incomplete config should yield an error
	cl, err = NewHttpClient(
		httpsClientConfig{"foobar", "client.key", "", true},
	)
	assert.Nil(t, cl)
	assert.NotNil(t, err)
}

func TestHttpClientUrl(t *testing.T) {
	u := buildURL("https://foo.bar")
	assert.Equal(t, "https://foo.bar", u)

	u = buildURL("http://foo.bar")
	assert.Equal(t, "http://foo.bar", u)

	u = buildURL("foo.bar")
	assert.Equal(t, "https://foo.bar", u)

}
