package models

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"encoding/json"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

type httpReqArgs struct {
	method string
	url    string
	body   string
}

type straatsHelper struct {
	appID  string
	appKey string
	apiKey string
	apiUrl string
	inited bool
}

type straatsLive struct {
	ID        string `json:"id" db:"id"`
	Title     string `json:"title" db:"title"`
	Content   string `json:"synopsis" db:"synopsis"`
	StartTime string `json:"start_time"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
	Status    string `json:"status"`
	Link      string `json:"embed_url"`
}

type straatsVod struct {
	ID        string    `json:"id"`
	Ready     bool      `json:"accomplished"`
	Available bool      `json:"available"`
	Link      string    `json:"embed_url"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *straatsHelper) Init() (err error) {
	if !s.inited {
		s.apiUrl = fmt.Sprintf("%s%s", viper.Get("straats.api_server"), "/v1")
		s.appID = viper.Get("straats.app_id").(string)
		s.appKey = viper.Get("straats.app_key").(string)
		key, err := s.getAPIKey()
		if err != nil {
			log.Printf("Init Straats API Error: %s", err.Error())
			return err
		}
		s.apiKey = key
		s.inited = true
	}
	return nil
}
func (s *straatsHelper) GetLiveVideo(id string) (vodList []straatsVod, err error) {
	resp, err := s.makeRequest(httpReqArgs{"GET", fmt.Sprintf("/app/lives/%s/videos", id), ``})
	if err != nil {
		return vodList, err
	}
	total, _ := strconv.Atoi(resp.Header.Get("total"))
	if total == 0 {
		return vodList, nil
	}

	if err = json.NewDecoder(resp.Body).Decode(&vodList); err != nil {
		log.Println("Error scanning Vod list: ", err)
		return vodList, err
	}

	return vodList, nil
}

func (s *straatsHelper) GetLiveList(t time.Time) (liveList []straatsLive, err error) {
	var (
		page     int  = 1
		per_page int  = 20
		looping  bool = true
	)

	for looping {
		resp, err := s.makeRequest(httpReqArgs{"GET", fmt.Sprintf("/app/lives?page=%d&per_page=%d&sort=-created_at", page, per_page), ``})
		if err != nil {
			return liveList, err
		}
		total, _ := strconv.Atoi(resp.Header.Get("total"))

		tmpLiveList := []straatsLive{}
		if err = json.NewDecoder(resp.Body).Decode(&tmpLiveList); err != nil {
			log.Println("Error scanning Live list: ", err)
			return liveList, err
		}
		for _, v := range tmpLiveList {
			if v.Status == "ended" {
				livet, _ := time.Parse(time.RFC3339, v.EndedAt)
				if livet.After(t) {
					liveList = append(liveList, v)
				}
				if livet.Before(t) {
					looping = false
				}
			} else {
				liveList = append(liveList, v)
			}

			if per_page*page > total {
				looping = false
			}
		}
		page++
	}
	return
}
func (s *straatsHelper) getAPIKey() (key string, err error) {

	o := struct {
		Key     string `json:"token"`
		Account string `json:"account_id"`
	}{}

	resp, err := s.makeRequest(httpReqArgs{"POST", "/app/token", fmt.Sprintf(`{
		"client_id": "%s",
		"client_secret": "%s"
	}`, s.appID, s.appKey)})
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
		log.Println("Error scanning Straats req: ", err)
		return "", err
	}

	return o.Key, nil
}
func (s *straatsHelper) refreshAPIKey() (err error) {
	key, err := s.getAPIKey()
	if err != nil {
		return err
	}
	s.apiKey = key
	return nil
}

func (s *straatsHelper) makeRequest(args httpReqArgs) (resp *http.Response, err error) {

	client := &http.Client{}
	jsonStr := []byte{}
	if args.body != "" {
		jsonStr = []byte(args.body)
	}

	req, err := http.NewRequest(args.method, fmt.Sprintf("%s%s", s.apiUrl, args.url), bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println("Error building Straats req: ", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	}

	resp, err = client.Do(req)
	if err != nil {
		log.Println("Error making Straats req: ", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {

		log.Println(resp.StatusCode)
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(responseData))
		log.Println(args)

		decoder := json.NewDecoder(resp.Body)
		e := struct {
			Error string `json:"error"`
		}{}
		if err := decoder.Decode(&e); err != nil {
			log.Println("Error scanning Straats error: ", err, " Error", e)
			return nil, err
		}

		if e.Error == "401 JWT expired" {
			//Refresh API Key
			s.apiKey = ""
			err = s.refreshAPIKey()
			if err != nil {
				log.Printf("API Key refresh error: %s", err.Error())
			}
			return s.makeRequest(args)
		}

		return nil, errors.New(fmt.Sprintf("Error Straats API Response: %s", e.Error))

		return nil, errors.New("")
	}

	return resp, nil
}

var StraatsHelper straatsHelper
