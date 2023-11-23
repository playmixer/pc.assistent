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
	assist.AddCommand([]string{"включи свет в ванне", "включи в ванной свет"}, func(ctx context.Context, a *smarty.Assiser) {})
	assist.AddCommand([]string{"выключи свет в ванне", "выключи в ванной свет"}, func(ctx context.Context, a *smarty.Assiser) {})
}

func TestRotateCommand(t *testing.T) {
	type testRotate struct {
		cmd string
		i   int
		p   int
	}

	cases := []testRotate{
		testRotate{"тсет", 0, 0},                             //0
		testRotate{"тест", 0, 100},                           //1
		testRotate{"скажи который час", 1, 100},              //2
		testRotate{"какое сейчас время", 1, 100},             //3
		testRotate{"сколько сейчас времени", 1, 100},         //4
		testRotate{"подскажи время", 0, 0},                   //5
		testRotate{"включи", 0, 0},                           //6
		testRotate{"включи свет в ванне пожалуйста", 2, 100}, //7
		testRotate{"включи свет в ванной", 2, 100},           //8
		testRotate{"выключить свет в ванной", 0, 0},          //9
		testRotate{"выключи свет в ванной", 3, 100},          //10
	}
	for idx, c := range cases {
		if i, p := assist.RotateCommand(c.cmd); i != c.i || p != c.p {
			t.Fatalf("case#%v idx#%v percent#%v %s", idx, i, p, c.cmd)
		}
	}

}
