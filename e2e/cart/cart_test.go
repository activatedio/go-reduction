package cart

import (
	"context"
	"github.com/activatedio/go-reduction"
	"github.com/activatedio/go-reduction/e2e/support"
	rmux "github.com/activatedio/go-reduction/mux"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/swaggest/openapi-go/openapi3"
	"go.uber.org/fx"
	"net/http"
	"reflect"
	"testing"
)

type Cart struct {
	Status    string
	ItemCount int `json:"item_count"`
}

type AddItem struct {
	Qty int `json:"qty"`
}

type Place struct {
	reduction.Empty
}

func Test_Cart_WithInitNoRefresh(t *testing.T) {

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

		resp, err = client.R().SetResult(cart).Post("/cart/place")
		check(err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, &Cart{
			Status:    "Placed",
			ItemCount: 20,
		}, cart)

		swaggerResult := &openapi3.Spec{}

		// Test swagger
		resp, err = client.R().SetResult(swaggerResult).Get("/swagger.json")
		check(err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Len(t, swaggerResult.Paths.MapOfPathItemValues, 3)

	}, fx.Module("fixture", fx.Invoke(func(factory reduction.Factory, router *mux.Router) {

		r := factory.NewReduction()

		r.Builder().State(reflect.TypeFor[Cart]()).Init(func(ctx context.Context) (*Cart, error) {
			return &Cart{ItemCount: 0}, nil
		}).Action(reflect.TypeFor[AddItem](), func(ctx context.Context, state *Cart, action *AddItem) (*Cart, error) {
			state.ItemCount = state.ItemCount + action.Qty
			return state, nil
		}).Action(reflect.TypeFor[Place](), func(ctx context.Context, state *Cart, action *Place) (*Cart, error) {
			state.Status = "Placed"
			return state, nil
		})

		check(rmux.Mount(router, "", r))
	}))))

}

func Test_Cart_NoInitWithRefresh(t *testing.T) {

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
			ItemCount: 21,
		}, cart)

		resp, err = client.R().SetResult(cart).Post("/cart/place")
		check(err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, &Cart{
			Status:    "Placed",
			ItemCount: 22,
		}, cart)

	}, fx.Module("fixture", fx.Invoke(func(factory reduction.Factory, router *mux.Router) {

		r := factory.NewReduction()

		r.Builder().State(reflect.TypeFor[Cart]()).
			Refresh(func(ctx context.Context, state *Cart) (*Cart, error) {
				state.ItemCount = state.ItemCount + 1
				return state, nil
			}).
			Action(reflect.TypeFor[AddItem](), func(ctx context.Context, state *Cart, action *AddItem) (*Cart, error) {
				state.ItemCount = state.ItemCount + action.Qty
				return state, nil
			}).Action(reflect.TypeFor[Place](), func(ctx context.Context, state *Cart, action *Place) (*Cart, error) {
			state.Status = "Placed"
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
