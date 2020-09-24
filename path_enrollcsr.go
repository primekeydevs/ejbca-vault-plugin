/*
 * Copyright (c) 2019 Entrust Datacard Corporation.
 * All rights reserved.
 */

package main

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathEnrollCSR(b *backend) *framework.Path {
	ret := &framework.Path{
		Pattern: "enrollCSR/" + framework.GenericNameRegex("profileId") + "/?$",

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{Callback: b.opWriteEnrollCSR},
		},

		HelpSynopsis:    "CSR Enrollment",
		HelpDescription: "Enroll for a certificate from a CSR",
		Fields:          addCommonFields(map[string]*framework.FieldSchema{}),
	}

	ret.Fields["csr"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Default:     "",
		Description: "PEM-encoded CSR to send to the CA.",
		Required:    true,
	}

	return ret
}
