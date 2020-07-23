package confmongo

import (
	"fmt"
	"testing"
)

func TestMongo_Init(t *testing.T) {
	m := Mongo{}
	m.Init()
	fmt.Println(m.LivenessCheck())
}
