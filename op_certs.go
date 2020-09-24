/*
 * Copyright (c) 2020 Entrust Datacard Corporation.
 * All rights reserved.
 */
/*
 * Modified by PrimeKey Solutions AB, Copyright (c) 2020
 */

package main

import (
	"context"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func (b *backend) opListCerts(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {

	profileId := data.Get("profileId").(string)

	entries, err := req.Storage.List(ctx, "issued/"+profileId+"/")
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(entries), nil
}

func (b *backend) opReadCert(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	profileId := data.Get("profileId").(string)
	serial := data.Get("serial").(string)

	storageEntry, err := req.Storage.Get(ctx, "issued/"+profileId+"/"+serial)
	if err != nil {
		return logical.ErrorResponse("Could not read certificate with the serial number: " + serial), err
	}
	if storageEntry == nil {
		return logical.ErrorResponse("Could not find certificate with the serial number: " + serial), nil
	}

	var rawData map[string]interface{}
	err = storageEntry.DecodeJSON(&rawData)

	if err != nil {
		return logical.ErrorResponse("JSON decoding failed for certificate: " + serial), err
	}

	resp := &logical.Response{
		Data: rawData,
	}

	return resp, nil
}

