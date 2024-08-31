package cart

import (
	"context"
	"github.com/activatedio/go-reduction"
	"github.com/activatedio/go-reduction/e2e/support"
	rmux "github.com/activatedio/go-reduction/mux"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"net/http"
	"reflect"
	"testing"
)

type Cart struct {
	ItemCount int `json:"item_count"`
}

type AddItem struct {
	Qty int `json:"qty"`
}

func Test_Cart_WithInit(t *testing.T) {

	t.Run(support.Wrap(func(client *resty.Client) {

		newCart := func() *Cart {
			return &Cart{}
		}

		cart := newCart()

		resp, err := client.R().SetResult(cart).Get("/cart")
		check(err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, &Cart{
			ItemCount: 0,
		}, cart)

		cart = newCart()

		resp, err = client.R().SetResult(cart).SetBody(&AddItem{
			Qty: 10,
		}).Post("/cart/add_item")
		check(err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, &Cart{
			ItemCount: 10,
		}, cart)

		resp, err = client.R().SetResult(cart).SetBody(&AddItem{
			Qty: 10,
		}).Post("/cart/add_item")
		check(err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, &Cart{
			ItemCount: 20,
		}, cart)

	}, fx.Module("fixture", fx.Invoke(func(factory reduction.Factory, router *mux.Router) {

		r := factory.NewReduction()

		r.Builder().State(reflect.TypeFor[Cart]()).Init(func(ctx context.Context) (*Cart, error) {
			return &Cart{ItemCount: 0}, nil
		}).Action(reflect.TypeFor[AddItem](), func(ctx context.Context, state *Cart, action *AddItem) (*Cart, error) {
			state.ItemCount = state.ItemCount + action.Qty
			return state, nil
		})

		check(rmux.Mount(router, "", r))
	}))))

}

func Test_Cart_NoInit(t *testing.T) {

	t.Run(support.Wrap(func(client *resty.Client) {

		newCart := func() *Cart {
			return &Cart{}
		}

		cart := newCart()

		resp, err := client.R().SetResult(cart).Get("/cart")
		check(err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, &Cart{
			ItemCount: 0,
		}, cart)

		cart = newCart()

		resp, err = client.R().SetResult(cart).SetBody(&AddItem{
			Qty: 10,
		}).Post("/cart/add_item")
		check(err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, &Cart{
			ItemCount: 10,
		}, cart)

		resp, err = client.R().SetResult(cart).SetBody(&AddItem{
			Qty: 10,
		}).Post("/cart/add_item")
		check(err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, &Cart{
			ItemCount: 20,
		}, cart)

	}, fx.Module("fixture", fx.Invoke(func(factory reduction.Factory, router *mux.Router) {

		r := factory.NewReduction()

		r.Builder().State(reflect.TypeFor[Cart]()).
			Action(reflect.TypeFor[AddItem](), func(ctx context.Context, state *Cart, action *AddItem) (*Cart, error) {
				state.ItemCount = state.ItemCount + action.Qty
				return state, nil
			})

		check(rmux.Mount(router, "", r))
	}))))

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
