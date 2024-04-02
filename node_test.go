package ptrie

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/assertly"
	"github.com/viant/toolbox"
	"log"
	"strings"
	"testing"
)

func TestNode_LoadNode(t *testing.T) {
	useCases := []struct {
		description string
		keywords    []string
	}{
		{
			description: "basic_encode",
			keywords:    []string{"abc", "zyx", "mln", "abcd", "abz"},
		},
		{
			description: "prefix_encode",
			keywords:    []string{"abc", "zyx", "abcd"},
		},
		{
			description: "edge_encode",
			keywords:    []string{"abc", "ac", "zyx"},
		},
	}

	for _, useCase := range useCases {
		node := newValueNode[uint32]([]byte("/"), 0)
		for i, keyword := range useCase.keywords {
			node.add(newValueNode[uint32]([]byte(keyword), uint32(i+1)), nil)
		}
		data := node.Data()

		cloned := &Node[uint32]{}
		cloned.LoadNode(data)
		assert.True(t, node.Equals(cloned))
		if !assertly.AssertValues(t, node, cloned, useCase.description) {
			_ = toolbox.DumpIndent(node, true)
			fmt.Println("CLONE BEGIN")
			_ = toolbox.DumpIndent(cloned, true)
			fmt.Println("CLONE END")
		}
	}

}

func TestNode_Decode(t *testing.T) {
	useCases := []struct {
		description string
		keywords    []string
	}{
		{
			description: "basic_encode",
			keywords:    []string{"abc", "zyx", "mln"},
		},
		{
			description: "prefix_encode",
			keywords:    []string{"abc", "zyx", "abcd"},
		},
		{
			description: "edge_encode",
			keywords:    []string{"abc", "ac", "zyx"},
		},
	}

	for _, useCase := range useCases {
		node := newValueNode[uint32]([]byte("/"), 0)
		for i, keyword := range useCase.keywords {
			node.add(newValueNode[uint32]([]byte(keyword), uint32(i+1)), nil)
		}
		writer := new(bytes.Buffer)
		err := node.Encode(writer)
		if !assert.Nil(t, err, useCase.description) {
			log.Print(err)
			continue
		}
		assert.EqualValues(t, writer.Len(), node.size(), useCase.description)

		cloned := &Node[uint32]{}
		err = cloned.Decode(bytes.NewReader(writer.Bytes()))
		if assert.Nil(t, err, useCase.description) {
			continue
		}

		if !assertly.AssertValues(t, node, cloned, useCase.description) {
			_ = toolbox.DumpIndent(node, true)
			_ = toolbox.DumpIndent(cloned, true)
		}
	}

	//test error case
	reader := strings.NewReader("test is error")
	node := &Node[uint32]{}
	err := node.Decode(reader)
	assert.NotNil(t, err)
}

func TestNode_walk(t *testing.T) {
	useCases := []struct {
		description string
		keywords    []string
	}{
		{
			description: "basic_encode",
			keywords:    []string{"abc", "zyx", "mln"},
		},
		{
			description: "prefix_encode",
			keywords:    []string{"abc", "zyx", "abcd"},
		},
		{
			description: "edge_encode",
			keywords:    []string{"abc", "ac", "zyx"},
		},
		{
			description: "merge_node",
			keywords:    []string{"abc", "ac", "zyx", "abc", "abcdefx"},
		},
	}

	for _, useCase := range useCases {
		node := newValueNode[uint32]([]byte(""), 0)
		var expect = make(map[string]uint32)
		var actual = make(map[string]uint32)
		for i, keyword := range useCase.keywords {
			expect[string(keyword)] = uint32(i + 1)
			node.add(newValueNode[uint32]([]byte(keyword), uint32(i+1)), nil)
		}
		node.walk([]byte{}, func(key []byte, valueIndex uint32) {
			actual[string(key)] = valueIndex
		})
		assert.Equal(t, expect, actual, useCase.description)
	}
}

func TestNode_match(t *testing.T) {
	useCases := []struct {
		description      string
		keywords         []string
		input            string
		matchAll         bool
		expectHasMatched bool
		expectMatched    map[string]uint32
	}{
		{
			description:      "prefix match",
			keywords:         []string{"abc", "zyx", "mln", "bar", "abcd", "abcdex", "bc"},
			input:            "bc is",
			expectHasMatched: true,
			expectMatched: map[string]uint32{
				"bc": uint32(7),
			},
		},

		{
			description:      "exact match",
			keywords:         []string{"abc", "zyx", "mln", "abcd", "abcdex"},
			input:            "abc",
			expectHasMatched: true,
			expectMatched: map[string]uint32{
				"abc": uint32(1),
			},
		},

		{
			description:      "prefix multi match",
			keywords:         []string{"abc", "zyx", "mln", "abcd", "abcdex"},
			input:            "abcdex",
			expectHasMatched: true,
			matchAll:         true,
			expectMatched: map[string]uint32{
				"abc":    uint32(1),
				"abcd":   uint32(4),
				"abcdex": uint32(5),
			},
		},
		{
			description:      "first multi match",
			keywords:         []string{"abc", "zyx", "mln", "abcd", "abcdex"},
			input:            "abcdex",
			expectHasMatched: true,
			matchAll:         false,
			expectMatched: map[string]uint32{
				"abc": uint32(1),
			},
		},
		{
			description:      "first prefix multi match",
			keywords:         []string{"abc", "zyx", "mln", "abcd", "abcdex", "zz"},
			input:            "abcdex zz",
			expectHasMatched: true,
			matchAll:         false,
			expectMatched: map[string]uint32{
				"abc": uint32(1),
			},
		},
		{
			description:      "first prefix multi match2",
			keywords:         []string{"abcz", "abrz", "mln", "a", "abc", "ab", "zz", "abcd"},
			input:            "abcdex zz",
			expectHasMatched: true,
			matchAll:         true,
			expectMatched: map[string]uint32{
				"a":    uint32(4),
				"abc":  uint32(5),
				"ab":   uint32(6),
				"abcd": uint32(8),
			},
		},

		{
			description:      "mult level amtch",
			keywords:         []string{"ab", "abc", "abcd", "a"},
			input:            "abcd",
			expectHasMatched: true,
			matchAll:         true,
			expectMatched: map[string]uint32{
				"abc":  uint32(2),
				"ab":   uint32(1),
				"abcd": uint32(3),
				"a":    uint32(4),
			},
		},
		{
			description: "mult level amtch",
			keywords: []string{"petest.com.br",
				"petest.com¬†",
				"petest.com,¬†",
				"petest.com",
				".com",
				"pe",
				"petest",
				"e",
				"r",
				"pe",
				"hilton.com"},
			input:            "petest.com",
			expectHasMatched: true,
			matchAll:         true,
			expectMatched: map[string]uint32{
				"petest.com": uint32(4),
				"pe":         uint32(6),
				"petest":     uint32(7),
			},
		},
	}

	for _, useCase := range useCases {
		node := newValueNode[uint32]([]byte(""), 0)
		for i, keyword := range useCase.keywords {
			node.add(newValueNode[uint32]([]byte(keyword), uint32(i+1)), func(prev uint32) uint32 {
				return prev
			})
		}
		var actualMatched = make(map[string]uint32)
		actualHasMatched := node.match([]byte(useCase.input), 0, func(key []byte, valueIndex uint32) bool {
			actualMatched[string(key)] = valueIndex
			return useCase.matchAll
		})
		assert.Equal(t, useCase.expectHasMatched, actualHasMatched, useCase.description)
		if useCase.expectHasMatched {
			assert.Equal(t, useCase.expectMatched, actualMatched, useCase.description)
		}

	}
}
