/*
 * Copyright 2019 SUSE LINUX GmbH, Nuernberg, Germany..
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

//go:generate sh -c "GO111MODULE=off deepcopy-gen -O zz_generated.deepcopy -i ./... -h ../../hack/boilerplate.go.txt"

package config

import (
  "io/ioutil"
  "os"
  "reflect"
  "testing"
)

const (
  configContent = `---
apiVersion: kubic.suse.com/v1alpha2
kind: KubicInitConfiguration
features:
  PSP: true
runtime:
  engine: crio
paths:
  kubeadm: /usr/bin/kubeadm
auth:
  oidc:
   issuer: https://some.name.com:32000
   clientID: kubernetes
   ca: /etc/kubernetes/pki/ca.crt
   username: email
   groups: groups
certificates:
  directory: /etc/kubernetes/pki
  caCrt:
  caCrtHash:
etcd:
  local:
    serverCertSANs: []
    peerCertSANs: []
manager:
  image: "kubic-init:latest"
clusterFormation:
  seeder: some-node.com
  token: 94dcda.c271f4ff502789ca
  autoApprove: false
network:
  bind:
    address: 0.0.0.0
    interface: eth0
  podSubnet: "172.16.0.0/13"
  serviceSubnet: "172.24.0.0/16"
  proxy:
    http: my-proxy.com:8080
    https: my-proxy.com:8080
    noProxy: localdomain.com
    systemwide: false
  dns:
    domain: someDomain.local
    externalFqdn: some.name.com
  cni:
    driver: flannel
    image: registry.opensuse.org/devel/caasp/kubic-container/container/kubic/flannel:0.9.1
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
DiH5uEqBXExjrj0FslxcVKdVj5glVcSmkLwZKbEU1OKwleT/iXFhvooWhQ==
-----END CERTIFICATE-----"
          fingerprint: "E8:73:0C:C5:84:B1:EB:17:2D:71:54:4D:89:13:EE:47:36:43:8D:BF:5D:3C:0F:5B:FC:75:7E:72:28:A9:7F:73"
          hashalgorithm: "SHA256"
`
  filename = "kubic-init.mirrors.yaml"
)

// TestFileAndDefaultsToKubicInitConfig will test loading the configuration `yaml` file and parsing it to the defined struct
func TestFileAndDefaultsToKubicInitConfig(t *testing.T) {
  registries := []Registry{
    {"https://registry.suse.com",
      []Mirror{{URL: "https://airgapped.suse.com"},
        {URL: "https://airgapped2.suse.com", Certificate: `-----BEGIN CERTIFICATE----- MIIGJzCCBA+gAwIBAgIBATANBgkqhkiG9w0BAQUFADCBsjELMAkGA1UEBhMCRlIx DzANBgNVBAgMBkFsc2FjZTETMBEGA1UEBwwKU3RyYXNib3VyZzEYMBYGA1UECgwP hnx8SB3sVJZHeer8f/UQQwqbAO+Kdy70NmbSaqaVtp8jOxLiidWkwSyRTsuU6D8i DiH5uEqBXExjrj0FslxcVKdVj5glVcSmkLwZKbEU1OKwleT/iXFhvooWhQ== -----END CERTIFICATE-----`,
          Fingerprint: "E8:73:0C:C5:84:B1:EB:17:2D:71:54:4D:89:13:EE:47:36:43:8D:BF:5D:3C:0F:5B:FC:75:7E:72:28:A9:7F:73", HashAlgorithm: "SHA256"}}},
    {"https://registry.io",
      []Mirror{{URL: "https://airgapped.suse.com"},
        {URL: "https://airgapped2.suse.com", Certificate: `-----BEGIN CERTIFICATE----- MIIGJzCCBA+gAwIBAgIBATANBgkqhkiG9w0BAQUFADCBsjELMAkGA1UEBhMCRlIx DzANBgNVBAgMBkFsc2FjZTETMBEGA1UEBwwKU3RyYXNib3VyZzEYMBYGA1UECgwP hnx8SB3sVJZHeer8f/UQQwqbAO+Kdy70NmbSaqaVtp8jOxLiidWkwSyRTsuU6D8i DiH5uEqBXExjrj0FslxcVKdVj5glVcSmkLwZKbEU1OKwleT/iXFhvooWhQ== -----END CERTIFICATE-----`,
          Fingerprint: "E8:73:0C:C5:84:B1:EB:17:2D:71:54:4D:89:13:EE:47:36:43:8D:BF:5D:3C:0F:5B:FC:75:7E:72:28:A9:7F:73", HashAlgorithm: "SHA256"}}}}
  err := ioutil.WriteFile(filename, []byte(configContent), os.FileMode(0644))
  if err != nil {
    t.Fatalf("faliled to write config file: %s", err)
  }
  defer os.RemoveAll(filename)
  type args struct {
    cfgPath string
  }
  tests := []struct {
    name       string
    args       args
    want       *KubicInitConfiguration
    wantStruct interface{}
    wantErr    bool
  }{
    {"test_missing_file", args{"kubic-init.yaml"}, nil, nil, true},
    {"test_mirror_config", args{filename}, nil, registries, false},
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      got, err := FileAndDefaultsToKubicInitConfig(tt.args.cfgPath)
      if (err != nil) != tt.wantErr {
        t.Errorf("FileAndDefaultsToKubicInitConfig() error = %v, wantErr %v", err, tt.wantErr)
        return
      }
      switch tt.name {
      case "test_mirror_config":
        if !reflect.DeepEqual(got.Bootstrap.Registries, tt.wantStruct) {
          t.Errorf("FileAndDefaultsToKubicInitConfig() = %v, want %v", got.Bootstrap.Registries, tt.wantStruct)
          return
        }
      default:
        if got != tt.want {
          t.Errorf("FileAndDefaultsToKubicInitConfig() = %v, want %v", got, tt.want)
          return
        }
      }
    })
  }
}
