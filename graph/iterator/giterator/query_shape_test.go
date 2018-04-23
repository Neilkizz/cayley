// Copyright 2014 The Cayley Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package giterator_test

import (
	"reflect"
	"testing"

	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/cayley/graph/graphmock"
	. "github.com/cayleygraph/cayley/graph/iterator"
	"github.com/cayleygraph/cayley/graph/iterator/giterator"
	"github.com/cayleygraph/cayley/quad"
)

func hasaWithTag(namer giterator.Namer, quads graph.QuadIndexer, tag string, target string) *giterator.HasA {
	and := NewAnd()

	obj := NewFixed(namer.ValueOf(quad.Raw(target)))
	and.AddSubIterator(giterator.NewLinksTo(quads, Tag(obj, tag), quad.Object))

	pred := NewFixed(namer.ValueOf(quad.Raw("status")))
	and.AddSubIterator(giterator.NewLinksTo(quads, pred, quad.Predicate))

	return giterator.NewHasA(quads, and, quad.Subject)
}

func TestQueryShape(t *testing.T) {
	qs := &graphmock.Oldstore{Data: []string{
		1: "cool",
		2: "status",
		3: "fun",
		4: "name",
	}}

	// Given a single linkage iterator's shape.
	var hasa Iterator = hasaWithTag(qs, qs, "tag", "cool")
	hasa = Tag(hasa, "top")

	shape := make(map[string]interface{})
	giterator.OutputQueryShapeForIterator(hasa, qs, shape)

	nodes := shape["nodes"].([]giterator.Node)
	if len(nodes) != 3 {
		t.Errorf("Failed to get correct number of nodes, got:%d expect:4", len(nodes))
	}
	links := shape["links"].([]giterator.Link)
	if len(nodes) != 3 {
		t.Errorf("Failed to get correct number of links, got:%d expect:1", len(links))
	}

	// Nodes should be correctly tagged.
	nodes = shape["nodes"].([]giterator.Node)
	for i, expect := range [][]string{{"tag"}, nil, {"top"}} {
		if !reflect.DeepEqual(nodes[i].Tags, expect) {
			t.Errorf("Failed to get correct tag for node[%d], got:%s expect:%s", i, nodes[i].Tags, expect)
		}
	}
	if !nodes[1].IsLinkNode {
		t.Error("Failed to get node[1] as link node")
	}

	// Link should be correctly typed.
	nodes = shape["nodes"].([]giterator.Node)
	link := shape["links"].([]giterator.Link)[0]
	if link.Source != nodes[2].ID {
		t.Errorf("Failed to get correct link source, got:%v expect:%v", link.Source, nodes[2].ID)
	}
	if link.Target != nodes[0].ID {
		t.Errorf("Failed to get correct link target, got:%v expect:%v", link.Target, nodes[0].ID)
	}
	if link.LinkNode != nodes[1].ID {
		t.Errorf("Failed to get correct link node, got:%v expect:%v", link.LinkNode, nodes[1].ID)
	}
	if link.Pred != 0 {
		t.Errorf("Failed to get correct number of predecessors:%v expect:0", link.Pred)
	}

	// Given a name-of-an-and-iterator's shape.
	andInternal := NewAnd()

	var hasa1 Iterator = hasaWithTag(qs, qs, "tag1", "cool")
	hasa1 = Tag(hasa1, "hasa1")
	andInternal.AddSubIterator(hasa1)

	var hasa2 Iterator = hasaWithTag(qs, qs, "tag2", "fun")
	hasa2 = Tag(hasa2, "hasa2")
	andInternal.AddSubIterator(hasa2)

	pred := NewFixed(qs.ValueOf(quad.Raw("name")))

	and := NewAnd()
	and.AddSubIterator(giterator.NewLinksTo(qs, andInternal, quad.Subject))
	and.AddSubIterator(giterator.NewLinksTo(qs, pred, quad.Predicate))

	shape = make(map[string]interface{})
	giterator.OutputQueryShapeForIterator(giterator.NewHasA(qs, and, quad.Object), qs, shape)

	links = shape["links"].([]giterator.Link)
	if len(links) != 3 {
		t.Errorf("Failed to find the correct number of links, got:%d expect:3", len(links))
	}
	nodes = shape["nodes"].([]giterator.Node)
	if len(nodes) != 7 {
		t.Errorf("Failed to find the correct number of nodes, got:%d expect:7", len(nodes))
	}
	var n int
	for _, node := range nodes {
		if node.IsLinkNode {
			n++
		}
	}
	if n != 3 {
		t.Errorf("Failed to find the correct number of link nodes, got:%d expect:3", n)
	}
}