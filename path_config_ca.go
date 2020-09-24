/*
 * Copyright (c) 2019 Entrust Datacard Corporation.
 * All rights reserved.
 */
/*
 * Modified by PrimeKey Solutions AB, Copyright (c) 2020
 */

package main

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathConfig(b *backend) *framework.Path {
	ret := &framework.Path{
		Pattern: "config/" + framework.GenericNameRegex("profileId"),

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation:   &framework.PathOperation{Callback: b.opReadConfigCA},
			logical.UpdateOperation: &framework.PathOperation{Callback: b.opWriteConfigCA},
		},

		HelpSynopsis:    "EJBCA Configuration",
		HelpDescription: "Configures connection parameters including client cert and key.",
		Fields:          map[string]*framework.FieldSchema{},
	}

	ret.Fields["pem_bundle"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Default:     "",
		Description: `PEM encoded client certificate and key.`,
		Required:    true,
	}

	ret.Fields["profileId"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Default:     "",
		Description: `Profile id of the Vault EJBCA Profile to use for enrollment`,
		Required:    true,
	}

	ret.Fields["url"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Default:     "",
		Description: `URL for EJBCA REST end point including context path`,
		Required:    true,
	}

	ret.Fields["cacerts"] = &framework.FieldSchema{
		Type:    framework.TypeString,
		Default: "",
		Description: "PEM encoded TLS CA certificate chain.",
	}

	ret.Fields["caname"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Default:     "",
		Description: `The CA Name to be used for enrollment`,
		Required:    true,
	}

	ret.Fields["certprofile"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Default:     "",
		Description: `The Certificate Profile to be used for enrollment`,
		Required:    true,
	}

	ret.Fields["eeprofile"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Default:     "",
		Description: `The End Entity Profile to be used for enrollment`,
		Required:    true,
	}


	return ret
}
