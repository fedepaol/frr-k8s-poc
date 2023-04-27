/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FRRConfigurationSpec defines the desired state of FRRConfiguration
type FRRConfigurationSpec struct {
	Routers  []Router `json:"routers"`
	LogLevel string   `json:"logLevel"`
}

// Router represent a neighbor router we want FRR to connect to
type Router struct {
	ASN        uint32     `json:"asn"`
	ID         string     `json:"id"`
	VRF        string     `json:"vrf"`
	Neighbors  []Neighbor `json:"neighbors"`
	PrefixesV4 []string   `json:"prefixesV4"`
	PrefixesV6 []string   `json:"prefixesV6"`
}

type Neighbor struct {
	ASN                uint32          `json:"asn"`
	Address            string          `json:"address"`
	Port               uint16          `json:"port"`
	Password           string          `json:"passwd"`
	AllowedOutPrefixes AllowedPrefixes `json:"allowedOutPrefixes"`
	// TODO all the other parameters defined in the neighbor config
	// AllowedInPrefixes     AllowedPrefixes
	// PrefixesWithLocalPref []LocalPrefPrefixes
	// PrefixesWithCommunity []CommunityPrefixes
}

type AllowedPrefixes struct {
	Prefixes []string `json:"prefixes"`
	AllowAll bool     `json:"allowAll"`
}

/*
type LocalPrefPrefixes struct {
	Prefixes  []string
	LocalPref int
}

type CommunityPrefixes struct {
	Prefixes  []string
	Community string
}
*/

/*
type Config struct {
        Routers     []Router
        BFDProfiles []BFDProfile
}

type Neighbor struct {
        ASN                   string
        IP                    string
        Passwd                string
        AllowedOutPrefixes    AllowedPrefixes
        AllowedInPrefixes     AllowedPrefixes
        PrefixesWithLocalPref []LocalPrefPrefixes
        PrefixesWithCommunity []CommunityPrefixes
}

type AllowedPrefixes struct {
        Prefixes []string
        AllowAll bool
}

type LocalPrefPrefixes struct {
        Prefixes  []string
        LocalPref int
}

type CommunityPrefixes struct {
        Prefixes  []string
        Community string
}

type BFDProfile struct {
}
*/

// FRRConfigurationStatus defines the observed state of FRRConfiguration
type FRRConfigurationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FRRConfiguration is the Schema for the frrconfigurations API
type FRRConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FRRConfigurationSpec   `json:"spec,omitempty"`
	Status FRRConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FRRConfigurationList contains a list of FRRConfiguration
type FRRConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FRRConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FRRConfiguration{}, &FRRConfigurationList{})
}
