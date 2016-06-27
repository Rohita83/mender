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
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/mendersoftware/log"
	"github.com/pkg/errors"
)

const (
	minimumImageSize int64 = 4096 //kB
)

type Updater interface {
	GetScheduledUpdate(server string, deviceID string) (interface{}, error)
	FetchUpdate(url string) (io.ReadCloser, int64, error)
}

type UpdateClient struct {
	client       *http.Client
	minImageSize int64
}

func NewUpdateClient(conf httpsClientConfig) (*UpdateClient, error) {
	client, err := NewHttpClient(conf)
	if client == nil {
		return nil, errors.Wrap(err, "failed to create updater HTTP client")
	}

	up := UpdateClient{
		client:       client,
		minImageSize: minimumImageSize,
	}
	return &up, nil
}

func (u *UpdateClient) GetScheduledUpdate(server string, deviceID string) (interface{}, error) {
	return u.getUpdateInfo(processUpdateResponse, server, deviceID)
}

func (u *UpdateClient) getUpdateInfo(process RequestProcessingFunc, server string,
	deviceID string) (interface{}, error) {
	req, err := makeUpdateCheckRequest(server, deviceID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create update check request")
	}

	r, err := u.client.Do(req)
	if err != nil {
		log.Debug("Sending request error: ", err)
		return nil, errors.Wrapf(err, "update check request failed")
	}

	defer r.Body.Close()

	return process(r)
}

// Returns a byte stream which is a download of the given link.
func (u *UpdateClient) FetchUpdate(url string) (io.ReadCloser, int64, error) {
	req, err := makeUpdateFetchRequest(url)
	if err != nil {
		return nil, -1, errors.Wrapf(err, "failed to create update fetch request")
	}

	r, err := u.client.Do(req)
	if err != nil {
		log.Error("Can not fetch update image: ", err)
		return nil, -1, errors.Wrapf(err, "update fetch request failed")
	}

	log.Debugf("Received fetch update response %v+", r)

	if r.StatusCode != http.StatusOK {
		log.Errorf("Error fetching shcheduled update info: code (%d)", r.StatusCode)
		return nil, -1, errors.New("Error receiving scheduled update information.")
	}

	if r.ContentLength < 0 {
		return nil, -1, errors.New("Will not continue with unknown image size.")
	} else if r.ContentLength < u.minImageSize {
		log.Errorf("Image smaller than expected. Expected: %d, received: %d", u.minImageSize, r.ContentLength)
		return nil, -1, errors.New("Image size is smaller than expected. Aborting.")
	}

	return r.Body, r.ContentLength, nil
}

// possible API responses received for update request
const (
	updateResponseHaveUpdate = 200
	updateResponseNoUpdates  = 204
	updateResponseError      = 404
)

// have update for the client
type UpdateResponse struct {
	Image struct {
		URI      string
		Checksum string
		YoctoID  string `json:"yocto_id"`
		ID       string
	}
	ID string
}

func validateGetUpdate(update UpdateResponse) error {
	// check if we have JSON data correctky decoded
	if update.ID != "" && update.Image.ID != "" &&
		update.Image.URI != "" && update.Image.YoctoID != "" {
		log.Info("Correct request for getting image from: " + update.Image.URI)
		return nil
	}
	return errors.New("Missing parameters in encoded JSON response")
}

func processUpdateResponse(response *http.Response) (interface{}, error) {
	log.Debug("Received response:", response.Status)

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case updateResponseHaveUpdate:
		log.Debug("Have update available")

		var data UpdateResponse
		if err := json.Unmarshal(respBody, &data); err != nil {
			return nil, errors.Wrapf(err, "failed to parse response")
		}

		if err := validateGetUpdate(data); err != nil {
			return nil, err
		}

		return data, nil

	case updateResponseNoUpdates:
		log.Debug("No update available")
		return nil, nil

	case updateResponseError:
		return nil, errors.New("Client not authorized to get update schedule.")

	default:
		return nil, errors.New("Invalid response received from server")
	}
}

func makeUpdateCheckRequest(server, deviceID string) (*http.Request, error) {
	url := buildApiURL(server, "/deployments/devices/"+deviceID+"/update")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func makeUpdateFetchRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}
