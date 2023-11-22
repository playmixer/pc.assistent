package smarty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/playmixer/pc.assistent/pkg/smarty"
)

var (
	assist *smarty.Assiser
)

type rcgnz struct{}

func (r *rcgnz) Recognize(bufWav []byte) (string, error) {
	return "тест", nil
}

func init() {
	ctx := context.TODO()
	recognize := &rcgnz{}
	assist = smarty.New(ctx, recognize)

	assist.AddCommand([]string{"тест"}, func(ctx context.Context, a *smarty.Assiser) {})
	assist.AddCommand([]string{"который час", "какое время"}, func(ctx context.Context, a *smarty.Assiser) {})
}

func TestRotateCommand(t *testing.T) {
	fmt.Println(assist.RotateCommand("тест"))
	fmt.Println(assist.RotateCommand("скажи который час"))
	fmt.Println(assist.RotateCommand("подскажи время"))
}
