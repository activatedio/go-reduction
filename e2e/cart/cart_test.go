package cart

import (
	"fmt"
	"github.com/activatedio/reduction"
	"github.com/activatedio/reduction/e2e/support"
	rmux "github.com/activatedio/reduction/mux"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"go.uber.org/fx"
	"testing"
)

func Test_Cart(t *testing.T) {

	t.Run(support.Wrap(func(client *resty.Client) {

		fmt.Println(client)

	}, Fixture()))

}

func Fixture() fx.Option {
	return fx.Module("fixture", fx.Invoke(func(router *mux.Router) {

		r := reduction.NewReduction()

		rmux.Mount(router, "/cart", r)
	}))
}
