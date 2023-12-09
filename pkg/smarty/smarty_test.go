package smarty_test

import (
	"context"
	"testing"

	"pc.assistent/pkg/smarty"
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

	assist.AddCommand([]string{"тест"}, func(ctx context.Context, a *smarty.Assiser) {})                                          //1
	assist.AddCommand([]string{"который час", "какое время", "сколько времени"}, func(ctx context.Context, a *smarty.Assiser) {}) //2
	assist.AddCommand([]string{"включи свет в ванне", "включи в ванной свет"}, func(ctx context.Context, a *smarty.Assiser) {})   //3
	assist.AddCommand([]string{"выключи свет в ванне", "выключи в ванной свет"}, func(ctx context.Context, a *smarty.Assiser) {}) //4
	assist.AddCommand([]string{"запусти браузер"}, func(ctx context.Context, a *smarty.Assiser) {})                               //5
	assist.AddCommand([]string{"включи стим"}, func(ctx context.Context, a *smarty.Assiser) {})                                   //6
	assist.AddCommand([]string{"отключись", "выключись"}, func(ctx context.Context, a *smarty.Assiser) {})                        //7
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
func TestFoundCommandByToken(t *testing.T) {
	type testRotate struct {
		cmd   string
		i     int
		p     int
		found bool
	}

	cases := []testRotate{
		testRotate{"тсет", 0, 75, true},                            //0
		testRotate{"тест", 0, 100, true},                           //1
		testRotate{"скажи который час", 1, 100, true},              //2
		testRotate{"какое сейчас время", 1, 100, true},             //3
		testRotate{"сколько сейчас времени", 1, 100, true},         //4
		testRotate{"подскажи время", 1, 64, true},                  //5
		testRotate{"включи", 2, 100, true},                         //6
		testRotate{"включи свет в ванне пожалуйста", 2, 100, true}, //7
		testRotate{"включи свет в ванной", 2, 100, true},           //8
		testRotate{"выключить свет в ванной", 3, 95, true},         //9
		testRotate{"выключи свет в ванной", 3, 100, true},          //10
		testRotate{"запусти стим", 4, 74, true},                    //11
		testRotate{"отключись", 6, 100, true},                      //12
		testRotate{"включить", 6, 82, true},                        //13
	}
	for idx, c := range cases {
		if i, p, found := assist.FoundCommandByToken(c.cmd); i != c.i || p != c.p || found != c.found {
			t.Fatalf("case#%v idx=%v percent=%v found=%v %s", idx, i, p, found, c.cmd)
		}
	}

}
func TestFoundCommandByDistance(t *testing.T) {
	type testRotate struct {
		cmd   string
		i     int
		d     int
		found bool
	}

	cases := []testRotate{
		testRotate{"тсет", 0, 2, true},                            //0
		testRotate{"тест", 0, 0, true},                            //1
		testRotate{"скажи который час", 1, 6, true},               //2
		testRotate{"какое сейчас время", 1, 7, true},              //3
		testRotate{"сколько сейчас времени", 1, 7, true},          //4
		testRotate{"подскажи время", 1, 7, true},                  //5
		testRotate{"включи", 6, 3, true},                          //6
		testRotate{"включи свет в ванне пожалуйста", 2, 11, true}, //7
		testRotate{"включи свет в ванной", 2, 2, true},            //8
		testRotate{"выключить свет в ванной", 3, 4, true},         //9
		testRotate{"выключи свет в ванной", 3, 2, true},           //10
		testRotate{"запусти стим", 5, 6, true},                    //11
		testRotate{"отключись", 6, 0, true},                       //12
		testRotate{"включить", 6, 2, true},                        //13
	}
	for idx, c := range cases {
		if i, d, found := assist.FoundCommandByDistance(c.cmd); i != c.i || d != c.d || found != c.found {
			t.Fatalf("case#%v idx=%v distance=%v found=%v %s", idx, i, d, found, c.cmd)
		}
	}

}
func TestFoundCommandByRatio(t *testing.T) {
	type testRotate struct {
		cmd   string
		i     int
		r     int
		found bool
	}

	cases := []testRotate{
		testRotate{"тсет", 0, 75, true},                           //0
		testRotate{"тест", 0, 100, true},                          //1
		testRotate{"скажи который час", 1, 79, true},              //2
		testRotate{"какое сейчас время", 1, 76, true},             //3
		testRotate{"сколько сейчас времени", 1, 81, true},         //4
		testRotate{"подскажи время", 1, 64, true},                 //5
		testRotate{"включи", 6, 80, true},                         //6
		testRotate{"включи свет в ванне пожалуйста", 2, 78, true}, //7
		testRotate{"включи свет в ванной", 2, 92, true},           //8
		testRotate{"выключить свет в ванной", 3, 88, true},        //9
		testRotate{"выключи свет в ванной", 3, 93, true},          //10
		testRotate{"запусти стим", 4, 59, true},                   //11
		testRotate{"отключись", 6, 100, true},                     //12
		testRotate{"включить", 6, 82, true},                       //13
	}
	for idx, c := range cases {
		if i, r, found := assist.FoundCommandByRatio(c.cmd); i != c.i || r != c.r || found != c.found {
			t.Fatalf("case#%v idx=%v ratio=%v found=%v %s", idx, i, r, found, c.cmd)
		}
	}

}

func TestRotateCommand2(t *testing.T) {
	type testRotate struct {
		cmd   string
		i     int
		found bool
	}

	cases := []testRotate{
		testRotate{"тсет", 0, false},                          //0
		testRotate{"тест", 0, true},                           //1
		testRotate{"скажи который час", 1, true},              //2
		testRotate{"какое сейчас время", 1, true},             //3
		testRotate{"сколько сейчас времени", 1, true},         //4
		testRotate{"подскажи время", 0, false},                //5
		testRotate{"включи", 0, false},                        //6
		testRotate{"включи свет в ванне пожалуйста", 2, true}, //7
		testRotate{"включи свет в ванной", 2, true},           //8
		testRotate{"выключить свет в ванной", 3, true},        //9
		testRotate{"выключи свет в ванной", 3, true},          //10
		testRotate{"запусти стим", 0, false},                  //11
		testRotate{"отключись", 6, true},                      //12
		testRotate{"включить", 0, false},                      //13
		testRotate{"включи стин", 0, false},                   //14
	}
	for idx, c := range cases {
		if i, found := assist.RotateCommand2(c.cmd); i != c.i || found != c.found {
			t.Fatalf("case#%v idx=%v found=%v %s", idx, i, found, c.cmd)
		}
	}

}
