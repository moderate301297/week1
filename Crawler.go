package main

import (
    "bytes"
    "errors"
    "fmt"
    "golang.org/x/net/html"
	"io"
	"io/ioutil"
	"strings"
	"log"
	"net/http"
    "os"
    "os/signal"
    "gopkg.in/mgo.v2"
	"gopkg.in/robfig/cron.v2"
    "time"
)
//
type DataWeb struct {
	Title      string
	Body       string
	Timestamp  time.Time
}
var (
	IsDrop = true
)
// get title
func getTitle(doc *html.Node) (*html.Node, error) {
    var b *html.Node
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "title" {
            b = n
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)
    if b != nil {
        return b, nil
    }
    return nil, errors.New("Missing body the node tree")
}
// get body 
func getBody(doc *html.Node) (*html.Node, error) {
    var b *html.Node
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "body" {
            b = n
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)
    if b != nil {
        return b, nil
    }
    return nil, errors.New("Missing body the node tree")
}
// type html.Node into type string
func renderNode(n *html.Node) string {
    var buf bytes.Buffer
    w := io.Writer(&buf)
    html.Render(w, n)
    return buf.String()
}
// crawl website
func crawl() (title string, body string){
    var b string
	response, err := http.Get("https://medium.com/topic/programming")
	if err != nil {
			log.Fatal(err)
	} else {
        body,_ := ioutil.ReadAll(response.Body)
        b = string(body)
        defer response.Body.Close()      
    }
    // parse
    doc, _ := html.Parse(strings.NewReader(b))
    bnTitle, err := getTitle(doc)
    if err != nil {
        return
    }
    gtitle := renderNode(bnTitle)
    bnBody, err := getBody(doc)
    if err != nil {
        return
    }
    gbody := renderNode(bnBody)

    return gtitle, gbody
}
// save data in mongodb
 func saveData()  {
    session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
    defer session.Close()
    
	session.SetMode(mgo.Monotonic, true)
	if IsDrop {
		err = session.DB("dataweb").DropDatabase()
		if err != nil {
			panic(err)
		}
    }
    //
	c := session.DB("dataweb").C("data")
	index := mgo.Index{
		Key:        []string{"title"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
    }
    //
	err = c.EnsureIndex(index)
	if err != nil {
		panic(err)
    }
    //
    title, body := crawl()
	// Insert Datas
	err = c.Insert(&DataWeb{Title: title , Body: body , Timestamp: time.Now()})
	if err != nil {
		panic(err)
    } 
 }
 //
 func setMain() {
     saveData()
     fmt.Println("-->")
 }
 //
 func main() {
     //cron job
    cron := cron.New()
    cron.AddFunc("0 0 5 * * *", setMain)
    go cron.Start()
    sig := make(chan os.Signal)
    signal.Notify(sig, os.Interrupt, os.Kill)
    <-sig
 }