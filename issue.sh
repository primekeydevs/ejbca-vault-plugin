#!/bin/sh

helpFunction()
{
   echo ""
   echo "Usage: $0 -b parameterB "
   echo -e "\t-b Version of the EJBCA vault plugin to enroll against, e.g. 1"

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


#
# Enroll using locally stored CSR
#
vault write ejbcav$parameterB/config/PROFILE1 pem_bundle=@superadmin-bundle.pem url=https://ejbca.example.com:8443/ejbca/ejbca-rest-api/v1 cacerts=@ManagementCA.pem caname=ManagementCA eeprofile=User certprofile=Client
vault write ejbcav$parameterB/enrollCSR/PROFILE1 csr=@csr.pem username=tomas
vault list ejbcav$parameterB/issued/PROFILE1
