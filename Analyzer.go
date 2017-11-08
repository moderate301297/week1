package main

import(
	"golang.org/x/net/html"
	"gopkg.in/mgo.v2"
	"fmt"
	"log"
	"time"
	"strings"
	"strconv"
	"unicode"
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/robfig/cron.v2"
	"os"
	"os/signal"
)
type DataWeb struct {
	Title      string
	Body       string
	Timestamp  time.Time
}
// query all data from mongodb
func queryAll() {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("dataweb").C("data")
	var results []DataWeb
    err1 := c.Find(nil).All(&results)
	if err1 != nil {
		log.Printf("RunQuery : ERROR : %s\n", err1)
		return
	}
	fmt.Println(results)
}
// query body data form mongodb
func queryOne() (body string){
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("dataweb").C("data")
	results := DataWeb{}
    err1 := c.Find(nil).One(&results)
	if err1 != nil {
		log.Printf("RunQuery : ERROR : %s\n", err1)
		return
	}
	body = results.Body
	return body
}
// get href
func getHref(t *html.Node) (href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
		}
	}
	return
}
// extract url
func extract() (a[500] string, len int) {
	i := 0
	var j int
	str := queryOne()
	doc, _ := html.Parse(strings.NewReader(str))
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "a" {
			url := getHref(n)
			for j = 0; ; {
				if (a[j] == url){
					break
				} else if (j == i){
					a[i] = url
					i = i + 1
				} else {
					j = j + 1
				}
			}
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
	f(doc)
	len = i
	return a, i
}
// get like
func getLike(url string) (c string) {
	resp,_ := http.Get(url)
	b := resp.Body
	defer b.Close()
	z := html.NewTokenizer(b)
	depth := 0
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return
		case html.TextToken:
			if depth > 0 {
				l := z.Text()
				like := string(l)
				c = like
			}
		case html.StartTagToken, html.EndTagToken:
			tn, _ := z.TagName()
			if (len(tn) == 4 && tn[0] == 's') {
				tt := z.Next()
				depth = 0
				switch tt {
				case html.ErrorToken:
					return
				case html.StartTagToken, html.EndTagToken:
					tn, _ := z.TagName()
					if (len(tn) == 6 && tn[0] == 'b') {	
						if tt == html.StartTagToken {
							depth = 1
						} else {
							depth = 0
						}
					}
				}		
			}
		}
	}
	return c
}
// save data to mysql
func saveDaTa() {
	db, err := sql.Open("mysql", "root:30121996@tcp(localhost:3306)/dataweb")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//
	a,len := extract()
	//save length
	rowsLength, err := db.Query("select id from len_url")
	if err != nil {
		panic(err)
	}
	var idLength int
	for rowsLength.Next() {
		err = rowsLength.Scan(&idLength)
		if err != nil {
			panic(err)
		}
	}	
		if(idLength == 1) {
			length, err := db.Prepare("update len_url set length=? where id=?")
			if err != nil {
				panic(err)
			}
			_, err = length.Exec(len,"1")
			if err != nil {
				panic(err)
			}
		} else {
			length, err := db.Prepare("insert len_url set length=?,id=?")
			if err != nil {
				panic(err)
			}
			_, err = length.Exec(len,"1")
			if err != nil {
				panic(err)
			}
		}
	//save url
		rowsURL, err := db.Query("select id from string_url")
		if err != nil {
			panic(err)
		}
		var idURL int
		for rowsURL.Next() {
			err = rowsURL.Scan(&idURL)
			if err != nil {
				panic(err)
			}
		}

		if(idURL >= len) {
			for i := 0; i < len; i++ {
				url, err := db.Prepare("update string_url set arrary_url=? where id=?")
				if err != nil {
					panic(err)
				}
				_, err = url.Exec(a[i],i+1)
				if err != nil {
					panic(err)
				}
			}
		} else {
			for i:= 0; i < idURL; i++ {
				url, err := db.Prepare("update string_url set arrary_url=? where id=?")
				if err != nil {
					panic(err)
				}
				_, err = url.Exec(a[i],i+1)
				if err != nil {
					panic(err)
				}
			}
			for i:= idURL; i < len; i++ {
			url, err := db.Prepare("insert string_url set arrary_url=?,id=?")
			if err != nil {
				panic(err)
			}
			_, err = url.Exec(a[i],i+1)
			if err != nil {
				panic(err)
			}
		}
	}
	//saveTopLike
	b, length := parseTopLike()
	rowsURLLike, err := db.Query("select id from top_like")
	if err != nil {
		panic(err)
	}
	var idURLLike int
	for rowsURLLike.Next() {
		err = rowsURLLike.Scan(&idURLLike)
		if err != nil {
			panic(err)
		}
	}

	if(idURLLike >= length) {
		for i := 0; i < length; i++ {
			url, err := db.Prepare("update top_like set top_url=? where id=?")
			if err != nil {
				panic(err)
			}
			_, err = url.Exec(b[i],i+1)
			if err != nil {
				panic(err)
			}
		}
	} else {
		for i:= 0; i < idURLLike; i++ {
			url, err := db.Prepare("update top_like set top_url=? where id=?")
			if err != nil {
				panic(err)
			}
			_, err = url.Exec(b[i],i+1)
			if err != nil {
				panic(err)
			}
		}
		for i:= idURLLike; i < length; i++ {
		url, err := db.Prepare("insert top_like set top_url=?,id=?")
		if err != nil {
			panic(err)
		}
		_, err = url.Exec(b[i],i+1)
		if err != nil {
			panic(err)
		}
	}
}

}
// clear url advertisement
func clearURL() (b[100] string, j int) {

	a,len := extract()
	j = 0
	for i := 0; i < len; i++ {
		check := strings.Index(a[i], "topic_page") == -1
		if !check {
			b[j] = a[i]
			j ++
		}
	}
	return b, j
}
// parse string
func stripNonIntFloat(s string) string {
    f := func(c rune) bool {
        return !unicode.IsNumber(c) && (c != 46)
    }
    output := strings.FieldsFunc(s, f)
    if len(output) > 0 {
        return output[0]
    } else {
        return ""
    }
}
// get top like
func parseTopLike() (e[10] string, count int){
	var c[100] string
	var d[100] float64
	a,length := clearURL()
	for i:= 0; i < length; i++ {
		b := getLike(a[i])
		c[i] = b
	}
	for i := 0; i < length; i++ {
		check := strings.Index(c[i], "K") == -1
		if !check {
			s := stripNonIntFloat(c[i])
			v, err := strconv.ParseFloat(s, 10)
			if err != nil {
				fmt.Println("ERROR!")
			} else {
				d[i] = v * 1000
			}
		} else {
			v, err := strconv.ParseFloat(c[i], 10)
			if err != nil {
				fmt.Println("ERROR!")
			} else {
				d[i] = v
			}
		}
	}
	//arrange
	for i := 0; i < length; i++ {
		for j := i + 1; j < length; j++{
			if d[j] > d[i] {
				var tmp1 float64
				tmp1 = d[i]
				d[i] = d[j]
				d[j] = tmp1
				var tmp2 string
				tmp2 = a[i]
				a[i] = a[j]
				a[j] = tmp2
			}
		}
	}
	if (length >= 10) {
		for count = 0; count < 10; count++ {
			e[count] = a[count]
		}
	} else {
		for count = 0; count < length; count++ {
			e[count] = a[count]
		}
	}
	return e, count
}
func setMain() {
	saveDaTa()
	fmt.Println("-->")
}
func main() {
	//cron job
	cron := cron.New()
    cron.AddFunc("0 0 7 * * *", setMain)
    go cron.Start()
    sig := make(chan os.Signal)
    signal.Notify(sig, os.Interrupt, os.Kill)
    <-sig
}