package yandex

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"pc.assistent/pkg/listen"
)

var (
	yndx *YDX
	list *listen.Listener
)

func init() {
	// yndx = New("y0_AgAAAAA0g5MWAATuwQAAAADZofUOrWSgv_VEQHyExxedezRXnPXK_No", "b1gp2jj6l638bnvf2eij")
	yndx = New("y0_AgAAAAA0g5MWAATuwQAAAADZofUOrWSgv_VEQHyExxedezRXnPXK_No", "ajei3uk85p0p04pif8jf")
	list = listen.New(time.Second * 2)
}

func TestUpdToken(t *testing.T) {
	err := yndx.UpdIamToken()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(yndx.IamToken, yndx.IamTokenExpires)
	if yndx.IamToken == "" {
		t.Fatal("iam token is empty")
	}
	token1 := yndx.IamToken

	err = yndx.UpdIamToken()
	if err != nil {
		t.Fatal(err)
	}

	token2 := yndx.IamToken
	fmt.Println(yndx.IamToken, yndx.IamTokenExpires)
	if yndx.IamToken == "" {
		t.Fatal("iam token is empty")
	}

	if token1 != token2 {
		t.Fatal("tokens is not equale")
	}
}

func TestRecognizeByte(t *testing.T) {
	// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	// 	defer cancel()
	// 	list.Start(ctx)

	// waitFor:
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			// cancel()
	// 			break waitFor
	// 		case r := <-list.WavCh:
	// 			res, err := yndx.RecognizeByte(r)
	// 			if err != nil {
	// 				cancel()
	// 				t.Fatal(err)
	// 			}
	// 			fmt.Println(res)
	// 		}
	// 	}

	f, err := os.Open("f:\\Downloads\\test2.wav")
	if err != nil {
		t.Fatal(err.Error())
	}
	body, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err.Error())
	}

	res, err := yndx.RecognizeByte(body)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(res)
	if res.Result == "" {
		t.Fatal("result is empty")
	}

}
