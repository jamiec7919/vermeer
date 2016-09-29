package qbvh

import (
	m "github.com/jamiec7919/vermeer/math"
)

// This is a n-way BVH

type collapsedbvhnode struct {
	Bounds           m.BoundingBox
	Child, NChildren int32 //-ve if leaf
}

// Initially build this tree
type bvhnode struct {
	Bounds         m.BoundingBox
	LChild, RChild int32 //-ve if leaf
}

func buildbvh(boxes []m.BoundingBox, centroids []m.Vec3, indxs []int32, leafMax int) {

	nodes := make([]bvhnode, 1, len(boxes))

	if len(indxs) < leafMax {
		box := calcBox(boxes, indxs)
		leaf := uint32(1<<31) | (0)
		nodes[0] = bvhnode{box, int32(leaf), int32(len(indxs))}
		return
	}

	_, pivot, _ := binarySplit(boxes, centroids, leafMax, indxs)

	lchild, lbox := buildbvhRec(&nodes, boxes[:pivot], centroids[:pivot], indxs[:pivot], leafMax, 0)
	rchild, rbox := buildbvhRec(&nodes, boxes[pivot:], centroids[pivot:], indxs[pivot:], leafMax, 0+pivot)

	lbox.GrowBox(rbox)
	nodes[0].Bounds = lbox
	nodes[0].LChild = lchild
	nodes[0].RChild = rchild

}

/* Initially build a tree with bvhnode, then collapse into the collapsedbvhnode tree from the
top down.  This should allow easy allocation of indices - either the node is stored or not.
*/
func buildbvhRec(nodes *[]bvhnode, boxes []m.BoundingBox, centroids []m.Vec3, indxs []int32, leafMax, baseIdx int) (node int32, box m.BoundingBox) {

	if len(indxs) < leafMax {
		node = int32(len(*nodes))
		box = calcBox(boxes, indxs)
		*nodes = append(*nodes, bvhnode{box, int32(uint32(1<<31) | uint32(baseIdx)), int32(len(indxs))})
		return
	}

	_, pivot, _ := binarySplit(boxes, centroids, leafMax, indxs)

	node = int32(len(*nodes))
	*nodes = append(*nodes, bvhnode{})

	lchild, lbox := buildbvhRec(nodes, boxes[:pivot], centroids[:pivot], indxs[:pivot], leafMax, baseIdx)
	rchild, rbox := buildbvhRec(nodes, boxes[pivot:], centroids[pivot:], indxs[pivot:], leafMax, baseIdx+pivot)

	lbox.GrowBox(rbox)
	(*nodes)[node].Bounds = lbox
	(*nodes)[node].LChild = lchild
	(*nodes)[node].RChild = rchild

	return node, lbox
}

func collapsetree(bvh []bvhnode) (tree []collapsedbvhnode, indices []int32) {
	return nil, nil
}
