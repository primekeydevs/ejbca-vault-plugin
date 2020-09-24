/*
 * Copyright (c) 2019 Entrust Datacard Corporation.
 * All rights reserved.
 */
/*
 * Modified by PrimeKey Solutions AB, Copyright (c) 2020
 */

package main

import (
	"log"
	"context"
	"fmt"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type EJBCAConfigProfile struct {
	PEMBundle   string
	URL         string
	CACerts     string
	CAName	    string
	SubjectDn   string
	CertProfile string
	EEProfile   string
}

type CAsResponse struct {
        CAResponse []CAResponse `json:"certificate_authorities"`
}

type CAResponse struct {
        Id                          int                          `json:"id"`
        Name                        string                       `json:"name"`
        SubjectDn                   string                       `json:"subject_dn"`
        IssuerDn                    string                       `json:"issuer_dn"`
        ExpirationDate              string                       `json:"expiration_date"`
}

func (b *backend) opWriteConfigCA(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	profileId := data.Get("profileId").(string)

	certPem := data.Get("pem_bundle").(string)
	url := data.Get("url").(string)
	caCertPem := data.Get("cacerts").(string)
	caName := data.Get("caname").(string)
	certProfile := data.Get("certprofile").(string)
	eeProfile := data.Get("eeprofile").(string)

	if b.Logger().IsDebug() {
		b.Logger().Debug("profileId: " + profileId)
		b.Logger().Debug("certPem: " + certPem)
		b.Logger().Debug("url: " + url)
		b.Logger().Debug("caCertPem: " + caCertPem)
		b.Logger().Debug("caName: " + caName)
		b.Logger().Debug("certProfile: " + certProfile)
		b.Logger().Debug("eeProfile: " + eeProfile)
	}


	if len(certPem) == 0 {
		return logical.ErrorResponse("must provide PEM encoded client certificate"), nil
	}
	if len(profileId) == 0 {
		return logical.ErrorResponse("must provide Profile identifier"), nil
	}
	if len(url) == 0 {
		return logical.ErrorResponse("must provide server URL"), nil
	}
	if len(caCertPem) == 0 {
		return logical.ErrorResponse("must provide server CA certificate chain"), nil
	}
	if len(caName) == 0 {
		return logical.ErrorResponse("must provide CA name"), nil
	}
	if len(certProfile) == 0 {
		return logical.ErrorResponse("must provide certificate profile"), nil
	}
	if len(eeProfile) == 0 {
		return logical.ErrorResponse("must provide end entity profile"), nil
	}

	configCa := &EJBCAConfigProfile{
		certPem,
		url,
		caCertPem,
		caName,
		"",
		certProfile,
		eeProfile,
	}

	// Verify connection to EJBCA

	tlsClientConfig, err := getTLSConfig(ctx, req, configCa)
	if err != nil {
		log.Println("Error retrieving TLS configuration: %w", err)
		return logical.ErrorResponse("Error retrieving TLS configuration"), err
	}
	subjectdn, err := readCAsAndGetDN(tlsClientConfig, caName, url)
	if err != nil {
		log.Println("Error reading CAs from CA server: %w", err)
		return logical.ErrorResponse("Error reading CAs from CA server"), err
	}
	configCa.SubjectDn = subjectdn;

	// Store configuration
	storageEntry, err := logical.StorageEntryJSON("config/"+profileId, configCa)

	if err != nil {
		return logical.ErrorResponse("Error creating config storage entry"), err
	}

	err = req.Storage.Put(ctx, storageEntry)
	if err != nil {
		return logical.ErrorResponse("Could not store configuration"), err
	}

	respData := map[string]interface{}{
		"* Message": "Configuration successful",
		"ProfileId":    profileId,
		"URL":     url,
		"CAName":     caName,
		"Certificate Profile":     certProfile,
		"End Entity Profile":     eeProfile,
	}

	return &logical.Response{
		Data: respData,
	}, nil
}

func readCAsAndGetDN(tlsClientConfig *tls.Config, caName string, url string) (string, error) {
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsClientConfig,
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Get(url + "/ca")
	if err != nil {
		return "", fmt.Errorf("Error response: %w", err)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("CA response could not be read: %w", err)
	}

	if resp.StatusCode != 200 {
		var errorResponse *ErrorResponse
		err := json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			return "", fmt.Errorf("CA error response could not be parsed (%d)", resp.StatusCode)
		}
		return "", fmt.Errorf("Error from server: %s (%d)", errorResponse.Message, resp.StatusCode)
	}

	// Go through JSON and verify that our caname is in the list, this proves that "I" have the access rights needed to see this CA at least
	// That is a minimum set of verification that my access rights are ok and that this config will work
	// Return CAs DN to store in configuration
	var casResp *CAsResponse
	log.Println("response body: " + string(responseBody))
	err = json.Unmarshal(responseBody, &casResp)
	if err != nil {
		return "", fmt.Errorf("CA enrollment response could not be parsed: %w", err)
	}
	for i := 0; i < len(casResp.CAResponse); i++ {
		//log.Println("caResp: " + string(casResp.CAResponse[i].Name) + ": " + string(casResp.CAResponse[i].SubjectDn))
		if (casResp.CAResponse[i].Name == caName) {
			log.Println("Found CA entry with subjectdn: " + string(casResp.CAResponse[i].SubjectDn))
			return casResp.CAResponse[i].SubjectDn, nil
		}
	}

	return "", fmt.Errorf("Could not find CA on server", err)
}

func (b *backend) opReadConfigCA(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	profileId := data.Get("profileId").(string)

	storageEntry, err := req.Storage.Get(ctx, "config/"+profileId)
	if err != nil {
		return logical.ErrorResponse("Could not read configuration"), err
	}

	var rawData map[string]interface{}
	err = storageEntry.DecodeJSON(&rawData)

	if err != nil {
		return logical.ErrorResponse("JSON decoding failed"), err
	}

	resp := &logical.Response{
		Data: rawData,
	}

	return resp, nil
}
