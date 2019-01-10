# kubic-init configuration

## Table of Contents

### Configuring Kubic before bootstrapping the cluster

* [Preparing the bootstrap](config-pre.md)
* Deployments:
  * [Deployment examples](../deployments/README.md)
  * Using [`cloud-init`](../deployments/cloud-init/README.md).

### Post control-plane configuration

The Kubernetes cluster running on Kubic can be
configured

* [External authentication](https://github.com/kubic-project/dex-operator/blob/master/README.md)
* [Adding private Docker registries](https://github.com/kubic-project/registries-operator/blob/master/README.md)

### The configuration file

There is an example [kubic-init.yaml](../config/kubic-init.yaml) that has all the possible configurations for kubic.

#### Bootstrap

The bootstrap section is reserved to the configuration that is needed in actions that take part before initializing kubic.

##### Registries

Inside we have the section registries: this section will let you configure mirrors for registries

* multiple mirrors can be set

* each prefix can have multiple mirrors addresses

* each mirror can have certificates configured. This will be used for security to validate the origin of the server.

> Certificates: In this configuration you must add the Certificate content, the Fingerprint and the Hash Algorithm that was used.

This is useful for air-gapped environments, where it is not possible to access upstream registries and you have configured a local mirror.

In this scenario, you have to configure the daemon.json file for the container engine to be able to find the initial images, otherwise Kubic would not start if this was not configured.

Example:

```yaml
bootstrap:
  registries:
    - prefix: https://registry.suse.com
      mirrors:
        - url: https://airgapped.suse.com
        - url: https://airgapped2.suse.com
          certificate: "-----BEGIN CERTIFICATE-----
  MIIGJzCCBA+gAwIBAgIBATANBgkqhkiG9w0BAQUFADCBsjELMAkGA1UEBhMCRlIx
  DzANBgNVBAgMBkFsc2FjZTETMBEGA1UEBwwKU3RyYXNib3VyZzEYMBYGA1UECgwP
  hnx8SB3sVJZHeer8f/UQQwqbAO+Kdy70NmbSaqaVtp8jOxLiidWkwSyRTsuU6D8i
  DiH5uEqBXExjrj0FslxcVKdVj5glVcSmkLwZKbEU1OKwleT/iXFhvooWhQ==
  -----END CERTIFICATE-----"
          fingerprint: "E8:73:0C:C5:84:B1:EB:17:2D:71:54:4D:89:13:EE:47:36:43:8D:BF:5D:3C:0F:5B:FC:75:7E:72:28:A9:7F:73"
          hashalgorithm: "SHA256"
    - prefix: https://registry.io
      mirrors:
        - url: https://airgapped.suse.com
        - url: https://airgapped2.suse.com
          certificate: "-----BEGIN CERTIFICATE-----
  MIIGJzCCBA+gAwIBAgIBATANBgkqhkiG9w0BAQUFADCBsjELMAkGA1UEBhMCRlIx
  DzANBgNVBAgMBkFsc2FjZTETMBEGA1UEBwwKU3RyYXNib3VyZzEYMBYGA1UECgwP
  hnx8SB3sVJZHeer8f/UQQwqbAO+Kdy70NmbSaqaVtp8jOxLiidWkwSyRTsuU6D8i
  DiH5uEqBXExjrj0FslxcVKdVj5glVcSmkLwZKbEU1OKwleT/iXFhvooWhQ==<
  -----END CERTIFICATE-----"
          fingerprint: "E8:73:0C:C5:84:B1:EB:17:2D:71:54:4D:89:13:EE:47:36:43:8D:BF:5D:3C:0F:5B:FC:75:7E:72:28:A9:7F:73"
          hashalgorithm: "SHA256"
```