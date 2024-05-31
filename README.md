# vsphere-capacity-manager-data

`vcmd` coordinates the live data between vSphere and the infrastructure it is running upon - in 
our case IBM Cloud. 



### Secrets

Two `json` files need to be created prior to executing. One for all the vSphere environments
and another for all the IBM Cloud accounts where the vSphere environments reside.

IBM Cloud
```json
{
  "ibm-account-name": {
    "Username": "",
    "ApiToken": ""
  },
  "ibm-account-name-2": {
    "Username": "",
    "ApiToken": ""
  }
}
```

vSphere
```json
{
  "vcenter-hostname": {
    "Username": "",
    "Password": ""
  },
  "vcenter2-hostname": {
    "Username": "",
    "Password": ""
  }
}
```

#### Executing `vcmd`
```
$ ./bin/vcmd generate --help
Generate Failure Domains, Capacity data and IBM Cloud subnets

Usage:
  vcmd generate [flags]

Flags:
  -h, --help               help for generate
  -i, --ibmcloud string    vCenter JSON Auth File (default "ibmcloud.json")
  -m, --manifests string   Manifests output path (default "./manifests")
  -p, --pg string          Port Group substring defaults to ci-vlan- (default "ci-vlan-")
  -6, --subnet6 string     IPv6 Subnet defaults to fd65:a1a8:60ad (default "fd65:a1a8:60ad")
  -v, --vcenter string     vCenter JSON Auth File (default "vcenter.json")
```


```
./bin/vcmd -i ./secrets/ibmcloud.json -v ./secrets/vcenter.json -p "ci-vlan-" -6 "fd65:a1a8:60ad" -m ./manifests
```
