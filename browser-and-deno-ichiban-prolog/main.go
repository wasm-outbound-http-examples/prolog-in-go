// Based on https://github.com/ichiban/prolog/blob/v1.2.2/examples/call_go_from_prolog/main.go
package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/ichiban/prolog"
	"github.com/ichiban/prolog/engine"
)

func main() {
	p := prolog.New(nil, nil)

	p.Register2(engine.NewAtom("httpget"), func(_ *engine.VM, url, respText engine.Term, k engine.Cont, env *engine.Env) *engine.Promise {
		u, ok := env.Resolve(url).(engine.Atom)
		if !ok {
			return engine.Error(engine.TypeError(engine.NewAtom("atom"), url, env))
		}

		resp, err := http.Get(u.String())
		if err != nil {
			return engine.Error(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		env, ok = env.Unify(respText, engine.NewAtom(string(body)))
		if !ok {
			return engine.Bool(false)
		}

		return k(env)
	})

	if err := p.Exec(`:- set_prolog_flag(double_quotes, atom).`); err != nil {
		panic(err)
	}

	sols, err := p.Query(`httpget("https://httpbin.org/anything", RespTxt).`)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := sols.Close(); err != nil {
			panic(err)
		}
	}()

	if !sols.Next() {
		panic("no solutions")
	}

	var s struct {
		RespTxt string
	}
	if err := sols.Scan(&s); err != nil {
		panic(err)
	}

	fmt.Printf("text: %+v\n", s.RespTxt)
}
