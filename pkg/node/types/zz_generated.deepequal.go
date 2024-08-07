//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

// Code generated by deepequal-gen. DO NOT EDIT.

package types

// DeepEqual is an autogenerated deepequal function, deeply comparing the
// receiver with other. in must be non-nil.
func (in *Node) DeepEqual(other *Node) bool {
	if other == nil {
		return false
	}

	if in.Name != other.Name {
		return false
	}
	if in.Cluster != other.Cluster {
		return false
	}
	if ((in.IPAddresses != nil) && (other.IPAddresses != nil)) || ((in.IPAddresses == nil) != (other.IPAddresses == nil)) {
		in, other := &in.IPAddresses, &other.IPAddresses
		if other == nil {
			return false
		}

		if len(*in) != len(*other) {
			return false
		} else {
			for i, inElement := range *in {
				if !inElement.DeepEqual(&(*other)[i]) {
					return false
				}
			}
		}
	}

	if (in.IPv4AllocCIDR == nil) != (other.IPv4AllocCIDR == nil) {
		return false
	} else if in.IPv4AllocCIDR != nil {
		if !in.IPv4AllocCIDR.DeepEqual(other.IPv4AllocCIDR) {
			return false
		}
	}

	if ((in.IPv4SecondaryAllocCIDRs != nil) && (other.IPv4SecondaryAllocCIDRs != nil)) || ((in.IPv4SecondaryAllocCIDRs == nil) != (other.IPv4SecondaryAllocCIDRs == nil)) {
		in, other := &in.IPv4SecondaryAllocCIDRs, &other.IPv4SecondaryAllocCIDRs
		if other == nil {
			return false
		}

		if len(*in) != len(*other) {
			return false
		} else {
			for i, inElement := range *in {
				if !inElement.DeepEqual((*other)[i]) {
					return false
				}
			}
		}
	}

	if (in.IPv6AllocCIDR == nil) != (other.IPv6AllocCIDR == nil) {
		return false
	} else if in.IPv6AllocCIDR != nil {
		if !in.IPv6AllocCIDR.DeepEqual(other.IPv6AllocCIDR) {
			return false
		}
	}

	if ((in.IPv6SecondaryAllocCIDRs != nil) && (other.IPv6SecondaryAllocCIDRs != nil)) || ((in.IPv6SecondaryAllocCIDRs == nil) != (other.IPv6SecondaryAllocCIDRs == nil)) {
		in, other := &in.IPv6SecondaryAllocCIDRs, &other.IPv6SecondaryAllocCIDRs
		if other == nil {
			return false
		}

		if len(*in) != len(*other) {
			return false
		} else {
			for i, inElement := range *in {
				if !inElement.DeepEqual((*other)[i]) {
					return false
				}
			}
		}
	}

	if ((in.IPv4HealthIP != nil) && (other.IPv4HealthIP != nil)) || ((in.IPv4HealthIP == nil) != (other.IPv4HealthIP == nil)) {
		in, other := &in.IPv4HealthIP, &other.IPv4HealthIP
		if other == nil {
			return false
		}

		if len(*in) != len(*other) {
			return false
		} else {
			for i, inElement := range *in {
				if inElement != (*other)[i] {
					return false
				}
			}
		}
	}

	if ((in.IPv6HealthIP != nil) && (other.IPv6HealthIP != nil)) || ((in.IPv6HealthIP == nil) != (other.IPv6HealthIP == nil)) {
		in, other := &in.IPv6HealthIP, &other.IPv6HealthIP
		if other == nil {
			return false
		}

		if len(*in) != len(*other) {
			return false
		} else {
			for i, inElement := range *in {
				if inElement != (*other)[i] {
					return false
				}
			}
		}
	}

	if ((in.IPv4IngressIP != nil) && (other.IPv4IngressIP != nil)) || ((in.IPv4IngressIP == nil) != (other.IPv4IngressIP == nil)) {
		in, other := &in.IPv4IngressIP, &other.IPv4IngressIP
		if other == nil {
			return false
		}

		if len(*in) != len(*other) {
			return false
		} else {
			for i, inElement := range *in {
				if inElement != (*other)[i] {
					return false
				}
			}
		}
	}

	if ((in.IPv6IngressIP != nil) && (other.IPv6IngressIP != nil)) || ((in.IPv6IngressIP == nil) != (other.IPv6IngressIP == nil)) {
		in, other := &in.IPv6IngressIP, &other.IPv6IngressIP
		if other == nil {
			return false
		}

		if len(*in) != len(*other) {
			return false
		} else {
			for i, inElement := range *in {
				if inElement != (*other)[i] {
					return false
				}
			}
		}
	}

	if in.ClusterID != other.ClusterID {
		return false
	}
	if in.Source != other.Source {
		return false
	}
	if in.EncryptionKey != other.EncryptionKey {
		return false
	}
	if ((in.Labels != nil) && (other.Labels != nil)) || ((in.Labels == nil) != (other.Labels == nil)) {
		in, other := &in.Labels, &other.Labels
		if other == nil {
			return false
		}

		if len(*in) != len(*other) {
			return false
		} else {
			for key, inValue := range *in {
				if otherValue, present := (*other)[key]; !present {
					return false
				} else {
					if inValue != otherValue {
						return false
					}
				}
			}
		}
	}

	if ((in.Annotations != nil) && (other.Annotations != nil)) || ((in.Annotations == nil) != (other.Annotations == nil)) {
		in, other := &in.Annotations, &other.Annotations
		if other == nil {
			return false
		}

		if len(*in) != len(*other) {
			return false
		} else {
			for key, inValue := range *in {
				if otherValue, present := (*other)[key]; !present {
					return false
				} else {
					if inValue != otherValue {
						return false
					}
				}
			}
		}
	}

	if in.NodeIdentity != other.NodeIdentity {
		return false
	}
	if in.WireguardPubKey != other.WireguardPubKey {
		return false
	}
	if in.BootID != other.BootID {
		return false
	}

	return true
}
