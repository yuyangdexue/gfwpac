package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"
)

func main() {
	flagTemplate := flag.String("template", "template.pac", "template file")
	flagProxies := flag.String(
		"proxies",
		"SOCKS5 127.0.0.1:1081; "+
			"SOCKS 127.0.0.1:1081; "+
			"SOCKS5 127.0.0.1:1080; "+
			"SOCKS 127.0.0.1:1080; "+
			"DIRECT",
		"proxies",
	)
	flagOutput := flag.String("output", "gfwpac", "output file")
	flag.Parse()

	b, err := ioutil.ReadFile(*flagTemplate)
	if err != nil {
		panic(err)
	}

	t := template.Must(template.New("pac").Parse(string(b)))

	r, err := http.Get("https://git.io/gfwlist")
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	d, err := ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, r.Body))
	if err != nil {
		panic(err)
	}

	rs := strings.Split(string(d), "\n")
	rs = rs[1 : len(rs)-1]
	for i := 0; i < len(rs); i++ {
		if len(rs[i]) == 0 || rs[i][0] == '!' {
			rs = append(rs[:i], rs[i+1:]...)
			i--
		} else {
			rs[i] = "\"" + rs[i] + "\""
		}
	}

	buf := &bytes.Buffer{}
	t.Execute(buf, map[string]interface{}{
		"Rules":   "[" + strings.Join(rs, ",") + "]",
		"Proxies": "\"" + *flagProxies + "\"",
	})

	buf2 := &bytes.Buffer{}
	js.DefaultMinifier.Minify(minify.New(), buf2, buf, nil)

	if err := ioutil.WriteFile(
		*flagOutput,
		buf2.Bytes(),
		0644,
	); err != nil {
		panic(err)
	}
}
