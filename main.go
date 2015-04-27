package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Filo []string

func (f *Filo) Push(v string) {
	*f = append(*f, v)
}

func (f Filo) Peek() string {
	value := f[len(f)-1]
	return value
}

func (f Filo) nPeek(n int) string {
	value := f[len(f)-n]
	return value
}

func (f *Filo) Pop() string {
	value := (*f)[len(*f)-1]
	(*f) = (*f)[:len(*f)-1]
	return value
}

type Tag struct {
	Attr map[string]int
	Tag  map[string]string
}

func NewTag() Tag {
	return Tag{Attr: make(map[string]int), Tag: make(map[string]string)}
}

func main() {

	URL := ""
	APIKEY := []byte("")

	post := []byte("")
	post = append(post, APIKEY...)

	resp, err := http.Post(URL, "application/x-www-form-urlencoded; charset=UTF-8", bytes.NewReader(post))
	if err != nil {
		log.Fatalln("Unable to perform request")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Unable to fetch body")

	}

	fmt.Println(string(body))
	fmt.Println("\n")

	var currTag Filo
	tags := make(map[string]Tag)

	index := bytes.Index(body, []byte("<RESPONSE"))
	for pos := index; pos < len(body); pos++ {

		if body[pos] == '>' {
			continue
		}

		if body[pos] == '<' && body[pos+1] != '/' { //Start of tag
			pos++
			var tagname []byte
			for ; body[pos] != '>' && body[pos] != ' '; pos++ {
				tagname = append(tagname, body[pos])
			}
			if len(tags) != 0 {
				tags[currTag.Peek()].Tag[string(tagname)] = "string"
			}
			currTag.Push(string(tagname))
			tags[string(tagname)] = NewTag()

			for ; body[pos] == ' '; pos++ { //Next is an attribute
				pos++
				var attrname []byte
				var value []byte
				for ; body[pos] != '='; pos++ {
					attrname = append(attrname, body[pos])
				}

				pos = pos + 2
				for ; body[pos] != '"'; pos++ {
					value = append(value, body[pos])
				}

				tags[string(tagname)].Attr[string(attrname)] = 1
			}

		} else if body[pos] == '<' && body[pos+1] == '/' { //End tag
			pos = pos + 2
			var tagname []byte
			for ; body[pos] != '>' && body[pos] != ' '; pos++ {
				tagname = append(tagname, body[pos])
			}
			n := currTag.Pop()
			if string(tagname) != n {
				fmt.Printf("Parse error. Poped tag not pushed\n")
			}
		} else {

			var value []byte
			for ; body[pos] != '<'; pos++ {
				value = append(value, body[pos])
			}
			if _, err := strconv.Atoi(string(value)); err == nil {
				tags[currTag.nPeek(2)].Tag[currTag.Peek()] = "int64"
			}
			pos--
		}

	}

	for key, tag := range tags {
		if len(tag.Attr) == 0 && len(tag.Tag) == 0 {
			continue
		}

		fmt.Printf("type %s struct {\n", strings.Title(strings.ToLower(key)))

		for k, _ := range tag.Attr {
			fmt.Printf("%s string `xml:\"%s,attr\"`\n", strings.Title(strings.ToLower(k)), k)
		}

		for k, t := range tag.Tag {
			if len(tags[k].Attr) == 0 && len(tags[k].Tag) == 0 {
				fmt.Printf("%s %s `xml:\"%s\"`\n", strings.Title(strings.ToLower(k)), t, k)
			} else {
				fmt.Printf("%s []%s `xml:\"%s\"`\n", strings.Title(strings.ToLower(k)), strings.Title(strings.ToLower(k)), k)
			}
		}

		fmt.Println("}\n")

	}
}
