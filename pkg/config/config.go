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
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/kubernetes/kubernetes/cmd/kubeadm/app/util/apiclient"
	"github.com/yuroyoro/swalker"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	clientset "k8s.io/client-go/kubernetes"

	kubicutil "github.com/kubic-project/kubic-init/pkg/util"
)

// CniConfiguration The CNI configuration
// Subnets details are specified in the kubeadm configuration file
type CniConfiguration struct {
	BinDir  string `json:"binDir,omitempty"`
	ConfDir string `json:"confDir,omitempty"`
	Driver  string `json:"driver,omitempty"`
	Image   string `json:"image,omitempty"`
}

// ClusterFormationConfiguration struct
type ClusterFormationConfiguration struct {
	Seeder      string `json:"seeder,omitempty"`
	Token       string `json:"token,omitempty"`
	AutoApprove bool   `json:"autoApprove,omitempty"`
}

// OIDCConfiguration struct
type OIDCConfiguration struct {
	Issuer   string `yaml:"issuer,omitempty"`
	ClientID string `yaml:"clientID,omitempty"`
	CA       string `yaml:"ca,omitempty"`
	Username string `yaml:"username,omitempty"`
	Groups   string `yaml:"groups,omitempty"`
}

// AuthConfiguration struct
type AuthConfiguration struct {
	OIDC OIDCConfiguration `yaml:"OIDC,omitempty"`
}

// CertsConfiguration struct
type CertsConfiguration struct {
	// TODO
	Directory string `yaml:"directory,omitempty"`
	CaHash    string `yaml:"caCrtHash,omitempty"`
}

// DNSConfiguration struct
type DNSConfiguration struct {
	Domain       string `yaml:"domain,omitempty"`
	ExternalFqdn string `yaml:"externalFqdn,omitempty"`
}

// ProxyConfiguration struct
type ProxyConfiguration struct {
	HTTP       string `yaml:"http,omitempty"`
	HTTPS      string `yaml:"https,omitempty"`
	NoProxy    string `yaml:"noProxy,omitempty"`
	SystemWide bool   `yaml:"systemWide,omitempty"`
}

// BindConfiguration struct
type BindConfiguration struct {
	Address   string `yaml:"address,omitempty"`
	Interface string `yaml:"interface,omitempty"`
}

// PathsConfigration struct
type PathsConfigration struct {
	Kubeadm string `yaml:"kubeadm,omitempty"`
}

// LocalEtcdConfiguration struct
type LocalEtcdConfiguration struct {
	ServerCertSANs []string `yaml:"serverCertSANs,omitempty"`
	PeerCertSANs   []string `yaml:"peerCertSANs,omitempty"`
}

// EtcdConfiguration struct
type EtcdConfiguration struct {
	LocalEtcd *LocalEtcdConfiguration `yaml:"local,omitempty"`
}

// NetworkConfiguration struct
type NetworkConfiguration struct {
	Bind          BindConfiguration  `yaml:"bind,omitempty"`
	Cni           CniConfiguration   `yaml:"cni,omitempty"`
	DNS           DNSConfiguration   `yaml:"dns,omitempty"`
	Proxy         ProxyConfiguration `yaml:"proxy,omitempty"`
	PodSubnet     string             `yaml:"podSubnet,omitempty"`
	ServiceSubnet string             `yaml:"serviceSubnet,omitempty"`
}

// RuntimeConfiguration struct
type RuntimeConfiguration struct {
	Engine string `yaml:"engine,omitempty"`
}

// FeaturesConfiguration struct
type FeaturesConfiguration struct {
	PSP bool `yaml:"PSP,omitempty"`
}

// ServicesConfiguration struct
type ServicesConfiguration struct {
}

// BootstrapConfiguration se the required configuration
// for bootstraping kubic-init
type BootstrapConfiguration struct {
	Registries []Registry `yaml:"registries,omitempty"`
}

// Registry struct
// Defines a registry mirror
// Prefix: string of the registry that will be replaced
// Mirrors: array with the values to replace the `Prefix`
type Registry struct {
	Prefix  string   `yaml:"prefix"`
	Mirrors []Mirror `yaml:"mirrors"`
}

// Mirror struct
// Defines the Mirrors to be used
// URL: url of the mirror registry.
// Certificate: certificate content for the registry.
// Fingerprint: fingerprint of the certificate to check validity.
// HashAlgorithm: hash algorithm used.
type Mirror struct {
	URL           string `yaml:"url"`
	Certificate   string `yaml:"certificate,omitempty"`
	Fingerprint   string `yaml:"fingerprint,omitempty"`
	HashAlgorithm string `yaml:"hashalgorithm,omitempty"`
}

// KubicInitConfiguration The kubic-init configuration
//
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
type KubicInitConfiguration struct {
	metav1.TypeMeta
	Network          NetworkConfiguration          `yaml:"network,omitempty"`
	Paths            PathsConfigration             `yaml:"paths,omitempty"`
	ClusterFormation ClusterFormationConfiguration `yaml:"clusterFormation,omitempty"`
	Certificates     CertsConfiguration            `yaml:"certificates,omitempty"`
	Etcd             EtcdConfiguration             `yaml:"etcd,omitempty"`
	Runtime          RuntimeConfiguration          `yaml:"runtime,omitempty"`
	Features         FeaturesConfiguration         `yaml:"features,omitempty"`
	Services         ServicesConfiguration         `yaml:"services,omitempty"`
	Auth             AuthConfiguration             `yaml:"auth,omitempty"`
	Bootstrap        BootstrapConfiguration        `yaml:"bootstrap,omitempty"`
}

// defaultConfiguration is the default configuration
var defaultConfiguration = KubicInitConfiguration{
	Certificates: CertsConfiguration{
		Directory: DefaultCertsDirectory,
	},
	Paths: PathsConfigration{
		Kubeadm: DefaultKubeadmPath,
	},
	Etcd: EtcdConfiguration{
		LocalEtcd: &LocalEtcdConfiguration{},
	},
	Network: NetworkConfiguration{
		PodSubnet:     DefaultPodSubnet,
		ServiceSubnet: DefaultServiceSubnet,
		DNS: DNSConfiguration{
			Domain: DefaultDNSDomain,
		},
		Cni: CniConfiguration{
			Driver:  DefaultCniDriver,
			BinDir:  DefaultCniBinDir,
			ConfDir: DefaultCniConfDir,
		},
	},
	ClusterFormation: ClusterFormationConfiguration{
		AutoApprove: true,
	},
	Runtime: RuntimeConfiguration{
		Engine: DefaultRuntimeEngine,
	},
	Features: FeaturesConfiguration{
		PSP: true,
	},
	Bootstrap: BootstrapConfiguration{},
}

// FileAndDefaultsToKubicInitConfig Load a Kubic configuration file, setting some default values
func FileAndDefaultsToKubicInitConfig(cfgPath string) (*KubicInitConfiguration, error) {
	var err error

	internalcfg := defaultConfiguration.DeepCopy()

	if len(cfgPath) > 0 {
		glog.V(1).Infof("[kubic] loading kubic-init configuration from '%s'", cfgPath)
		if os.Stat(cfgPath); err != nil {
			return nil, fmt.Errorf("%q does not exist: %v", cfgPath, err)
		}

		b, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read config from %q [%v]", cfgPath, err)
		}

		if err = yaml.Unmarshal(b, &internalcfg); err != nil {
			return nil, fmt.Errorf("unable to decode config from bytes: %v", err)
		}

		// TODO: check the internalcfg['kind']
	}

	// Overwrite some values with environment variables
	if seederEnv, found := os.LookupEnv(DefaultEnvVarSeeder); found {
		glog.V(3).Infof("[kubic] environment: setting cluster seeder %s", seederEnv)
		internalcfg.ClusterFormation.Seeder = seederEnv
	}

	if tokenEnv, found := os.LookupEnv(DefaultEnvVarToken); found {
		glog.V(3).Infof("[kubic] environment: setting cluster token '%s'", tokenEnv)
		internalcfg.ClusterFormation.Token = tokenEnv
	}

	// The seeder is a IP:PORT, so parse the current seeder and reformat it appropriately
	if len(internalcfg.ClusterFormation.Seeder) > 0 {
		seeder := internalcfg.ClusterFormation.Seeder
		if !strings.HasPrefix(seeder, "http") {
			seeder = fmt.Sprintf("https://%s", internalcfg.ClusterFormation.Seeder)
		}
		u, err := url.Parse(seeder)
		if err != nil {
			return nil, err
		}
		port := u.Port()

		// if no port has been provided, use the API server default port
		if len(port) == 0 {
			port = fmt.Sprintf("%d", DefaultAPIServerPort)
		}

		internalcfg.ClusterFormation.Seeder = fmt.Sprintf("%s:%s", u.Hostname(), port)
	}

	if glog.V(8) {
		marshalled, err := yaml.Marshal(internalcfg)
		if err != nil {
			return nil, err
		}
		glog.Infof("[kubic] after parsing the config file:\n%s", marshalled)
	}

	return internalcfg, nil
}

// ToConfigMap uploads the configuration to a "kubic-init.yaml" file in a ConfigMap
func (kubicCfg *KubicInitConfiguration) ToConfigMap(client clientset.Interface, name string, extraLabels map[string]string) error {
	filename := filepath.Base(DefaultKubicInitConfig)

	glog.V(3).Infof("[kubic] uploading to ConfigMap %s/%s the '%s' configuration",
		metav1.NamespaceSystem, name, filename)

	// TODO: check there is no sensible information in the kubicCfg and remove it...

	marshalled, err := yaml.Marshal(kubicCfg)
	if err != nil {
		return err
	}

	if err := apiclient.CreateOrUpdateConfigMap(client, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceSystem,
			Labels:    extraLabels,
		},
		Data: map[string]string{
			filename: string(marshalled),
		},
	}); err != nil {
		return err
	}

	glog.V(3).Infof("[kubic] configuration uploaded to ConfigMap %s/%s",
		metav1.NamespaceSystem, name)

	return nil
}

// SetVars parses a list of assignments (like "key=value"), where "key"
// is a path in the configuration hierarchy (ie, "Network.Cni.Driver")
func (kubicCfg *KubicInitConfiguration) SetVars(vars []string) error {
	if len(vars) > 0 {
		for _, v := range vars {
			components := strings.Split(v, "=")
			if len(components) != 2 {
				return fmt.Errorf("cannot parse '%s' as an assignment", v)
			}

			glog.V(8).Infof("[kubic] setting '%s'='%s'", components[0], components[1])
			boolVar, err := strconv.ParseBool(components[1])
			if err != nil {
				if err := swalker.Write(components[0], kubicCfg, components[1]); err != nil {
					return err
				}
			} else {
				if err := swalker.Write(components[0], kubicCfg, boolVar); err != nil {
					return err
				}
			}
		}

		if glog.V(8) {
			marshalled, err := yaml.Marshal(kubicCfg)
			if err != nil {
				return err
			}
			glog.Infof("[kubic] after parsing the list of variables:\n%s", marshalled)
		}
	}

	return nil
}

// IsSeeder returns if the config is a seeder config
func (kubicCfg KubicInitConfiguration) IsSeeder() bool {
	return len(kubicCfg.ClusterFormation.Seeder) == 0
}

// GetBindIP gets a valid IP address where we can bind
func (kubicCfg KubicInitConfiguration) GetBindIP() (net.IP, error) {
	if len(kubicCfg.Network.Bind.Interface) > 0 {
		ifc, err := net.InterfaceByName(kubicCfg.Network.Bind.Interface)
		if err != nil {
			return nil, err
		}

		addrs, err := ifc.Addrs()
		if err != nil {
			return nil, err
		}

		// just return the first IP (maybe we could do some smarter heuristics...)
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				return v.IP, nil
			case *net.IPAddr:
				return v.IP, nil
			}
		}

		return nil, fmt.Errorf("No address found in %s", kubicCfg.Network.Bind.Interface)
	}

	defaultAddrStr := "0.0.0.0"
	if len(kubicCfg.Network.Bind.Address) > 0 {
		defaultAddrStr = kubicCfg.Network.Bind.Address
	}

	defaultAddr := net.ParseIP(defaultAddrStr)
	bindIP, err := utilnet.ChooseBindAddress(defaultAddr)
	if err != nil {
		return nil, err
	}
	return bindIP, nil
}

// GetPublicAPIAddress gets a DNS name (or IP address)
// that can be used for reaching the API server
func (kubicCfg KubicInitConfiguration) GetPublicAPIAddress() (string, error) {
	if len(kubicCfg.Network.DNS.ExternalFqdn) > 0 {
		return kubicCfg.Network.DNS.ExternalFqdn, nil
	}

	// ok, we don't have a user-provided DNS name, so we
	// must apply some heuristics...

	// 1. if this is the seeder, there will be an API server running here:
	// just return the local IP address as the IP address of the API server
	if kubicCfg.IsSeeder() {
		localIP, err := utilnet.ChooseHostInterface()
		if err != nil {
			return "", err
		}
		return localIP.String(), nil
	}

	return "", fmt.Errorf("cannot determine an public DNS name or address for the API server")
}

// GetServiceDNSName gets a FQDN DNS name in ther internal network for `name`
func (kubicCfg KubicInitConfiguration) GetServiceDNSName(obj kubicutil.ObjNamespacer) string {
	domain := kubicCfg.Network.DNS.Domain
	if len(obj.GetNamespace()) > 0 {
		return fmt.Sprintf("%s.%s.svc.%s", obj.GetName(), obj.GetNamespace(), domain)
	}
	return fmt.Sprintf("%s.svc.%s", obj.GetName(), domain)
}
