// Copyright 2015 go-dockerclient authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrNetworkAlreadyExists is the error returned by CreateNetwork when the
// network already exists.
var ErrNetworkAlreadyExists = errors.New("network already exists")

// Network represents a network.
//
// See https://goo.gl/1kmPKZ for more details.
type Network struct {
	Name       string
	ID         string `json:"Id"`
	Scope      string
	Driver     string
	Containers map[string]Endpoint
	Options    map[string]string
}

// Endpoint contains network resources allocated and used for a container in a network
//
// See https://goo.gl/1kmPKZ for more details.
type Endpoint struct {
	Name        string
	ID          string `json:"EndpointID"`
	MacAddress  string
	IPv4Address string
	IPv6Address string
}

// ListNetworks returns all networks.
//
// See https://goo.gl/1kmPKZ for more details.
func (c *Client) ListNetworks() ([]Network, error) {
	resp, err := c.do("GET", "/networks", doOptions{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var networks []Network
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, err
	}
	return networks, nil
}

// NetworkInfo returns information about a network by its ID.
//
// See https://goo.gl/1kmPKZ for more details.
func (c *Client) NetworkInfo(id string) (*Network, error) {
	path := "/networks/" + id
	resp, err := c.do("GET", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return nil, &NoSuchNetwork{ID: id}
		}
		return nil, err
	}
	defer resp.Body.Close()
	var network Network
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		return nil, err
	}
	return &network, nil
}

// CreateNetworkOptions specify parameters to the CreateNetwork function and
// (for now) is the expected body of the "create network" http request message
//
// See https://goo.gl/1kmPKZ for more details.
type CreateNetworkOptions struct {
	Name           string                 `json:"Name"`
	CheckDuplicate bool                   `json:"CheckDuplicate"`
	Driver         string                 `json:"Driver"`
	IPAM           IPAMOptions            `json:"IPAM"`
	Options        map[string]interface{} `json:"options"`
}

// IPAMOptions controls IP Address Management when creating a network
//
// See https://goo.gl/T8kRVH for more details.
type IPAMOptions struct {
	Driver string       `json:"Driver"`
	Config []IPAMConfig `json:"IPAMConfig"`
}

// IPAMConfig represents IPAM configurations
//
// See https://goo.gl/T8kRVH for more details.
type IPAMConfig struct {
	Subnet     string            `json:",omitempty"`
	IPRange    string            `json:",omitempty"`
	Gateway    string            `json:",omitempty"`
	AuxAddress map[string]string `json:"AuxiliaryAddresses,omitempty"`
}

// CreateNetwork creates a new network, returning the network instance,
// or an error in case of failure.
//
// See https://goo.gl/1kmPKZ for more details.
func (c *Client) CreateNetwork(opts CreateNetworkOptions) (*Network, error) {
	resp, err := c.do(
		"POST",
		"/networks/create",
		doOptions{
			data: opts,
		},
	)
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusConflict {
			return nil, ErrNetworkAlreadyExists
		}
		return nil, err
	}
	defer resp.Body.Close()

	type createNetworkResponse struct {
		ID string
	}
	var (
		network Network
		cnr     createNetworkResponse
	)
	if err := json.NewDecoder(resp.Body).Decode(&cnr); err != nil {
		return nil, err
	}

	network.Name = opts.Name
	network.ID = cnr.ID
	network.Driver = opts.Driver

	return &network, nil
}

// RemoveNetwork removes a network or an error in case of failure.
//
// See https://goo.gl/1kmPKZ for more details.
func (c *Client) RemoveNetwork(id string) error {
	resp, err := c.do("DELETE", "/networks/"+id, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchNetwork{ID: id}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// NoSuchNetwork is the error returned when a given network does not exist.
type NoSuchNetwork struct {
	ID string
}

func (err *NoSuchNetwork) Error() string {
	return fmt.Sprintf("No such network: %s", err.ID)
}
