package internal

import (
	"net"
)

// ipRadixTree 基于二进制 Trie 的 IP 前缀树，用于高效 CIDR 匹配。
// 将 IPv4/IPv6 地址按位插入和查找，匹配时间复杂度 O(k)，k 为地址位数（32 或 128）。
type ipRadixTree struct {
	v4Root *radixNode
	v6Root *radixNode
	size   int
}

// radixNode 前缀树节点
type radixNode struct {
	left  *radixNode // bit = 0
	right *radixNode // bit = 1
	leaf  bool       // 标记此节点是否对应一个 CIDR 前缀终点
}

// newIPRadixTree 创建新的 IP 前缀树
func newIPRadixTree() *ipRadixTree {
	return &ipRadixTree{
		v4Root: &radixNode{},
		v6Root: &radixNode{},
	}
}

// Insert 插入一个 CIDR 网段到前缀树中
func (t *ipRadixTree) Insert(cidr *net.IPNet) {
	ip := cidr.IP
	ones, bits := cidr.Mask.Size()

	var root *radixNode
	if bits == 32 {
		root = t.v4Root
		// 对于 IPv4，规范化到 4 字节
		ip = ip.To4()
	} else {
		root = t.v6Root
		ip = ip.To16()
	}

	node := root
	for i := 0; i < ones; i++ {
		// 获取第 i 位的值
		byteIdx := i / 8
		bitIdx := 7 - (i % 8)
		bit := (ip[byteIdx] >> bitIdx) & 1

		if bit == 0 {
			if node.left == nil {
				node.left = &radixNode{}
			}
			node = node.left
		} else {
			if node.right == nil {
				node.right = &radixNode{}
			}
			node = node.right
		}
	}
	node.leaf = true
	t.size++
}

// Contains 检查 IP 是否匹配前缀树中的任意 CIDR 网段
func (t *ipRadixTree) Contains(ip net.IP) bool {
	if ip == nil {
		return false
	}

	var root *radixNode
	var fullIP net.IP

	if ip4 := ip.To4(); ip4 != nil {
		root = t.v4Root
		fullIP = ip4
	} else {
		root = t.v6Root
		fullIP = ip.To16()
	}

	node := root
	maxBits := len(fullIP) * 8

	for i := 0; i < maxBits; i++ {
		byteIdx := i / 8
		bitIdx := 7 - (i % 8)
		bit := (fullIP[byteIdx] >> bitIdx) & 1

		if bit == 0 {
			node = node.left
		} else {
			node = node.right
		}

		if node == nil {
			return false
		}

		// 在路径中的任意节点标记为叶子，表示匹配到 CIDR 前缀
		if node.leaf {
			return true
		}
	}

	return false
}

// Size 返回前缀树中存储的 CIDR 数量
func (t *ipRadixTree) Size() int {
	return t.size
}
