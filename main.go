package main

//go:generate gopherjs build -m main.go -o js/index.js

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gopherjs/gopherjs/js"
	vue "github.com/oskca/gopherjs-vue"
	"github.com/pmalek/gopherjs-vue/news"
)

const NEWSAPIKEY = ""

type Article struct {
	*js.Object

	Title   string `js:"title"`
	URL     string `js:"url"`
	Author  string `js:"author"`
	Content string `js:"content"`
}

type Model struct {
	*js.Object // needed for bidirectional data bindings

	Articles   []*Article `js:"articles"`
	IsFetching bool       `js:"isFetching"`
}

func (m *Model) Fetch() {
	const pageSize = 3
	const count = 3

	m.IsFetching = true

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(count)
		ch := make(chan []news.Article)

		go func() {
			articles := vue.GetVM(m).Get("articles")
			for arts := range ch {
				for _, a := range arts {
					i := &Article{
						Object: js.Global.Get("Object").New(),
					}
					i.Title = a.Title
					i.URL = a.URL
					i.Author = a.Author
					i.Content = a.Content

					vue.Push(articles, i)
					time.Sleep(400 * time.Millisecond)
				}
				wg.Done()
			}
		}()

		for i := 1; i <= count; i++ {
			go func(i int) {
				url := "https://newsapi.org/v2/everything?" +
					"q=bitcoin&" +
					"from=2018-11-24&" +
					"sortBy=popularity&" +
					"language=en&" +
					"pageSize=" + strconv.Itoa(pageSize) + "&" +
					"page=" + strconv.Itoa(i) + "&" +
					"apiKey=" + NEWSAPIKEY
				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					println(err)
					wg.Done()
					return
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					println(err)
					wg.Done()
					return
				}
				defer resp.Body.Close()

				var nResp news.Response
				if err := json.NewDecoder(resp.Body).Decode(&nResp); err != nil {
					println(err)
					wg.Done()
					return
				}

				println("Received ", len(nResp.Articles), " articles")

				ch <- nResp.Articles
			}(i)
		}

		wg.Wait()
		close(ch)
		println("Done waiting")
		m.IsFetching = false
	}()
}

func main() {
	m := &Model{
		Object: js.Global.Get("Object").New(),
		// field assignment is required in this way to make data passing works
	}
	m.Articles = nil
	m.IsFetching = false

	// create the VueJS viewModel using a struct pointer
	vue.New("#app", m)
}
