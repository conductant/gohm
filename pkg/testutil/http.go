package testutil

import (
	. "gopkg.in/check.v1"
	"io/ioutil"
	"net/http"
)

func Get(c *C, url string, checker func(*http.Response, []byte)) {
	resp, err := http.Get(url)
	c.Assert(err, IsNil)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	checker(resp, body)
}

func GetWithRequest(c *C, url string, mod func(*http.Request) *http.Request, checker func(*http.Response, []byte)) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	client := new(http.Client)
	resp, err := client.Do(mod(req))
	c.Assert(err, IsNil)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	checker(resp, body)
}
