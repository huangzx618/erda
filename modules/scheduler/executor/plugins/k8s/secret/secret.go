// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// Package secret manipulates the k8s api of secret object
package secret

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

// Secret is the object to encapsulate docker
type Secret struct {
	addr   string
	client *httpclient.HTTPClient
}

// Option configures an Secret
type Option func(*Secret)

// New news an Secret
func New(options ...Option) *Secret {
	ns := &Secret{}

	for _, op := range options {
		op(ns)
	}

	return ns
}

// WithCompleteParams provides an Option
func WithCompleteParams(addr string, client *httpclient.HTTPClient) Option {
	return func(s *Secret) {
		s.addr = addr
		s.client = client
	}
}

// List list k8s secrets in 'namespace'
func (p *Secret) List(namespace string) (*apiv1.SecretList, error) {
	var b bytes.Buffer
	path := strutil.Concat("/api/v1/namespaces/", namespace, "/secrets")
	resp, err := p.client.Get(p.addr).
		Path(path).
		Do().
		Body(&b)
	if err != nil {
		return nil, errors.Errorf("failed to list secrets, namespace: %s, err: %v", namespace, err)
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("failed to list secrets, namespace: %s, statuscode: %d, body: %v",
			namespace, resp.StatusCode(), b.String())
	}
	secrets := &apiv1.SecretList{}
	if err := json.NewDecoder(&b).Decode(secrets); err != nil {
		return nil, err
	}
	return secrets, nil
}

// Get gets a k8s secret
func (p *Secret) Get(namespace, name string) (*apiv1.Secret, error) {
	var b bytes.Buffer
	path := strutil.Concat("/api/v1/namespaces/", namespace, "/secrets/", name)

	resp, err := p.client.Get(p.addr).
		Path(path).
		Do().
		Body(&b)

	if err != nil {
		return nil, errors.Errorf("failed to get secret, name: %s, (%v)", name, err)
	}
	if !resp.IsOK() {
		if resp.IsNotfound() {
			return nil, errors.Errorf("not found")
		}
		return nil, errors.Errorf("failed to get secret, name: %s, statuscode: %v, body: %v",
			name, resp.StatusCode(), b.String())
	}
	secret := &apiv1.Secret{}
	if err := json.NewDecoder(&b).Decode(secret); err != nil {
		return nil, err
	}
	return secret, nil
}

// Create creates a k8s ingress object
func (p *Secret) Create(secret *apiv1.Secret) error {
	var b bytes.Buffer
	path := strutil.Concat("/api/v1/namespaces/", secret.Namespace, "/secrets")

	resp, err := p.client.Post(p.addr).
		Path(path).
		JSONBody(secret).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to create secret, name: %s, (%v)", secret.Name, err)
	}

	if !resp.IsOK() {
		return errors.Errorf("failed to create secret, statuscode: %v, body: %v", resp.StatusCode(), b.String())
	}
	return nil
}

func (p *Secret) CreateIfNotExist(secret *apiv1.Secret) error {
	_, err := p.Get(secret.Namespace, secret.Name)
	if err == nil {
		return nil
	}
	if err.Error() == "not found" {
		return p.Create(secret)
	}
	return err
}
