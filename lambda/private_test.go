package main

import (
	"net/http"

	"github.com/mackee/tanukirpc"
)

var NewRegistryFactory = newRegistryFactory

func NewBase64Secret(b []byte) base64secret {
	return base64secret{secret: b}
}

func NewTestHandler(opts *Options) (http.Handler, error) {
	rf, err := newRegistryFactory(opts)
	if err != nil {
		return nil, err
	}

	r := tanukirpc.NewRouter(&registry{},
		tanukirpc.WithContextFactory(
			tanukirpc.NewContextHookFactory(rf.NewRegistry),
		),
	)
	r.Get("/ping", tanukirpc.NewHandler(
		func(ctx tanukirpc.Context[*registry], _ struct{}) (struct{}, error) {
			return struct{}{}, nil
		},
	))
	r.Get("/setup-session", tanukirpc.NewHandler(
		func(ctx tanukirpc.Context[*registry], _ struct{}) (struct{}, error) {
			reg := ctx.Registry()
			if err := reg.sessionAccessor.Set("email", "user@example.com"); err != nil {
				return struct{}{}, err
			}
			if err := reg.sessionAccessor.Save(ctx); err != nil {
				return struct{}{}, err
			}
			return struct{}{}, nil
		},
	))

	return r, nil
}
