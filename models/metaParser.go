package models

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crypto/tls"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/readr-media/readr-restful/config"
)

type OGInfo struct {
	Title       string `meta:"og:title" json:"og_title"`
	Description string `meta:"og:description" json:"og_description"`
	Image       string `meta:"og:image,og:image:url" json:"og_image"`
	SiteName    string `meta:"og:site_name" json:"og_site_name"`
}

type ogParser struct{}

func (o *ogParser) GetOGInfoFromUrl(urlStr string) (*OGInfo, error) {

	tr := &http.Transport{TLSClientConfig: &tls.Config{
		NextProtos: []string{"h1"},
	}}

	client := &http.Client{Transport: tr, Timeout: time.Duration(5 * time.Second)}

	req, err := http.NewRequest("GET", urlStr, nil)

	// for k, v := range OGParserHeaders {
	for k, v := range config.Config.Crawler.Headers {
		req.Header.Add(k, v)
	}

	if !regexp.MustCompile("\\.readr\\.tw\\/").MatchString(urlStr) {
		req.Header.Del("Cookie")
	}

	if regexp.MustCompile("\\.youtube\\.com\\/").MatchString(urlStr) {
		req.Header.Del("User-Agent")
		req.Header.Add("User-Agent", "facebookexternalhit/1.1")
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	return o.GetPageInfoFromResponse(resp)
}

func (o *ogParser) GetPageInfoFromResponse(response *http.Response) (*OGInfo, error) {
	info := OGInfo{}
	html, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	err = o.GetPageDataFromHtml(html, &info)

	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (o *ogParser) GetPageDataFromHtml(html []byte, data interface{}) error {
	buf := bytes.NewBuffer(html)
	doc, err := goquery.NewDocumentFromReader(buf)

	if err != nil {
		return err
	}

	return o.getPageData(doc, data)
}

func (o *ogParser) getPageData(doc *goquery.Document, data interface{}) error {
	var rv reflect.Value
	var ok bool
	if rv, ok = data.(reflect.Value); !ok {
		rv = reflect.ValueOf(data)
	}

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("Should not be non-ptr or nil")
	}

	rt := rv.Type()

	for i := 0; i < rv.Elem().NumField(); i++ {
		fv := rv.Elem().Field(i)
		field := rt.Elem().Field(i)

		switch fv.Type().Kind() {
		case reflect.Ptr:
			if fv.IsNil() {
				fv.Set(reflect.New(fv.Type().Elem()))
			}
			e := o.getPageData(doc, fv)

			if e != nil {
				return e
			}
		case reflect.Struct:
			e := o.getPageData(doc, fv.Addr())

			if e != nil {
				return e
			}
		case reflect.Slice:
			if fv.IsNil() {
				fv.Set(reflect.MakeSlice(fv.Type(), 0, 0))
			}

			switch field.Type.Elem().Kind() {
			case reflect.Struct:
				last := reflect.New(field.Type.Elem())
				for {
					data := reflect.New(field.Type.Elem())
					e := o.getPageData(doc, data.Interface())

					if e != nil {
						return e
					}

					//Ugly solution (I can't remove nodes. Why?)
					if !reflect.DeepEqual(last.Elem().Interface(), data.Elem().Interface()) {
						fv.Set(reflect.Append(fv, data.Elem()))
						last.Elem().Set(data.Elem())

					} else {
						break
					}
				}
			case reflect.Ptr:
				last := reflect.New(field.Type.Elem().Elem())
				for {
					data := reflect.New(field.Type.Elem().Elem())
					e := o.getPageData(doc, data.Interface())

					if e != nil {
						return e
					}

					//Ugly solution (I can't remove nodes. Why?)
					if !reflect.DeepEqual(last.Elem().Interface(), data.Elem().Interface()) {
						fv.Set(reflect.Append(fv, data))
						last.Elem().Set(data.Elem())

					} else {
						break
					}
				}
			default:
				if tag, ok := field.Tag.Lookup("meta"); ok {
					tags := strings.Split(tag, ",")

					for _, t := range tags {
						contents := []reflect.Value{}

						processMeta := func(idx int, sel *goquery.Selection) {
							if c, existed := sel.Attr("content"); existed {
								if field.Type.Elem().Kind() == reflect.String {
									contents = append(contents, reflect.ValueOf(c))
								} else {
									i, e := strconv.Atoi(c)

									if e == nil {
										contents = append(contents, reflect.ValueOf(i))
									}
								}

								fv.Set(reflect.Append(fv, contents...))
							}
						}

						doc.Find(fmt.Sprintf("meta[property=\"%s\"]", t)).Each(processMeta)

						doc.Find(fmt.Sprintf("meta[name=\"%s\"]", t)).Each(processMeta)

						fv = reflect.Append(fv, contents...)
					}
				}
			}
		default:
			if tag, ok := field.Tag.Lookup("meta"); ok {

				tags := strings.Split(tag, ",")

				content := ""
				existed := false
				sel := (*goquery.Selection)(nil)
				for _, t := range tags {
					if sel = doc.Find(fmt.Sprintf("meta[property=\"%s\"]", t)).First(); sel.Size() > 0 {
						content, existed = sel.Attr("content")
					}

					if !existed {
						if sel = doc.Find(fmt.Sprintf("meta[name=\"%s\"]", t)).First(); sel.Size() > 0 {
							content, existed = sel.Attr("content")
						}
					}

					if !existed && t == "og:title" {
						if sel = doc.Find(fmt.Sprintf("title")).First(); sel.Size() > 0 {
							existed = true
							content = sel.Text()
						}
					}

					if existed {
						if fv.Type().Kind() == reflect.String {
							fv.Set(reflect.ValueOf(content))
						} else if fv.Type().Kind() == reflect.Int {
							if i, e := strconv.Atoi(content); e == nil {
								fv.Set(reflect.ValueOf(i))
							}
						}
						break
					}
				}
			}
		}
	}
	return nil
}

var OGParser ogParser

// var OGParserHeaders map[string]string
