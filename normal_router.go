package cedar

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var FileType = map[string]string{"html": "text/html", "css": "text/css", "txt": "text/plain", "zip": "application/x-zip-compressed", "png": "image/png", "jpg": "image/jpeg"}

type Groups struct {
	Tree *Trie
	Path string
}

func writeStaticFile(path string, filename []string, w http.ResponseWriter) {

	if pusher, ok := w.(http.Pusher); ok {
		//Push is supported.
		options := &http.PushOptions{
			Header: http.Header{
				"Accept-Encoding": {"Content-Type:" + FileType[filename[len(filename)-1]]},
			},
		}
		if err := pusher.Push("."+path, options); err != nil {
			goto end
		}
	} else {
		goto end
	}
end:
	w.Header().Set("Content-Type", FileType[filename[len(filename)-1]])
	fs, err := os.OpenFile("."+path, os.O_RDONLY, 0666)
	if err != nil {
		log.Println(err)
	}
	data, err := ioutil.ReadAll(fs)
	if err != nil {
		log.Println(err)
	}
	_, err = w.Write(data)
}
func (mux *Trie) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) > 7 && r.URL.Path[1:7] == "static" {
		filename := SplitString([]byte(r.URL.Path[8:]), []byte("."))
		writeStaticFile(r.URL.Path, filename, w)
		return
	}
	go func() {
		for k, v := range mux.globalFunc {
			if err := v.Fn(w, r); err != nil {
				log.Panicln(k, err)
			}
		}
	}()
	me, handf, hand := mux.Find(r.URL.Path)
	log.Println(me, r.URL.Path)
	if r.Method != me {
		w.Header().Set("Content-type", "text/html")
		w.Header().Set("charset", "UTF-8")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("<span style=\"font-size=500px\">404</span>"))
		return
	}
	if hand != nil {
		hand.ServeHTTP(w, r)
	}
	if handf != nil {
		handf(w, r)
	}

}

func (mux *Trie) Group(path string, fn func(groups *Groups)) {
	g := new(Groups)
	g.Tree = mux
	g.Path = path
	fn(g)
}
func (mux *Trie) Template(w http.ResponseWriter, path string) {
	writeStaticFile(path+".html", []string{"", "html"}, w)
}
func (mux *Trie) GlobalFunc(name string, fn func(w http.ResponseWriter, r *http.Request) error) {
	mux.globalFunc = append(mux.globalFunc, &GlobalFunc{
		Name: name,
		Fn:   fn,
	})
}
func (mux *Groups) Get(path string, handlerFunc http.HandlerFunc, handler http.Handler) {
	mux.Tree.Get(mux.Path+path, handlerFunc, handler)
}
func (mux *Groups) Post(path string, handlerFunc http.HandlerFunc, handler http.Handler) {
	mux.Tree.Post(mux.Path+path, handlerFunc, handler)
}
func (mux *Groups) Put(path string, handlerFunc http.HandlerFunc, handler http.Handler) {
	mux.Tree.Put(mux.Path+path, handlerFunc, handler)
}
func (mux *Groups) Delete(path string, handlerFunc http.HandlerFunc, handler http.Handler) {
	mux.Tree.Delete(mux.Path+path, handlerFunc, handler)
}
func (mux *Groups) Group(path string, fn func(groups *Groups)) {
	g := new(Groups)
	g.Path = mux.Path + path
	g.Tree = mux.Tree
	fn(g)
}
func (mux *Trie) Get(path string, handlerFunc http.HandlerFunc, handler http.Handler) {
	mux.Insert(http.MethodGet, path, handlerFunc, handler)
}
func (mux *Trie) Post(path string, handlerFunc http.HandlerFunc, handler http.Handler) {
	mux.Insert(http.MethodPost, path, handlerFunc, handler)
}
func (mux *Trie) Put(path string, handlerFunc http.HandlerFunc, handler http.Handler) {
	mux.Insert(http.MethodPut, path, handlerFunc, handler)
}
func (mux *Trie) Delete(path string, handlerFunc http.HandlerFunc, handler http.Handler) {
	mux.Insert(http.MethodDelete, path, handlerFunc, handler)
}
