#!/bin/sh

helpFunction()
{
   echo ""
   echo "Usage: $0 -b parameterB "
   echo -e "\t-b Build number version for the plugin, this should be incremented after each build, e.g. 1"
   exit 1 # Exit script after printing help
}

while getopts "b:" opt
do
   case "$opt" in
      b ) parameterB="$OPTARG" ;;
      ? ) helpFunction ;; # Print helpFunction in case parameter is non-existent
   esac
done

# Print helpFunction in case parameters are empty
if [ -z "$parameterB" ] 
then
   echo "Some or all of the parameters are empty";
   helpFunction
fi

# Build the plugin
go build -o out/ejbca-vault-plugin-v$parameterB
SHA256=`sha256sum out/ejbca-vault-plugin-v$parameterB | awk '{ print $1 }'`
echo "SHA256: $SHA256"

# Disable existing plugin and register and enable new version
#vault secrets disable ejbca
#vault write sys/plugins/catalog/ejbca-vault-plugin sha256=${SHA256} command="ejbca-vault-plugin"
vault plugin register -sha256=${SHA256} secret ejbca-vault-plugin-v$parameterB
vault secrets enable -path=ejbcav$parameterB ejbca-vault-plugin-v$parameterB

#
# Enroll using locally stored CSR
#
#vault write ejbca/config/PROFILE1 pem_bundle=@superadmin-bundle.pem url=https://ejbca.example.com:8443/ejbca/ejbca-rest-api/v1 cacerts=@ManagementCA.pem caname=ManagementCA eeprofile=User certprofile=Client
#vault write ejbca/enrollCSR/PROFILE1 csr=@csr.pem username=tomas
#vault list ejbca/issued/PROFILE1
