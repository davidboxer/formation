package controller

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"hash/fnv"
	"k8s.io/apimachinery/pkg/runtime"
)

// HashObject returns the hash of a Object hash by a Codec
func HashObject(obj runtime.Object) string {
	hf := fnv.New32()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	_, _ = printer.Fprintf(hf, "%#v", obj)
	return fmt.Sprint(hf.Sum32())
}
