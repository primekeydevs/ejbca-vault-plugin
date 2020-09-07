![PrimeKey](primekey_logo.png)

# Community supported 
We welcome contributions.

ejbca-vault-plugin is open source and community supported, meaning that there is **no SLA** applicable for this tool.

To report a problem or suggest a new feature, use the **[Issues](../../issues)** tab. If you want to contribute actual bug fixes or proposed enhancements, use the **[Pull requests](../../pulls)** tab.

# License
Apache License 2.0.

# EJBCA Vault plugin

HashiCorp Vault is a popular product to manage secrets, and when using microservices at scale, there are many services and thus many secrets to manage. HashiCorp Vault comes with a built in Certification Authority, but using that standalone will create a separate PKI which is not connected to the corporate PKI, which is not desired in many organization as it will not meet regulatory or other security requirements. In order to incorporate Vault PKI into a controlled, corporately managed PKI there are (at least) two different ways.

* Plug into your own issuing CA using a Vaule secrets plugin, 
* or use the Vault provisions to root to an external CA rather than self-sign its own CA certificate

The EJBCA Vault secrets plugin, and a drop in replacement of Vault's built in PKI, that allows you to plug Vault into your issuing CA. You will use Vault to issue certificates, just as you would with the Vault built in CA, but the issuance come from EJBCA. 


## Security
The EJBCA Vault plugin uses the [EJBCA REST API](https://doc.primekey.com/ejbca/ejbca-operations/ejbca-ca-concept-guide/protocols/ejbca-rest-interface) to communicate with EJBCA, so the EJBCA instance can be anywhere where is reachable with https connections. The EJBCA REST API uses mutually authenticated TLS connections (https/mTLS). To EJBCA Vault acts as a PKI Registration Authority, and using the [role based access control](https://doc.primekey.com/ejbca/ejbca-operations/ejbca-ca-concept-guide/roles-and-access-rules) in EJBCA the capabilities that Vault has in the EJBCA instance can be controled fine grained. Vault will not be able to issue certificates from other CAs managed than the ones allowed in the EJBCA instance. 

## EJBCA Requirements
To issue certificates from EJBCA, through Vault, has the same requirements as issuing through any RA.
* A [CA](https://doc.primekey.com/ejbca/ejbca-operations/ejbca-ca-concept-guide/certificate-authority-overview) in EJBCA, that can issue the desired type of certificates, usign the desired algorithm
* A [Certificate Profile](https://doc.primekey.com/ejbca/ejbca-operations/ejbca-ca-concept-guide/certificate-profiles-overview), configured to issue the desired type of certificates from the CA
* An [End Entity Profile](https://doc.primekey.com/ejbca/ejbca-operations/ejbca-ca-concept-guide/end-entities-overview/end-entity-profiles-overview), configured to issue certificates with the desired Subject fields from the CA, using the Certificate Profile
* A [Role](https://doc.primekey.com/ejbca/ejbca-operations/ejbca-ca-concept-guide/roles-and-access-rules) with Access Rules, describing what Vault is allowed to issue
* A Role Member, which is Vault, beloning to the corresponding Role
* A Client certificate and proivate key, used to authenticate the Role Member over mutually authenticated TLS

## Build

The plugin is written in [Go](https://golang.org/dl/), as that is the language chosen by Vault.

Installing Go on an Ubuntu Linux systems can easily be done by:
```
> sudo snap install go
```
After that you can build the plugin.

Build the plugin, using Go, with the following command:

```
> go build -o out/ejbca-vault-plugin
```

This will build the plugin and store the resulting executable as `out/ejbca_vault_plugin`


## Installation

The EJBCA Vault plugin is installed like other [Vault plugins](https://www.vaultproject.io/docs/internals/plugins.html). The plugin executable must reside in the Vault plugin directory, and the SHA256 hash of the plugin executable must be known to the administrator installing the plugin.

### Start Vault
Manaing Vault in a production enviroinment is outside the scope of this instruction, so we start a development server with an appointed plugin directory (straight into our build directory), and a defined root token, change this root token to something random of your own:

```
vault server -dev -dev-root-token-id=gUgvfVcVzdKH -dev-plugin-dir=/home/user/git/ejbca-vault-plugin/out -log-level=debug
```

To use the vault CLI you need to set environment variables:

```
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='gUgvfVcVzdKH'
```
Now you are ready to run Vault CLI commands.

### Register and enable plugin

To register the plugin you must first get the SHA256 hash of the plugin executable, after which you can register the plugin with Vault:

```
SHA256=$(sha256sum /home/user/git/ejbca-vault-plugin/out/ejbca_vault_plugin| cut -d' ' -f1)
vault write sys/plugins/catalog/ejbca_vault_plugin sha256="${SHA256}" command="ejbca-vault-plugin"
```

`"${SHA256}"` is the SHA256 of the Vault plugin executable, and can be specified manually on the command line if you prefer that instead of using a shell variable. 

After registration you can enable the plugin, which gives a path for further commands to vault to use the plugin:

```
vault secrets enable -path=ejbca ejbca-vault-plugin
```

### Disable plugin
If you want to disable the plugin, without re-registering it, that can easily be done:

```
vault secrets disable ejbca
```

Note that disabling the plugin will remove all the certificates stored, i.e. the 'list' command below will return empty. When upgrading the plugin in production another method is recommended.

### Rebuild and re-deploy
Example command how to perform a full rebuild - redeploy cycle (for example when modifying the plugin) can be found in the `redeploy.sh` file in this repo.

Note that disabling the plugin will remove all the certificates stored, i.e. the 'list' command below will return empty. When upgrading the plugin in production another method is recommended.

## Configuration

Configure an EJBCA secret that maps a name in Vault to connection and profile settings for enrolling certificate using EJBCA. To enable the connection between Vault and EJBCA you obtain a client certificate and private key for EJBCA, from EJBCA Admins.
You can maintain multiple configurations of the plugin for issuing different types of certificates, or from different CAs.

### Instance Configuration

Configure the plugin by writing to the `/config/profile` endpoint where {profile} is an arbitrary label of the specific certificates you want to issue using Vault. You can enable multiple Vault profiles that issues different types of certificates. From different CAs, Certificate- and End Entity Profiles in EJBCA, and also from different EJBCA servers. 

The connection to EJBCA requires two properties:
* **pem_bundle** - The certificate and private key, in PEM format, to authenticate to the EJBCA REST API end-point
* **url** - The URL for the server (CA or RA) including the context path
* **cacerts** - The complete certificate chain for the TLS CA in PEM format

The configuration of CA and profiles to issue from requires three properties:
* **caname** - The name of the Certification Authority in EJBCA to issue from
* **certprofile** - The Certificate Profile to issue from
* **eeprofile** - The End Entity Profile to issue from

#### Example

```
vault write ejbca/config/TLSServer pem_bundle=@admin-bundle.pem url=https://ejbca.example.com:8443/ejbca/ejbca-rest-api/v1 cacerts=@admin-TLS-chain.pem caname=TLSAssuredCA certprofile=TLSServer eeprofile=InternalTLSServer
```

The above write operation will connect to EJBCA for a connecting test and set the configuration properties for issuing. 

The read operation will display these configuration properties (as written to JSON in Vault).

>`vault read ejbca/config/TLSServer`

>`vault read -field=CACerts ejbca/config/TLSServer`

>`vault read -field=PEMBundle ejbca/config/TLSServer`

>`vault read -field=URL ejbca/config/TLSServer`

>`vault read -field=CaName ejbca/config/TLSServer`

>`vault read -field=CertProfile ejbca/config/TLSServer`

>`vault read -field=EEProfile ejbca/config/TLSServer`

## Usage

To issue a new certificate, write a CSR to the sign endpoint including the profile identifier. THe profile identifier points to the configuration with the specified CA, Certificate- and End Entity Profile. 

Issuing a certificate requires three properties:
* **csr** - A PKCS#10 CSR in PEM format
* **username** - The username of an [End Entity](https://doc.primekey.com/ejbca/ejbca-operations/ejbca-ca-concept-guide/end-entities-overview) in EJBCA that the certificate will be issued for. If you are used to using EJBCA through the UI, you know a one-time enrollment code is often used, a long random enrolment code (one time) is used in the background by the plugin

```
vault write ejbca/sign/TLSServer csr=@csr.pem username=tomas
```

The list opereation will return the serial numbers of all the certificates in the secrets engine for the specific CA profile. The read operation with the required serial value will return the certificate and its private key if available. The serial number is the hexadecima representation of the serial, a long random number.

>`vault list ejbca/sign/TLSServer`

>`vault list ejbca/sign/TLSServer serial=6718965123650008458`

>`vault read -field=certificate ejbca/sign/TLSServer serial=6718965123650008458` 

### Multiple Configurations
You can enable another configuration simply by configuring a different profileId (the last part of the path). 

Profile 1:
```
vault write ejbca/config/TLSServer pem_bundle=@admin-bundle.pem url=https://ejbca.example.com:8443/ejbca/ejbca-rest-api/v1 cacerts=@admin-TLS-chain.pem caname=TLSAssuredCA certprofile=TLSServer eeprofile=InternalTLSServer
vault write ejbca/sign/TLSServer csr=@csr.pem username=tomas
```

Profile 2:
```
vault write ejbca/config/PROFILE1 pem_bundle=@admin-bundle.pem url=https://ejbca.example.com:8443/ejbca/ejbca-rest-api/v1 cacerts=@admin-TLS-chain.pem caname=MyOtherCA certprofile=MyCertProfile eeprofile=MyEEProfile
vault write ejbca/sign/PROFILE1 csr=@csr.pem username=user1
```

# Future Improvements
Features that are not implemented in this version of the plugin, but can be potential for future feature enhancements are:
* Custom certificate validity. Currently the validity is specified in the Certificate Profile, while EJBCA has capability of custom request validity, this is not available through the plugin.
* Custom subject DN and altName. Currently the subject DN and the altName are taken from the CSR (and validated against the profiles), while EJBCA has the capability of non-CSR subject attributes, this is not available through the plugin.
* Server side generated keystores. Currently only CSR based certificate issuance is available in the plugin, while EJBCA has capabiity to issue server side generated private keys (delivered as for example a PKCS#12 keystrore), EJBCA can not do this from scratch using REST API yet, hence this command is not implemented in the plugin
* Randomize username, for some use cases the username to issue for is not relevant, and we can simply use random username instead
* Keep the private key, used to authenticate the plugin to EJBCA, in itself as a secret in Vault (additional private key protection)
