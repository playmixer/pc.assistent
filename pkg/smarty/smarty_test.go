package smarty_test

import (
	"context"
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
	assist.AddCommand([]string{"который час", "какое время", "сколько времени"}, func(ctx context.Context, a *smarty.Assiser) {})
	assist.AddCommand([]string{"включи свет в ванне", "включи в ванне свет"}, func(ctx context.Context, a *smarty.Assiser) {})
}

func TestRotateCommand(t *testing.T) {
	type testRotate struct {
		cmd string
		i   int
		p   int
	}

	cases := []testRotate{
		testRotate{"тсет", 0, 0},
		testRotate{"тест", 0, 100},
		testRotate{"скажи который час", 1, 100},
		testRotate{"какое сейчас время", 1, 100},
		testRotate{"сколько сейчас времени", 1, 100},
		testRotate{"подскажи время", 0, 0},
		testRotate{"включи", 0, 0},
		testRotate{"включи свет в ванне пожалуйста", 2, 100},
	}
	for idx, c := range cases {
		if i, p := assist.RotateCommand(c.cmd); i != c.i || p != c.p {
			t.Fatalf("case#%v idx#%v percent#%v %s", idx, i, p, c.cmd)
		}
	}

}
