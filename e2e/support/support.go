package support

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"go.uber.org/fx"
	"net"
	"net/http"
	"testing"
	"time"
)

func Wrap(callback func(client *resty.Client), options ...fx.Option) (string, func(t *testing.T)) {

	var rs RunningServer

	opts := append(options, NewModule())
	opts = append(opts, fx.Populate(&rs))

	return "default", func(t *testing.T) {

		ctx := context.Background()
		app := fx.New(opts...)

		check(app.Start(ctx))

		r := resty.New().SetBaseURL(fmt.Sprintf("http://%s:%d", rs.Host(), rs.Port()))

		callback(r)

		check(app.Stop(ctx))

	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type RunningServer interface {
	Host() string
	Port() int
}

type serverDesc struct {
	host string
	port int
}

func (s *serverDesc) Host() string {
	return s.host
}

func (s *serverDesc) Port() int {
	return s.port
}

func NewRouter() *mux.Router {

	m := mux.NewRouter()

	m.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	return m
}

func NewServer(lifecycle fx.Lifecycle, r *mux.Router) RunningServer {

	d := &serverDesc{}

	s := http.Server{
		Handler: r,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			d.host = "127.0.0.1"
			l, err := net.Listen("tcp", fmt.Sprintf("%s:0", d.host))

			d.port = l.Addr().(*net.TCPAddr).Port
			check(err)

			go func() {
				s.Serve(l)
			}()

			return waitForHealthCheck(d.host, d.port)
		},
		OnStop: func(ctx context.Context) error {
			s.Shutdown(ctx)
			return nil
		},
	})

	return d
}

func NewModule() fx.Option {
	return fx.Module("e2e.fixture", fx.Provide(
		NewRouter,
		NewServer,
	))
}

func waitForHealthCheck(host string, port int) error {

	r := resty.New().SetBaseURL(fmt.Sprintf("http://%s:%d", host, port)).R()

	for i := 0; i < 30; i++ {
		resp, err := r.Get("/healthz")
		check(err)
		if resp.StatusCode() == http.StatusOK {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return errors.New("unable to check health")
}
