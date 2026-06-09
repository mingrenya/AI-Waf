package internal

import (
	"net"
	"testing"
)

func TestIPRadixTree_IPv4(t *testing.T) {
	tree := newIPRadixTree()

	// Insert CIDRs
	_, cidr1, _ := net.ParseCIDR("192.168.1.0/24")
	_, cidr2, _ := net.ParseCIDR("10.0.0.0/8")
	_, cidr3, _ := net.ParseCIDR("172.16.0.0/12")

	tree.Insert(cidr1)
	tree.Insert(cidr2)
	tree.Insert(cidr3)

	if tree.Size() != 3 {
		t.Fatalf("expected size 3, got %d", tree.Size())
	}

	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.100", true},
		{"192.168.1.1", true},
		{"192.168.2.1", false}, // 192.168.2.x not in 192.168.1.0/24
		{"192.168.0.1", false},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"11.0.0.1", false},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"172.32.0.1", false},
		{"8.8.8.8", false},
	}

	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		got := tree.Contains(ip)
		if got != tt.expected {
			t.Errorf("IP %s: expected %v, got %v", tt.ip, tt.expected, got)
		}
	}
}

func TestIPRadixTree_OverlappingCIDRs(t *testing.T) {
	tree := newIPRadixTree()

	// Insert larger then smaller overlapping CIDR
	_, cidr1, _ := net.ParseCIDR("10.0.0.0/8")
	_, cidr2, _ := net.ParseCIDR("10.1.0.0/16")

	tree.Insert(cidr1)
	tree.Insert(cidr2)

	// Both should match
	if !tree.Contains(net.ParseIP("10.1.0.1")) {
		t.Error("10.1.0.1 should match")
	}
	if !tree.Contains(net.ParseIP("10.2.0.1")) {
		t.Error("10.2.0.1 should match (only in /8, not /16)")
	}
}

func TestIPRadixTree_ExactIP(t *testing.T) {
	tree := newIPRadixTree()

	// /32 is exact IP match
	_, cidr, _ := net.ParseCIDR("1.2.3.4/32")
	tree.Insert(cidr)

	if !tree.Contains(net.ParseIP("1.2.3.4")) {
		t.Error("exact IP should match /32")
	}
	if tree.Contains(net.ParseIP("1.2.3.5")) {
		t.Error("1.2.3.5 should not match 1.2.3.4/32")
	}
}

func TestIPRadixTree_Empty(t *testing.T) {
	tree := newIPRadixTree()
	if tree.Contains(net.ParseIP("1.2.3.4")) {
		t.Error("empty tree should not match anything")
	}
}

func BenchmarkIPRadixTree_Contains(b *testing.B) {
	tree := newIPRadixTree()

	// Insert 1000 CIDRs
	for i := 0; i < 1000; i++ {
		ip := net.IPv4(byte(i>>8), byte(i&0xFF), 0, 0)
		_, cidr, _ := net.ParseCIDR(ip.String() + "/24")
		tree.Insert(cidr)
	}

	testIP := net.ParseIP("192.168.1.100")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Contains(testIP)
	}
}

func BenchmarkLinearCIDR_Contains(b *testing.B) {
	var cidrNets []*net.IPNet

	// Insert 1000 CIDRs (same as above)
	for i := 0; i < 1000; i++ {
		ip := net.IPv4(byte(i>>8), byte(i&0xFF), 0, 0)
		_, cidr, _ := net.ParseCIDR(ip.String() + "/24")
		cidrNets = append(cidrNets, cidr)
	}

	testIP := net.ParseIP("192.168.1.100")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ipNet := range cidrNets {
			if ipNet.Contains(testIP) {
				break
			}
		}
	}
}
