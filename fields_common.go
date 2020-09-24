/*
 * Copyright (c) 2019 Entrust Datacard Corporation.
 * All rights reserved.
 */
/*
 * Modified by PrimeKey Solutions AB, Copyright (c) 2020
 */

package main

import "github.com/hashicorp/vault/logical/framework"

func addCommonFields(fields map[string]*framework.FieldSchema) map[string]*framework.FieldSchema {


	fields["profileId"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Description: `The CA profile to use for enrollment`,
	}

	fields["serial"] = &framework.FieldSchema{
		Type: framework.TypeString,
		Description: `The certificate serial number to use for fetching
		the certificate and any private key.`,
	}

	fields["caname"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Description: `The CA Name to be used for enrollment.`,
	}
	fields["username"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Description: `The Username to be used for enrollment.`,
	}
	fields["certprofile"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Description: `The Certificate Profile to be used for enrollment.`,
	}
	fields["eeprofile"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Description: `The End Entity Profile to be used for enrollment.`,
	}

	return fields
}
