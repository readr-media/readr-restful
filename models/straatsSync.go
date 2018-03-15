package models

import (
	"log"
	"strings"
	"time"

	"github.com/robfig/cron"
)

type straatsSync struct {
	DateLimit        time.Time
	MonitoredLives   map[string]straatsLive
	CronInstance     *cron.Cron
	SubCronRunning   bool
	SubCronInitiated bool
}

func (v *straatsSync) Cron() {
	if !v.SubCronInitiated {
		v.CronInstance = cron.New()
		v.CronInstance.AddFunc("@every 1m", v.SubCron)
		v.SubCronInitiated = true
		v.MonitoredLives = make(map[string]straatsLive)
	}

	v.CheckLivesStatus()

	if len(v.MonitoredLives) > 0 && !v.SubCronRunning {
		v.CronInstance.Start()
	}
}

func (v *straatsSync) SubCron() {
	if len(v.MonitoredLives) == 0 {
		v.CronInstance.Stop()
		return
	}

	v.CheckLivesStatus()
}

func (v *straatsSync) CheckLivesStatus() {

	t := v.GetLatestVideoTime()
	lives, err := v.GetLives(t)
	if err != nil {
		log.Println("Fail Getting Live List: %s", err.Error())
		return
	}

OuterLoop:
	for _, live := range lives {
		live_start_time, _ := time.Parse(time.RFC3339, live.StartTime)
		if live.Status == "ready" &&
			time.Duration(30*time.Minute) > live_start_time.Sub(time.Now()) &&
			time.Duration(24*30*time.Hour) > time.Now().Sub(live_start_time) {
			if _, ok := v.MonitoredLives[live.ID]; !ok || v.MonitoredLives[live.ID].Status != "ready" {
				err := v.UpdateLives(live)
				if err != nil {
					log.Printf("Error Update lives: %s", err.Error())
					continue
				}
				v.MonitoredLives[live.ID] = live
			}
		} else if live.Status == "started" {
			if _, ok := v.MonitoredLives[live.ID]; !ok || v.MonitoredLives[live.ID].Status != "started" {
				err := v.UpdateLives(live)
				if err != nil {
					log.Printf("Error Update lives: %s", err.Error())
					continue
				}
				v.MonitoredLives[live.ID] = live
			}
		} else if live.Status == "ended" {
			vods, err := v.GetVods(live.ID)
			if err != nil {
				continue
			}
			if len(vods) == 0 {
				if _, ok := v.MonitoredLives[live.ID]; !ok || v.MonitoredLives[live.ID].Status != "ended" {
					err := v.UpdateLives(live)
					if err != nil {
						log.Printf("Error Update lives: %s", err.Error())
						continue OuterLoop
					}
				}
			} else {
				for _, vod := range vods {
					if !vod.Ready {
						err := v.UpdateLives(live)
						if err != nil {
							log.Printf("Error Update lives: %s", err.Error())
							continue OuterLoop
						}
						continue OuterLoop
					}
				}
				err := v.UpdateLivesVod(live, vods)
				if err != nil {
					log.Printf("Error Update lives: %s", err.Error())
					continue
				}
				delete(v.MonitoredLives, live.ID)
			}
		}
	}
}
func (v *straatsSync) GetLatestVideoTime() (t time.Time) {
	var time NullTime
	query := "SELECT updated_at FROM posts WHERE type=? ORDER BY updated_at DESC;"
	_ = DB.Get(&time, query, PostType["video"])
	return time.Time
}
func (v *straatsSync) GetLives(time time.Time) ([]straatsLive, error) {
	return StraatsHelper.GetLiveList(time)
}
func (v *straatsSync) GetVods(id string) ([]straatsVod, error) {
	return StraatsHelper.GetLiveVideo(id)
}
func (v *straatsSync) UpdateLives(live straatsLive) (err error) {

	var post_id string
	_ = DB.Get(&post_id, "SELECT post_id FROM posts WHERE video_id=?;", live.ID)
	if post_id == "" {
		v.InsertLive(live)
	}

	switch live.Status {
	case "ready":
		_, err := DB.Exec("UPDATE posts SET type=?, link=?, active=?, updated_at=? WHERE video_id=?;", PostType["live"], live.Link, PostStatus["deactive"], time.Now(), live.ID)
		if err != nil {
			return err
		}
	case "started":
		t, _ := time.Parse(time.RFC3339, live.StartedAt)
		_, err := DB.Exec("UPDATE posts SET type=?, link=?, active=?, updated_at=?, published_at=? WHERE video_id=?;", PostType["live"], live.Link, PostStatus["active"], time.Now(), t, live.ID)
		if err != nil {
			return err
		}
	case "ended":
		t, _ := time.Parse(time.RFC3339, live.StartedAt)
		_, err := DB.Exec("UPDATE posts SET type=?, link=?, active=?, updated_at=?, published_at=? WHERE video_id=?;", PostType["live"], live.Link, PostStatus["deactive"], time.Now(), t, live.ID)
		if err != nil {
			return err
		}
	}
	return err
}
func (v *straatsSync) UpdateLivesVod(live straatsLive, vods []straatsVod) (err error) {

	var post_id string
	_ = DB.Get(&post_id, "SELECT post_id FROM posts WHERE video_id=?;", live.ID)
	if post_id == "" {
		v.InsertLive(live)
	}

	var (
		published_at time.Time
		urls         []string
	)
	for _, v := range vods {
		published_at = v.UpdatedAt
		urls = append(urls, v.Link)
	}
	_, err = DB.Exec("UPDATE posts SET type=?, link=?, active=?, updated_at=?, published_at=? WHERE video_id=?;", PostType["video"], strings.Join(urls, ";"), PostStatus["active"], time.Now(), published_at, live.ID)

	return err
}
func (v *straatsSync) InsertLive(live straatsLive) (err error) {
	_, err = DB.NamedExec(`INSERT INTO posts (title, content, video_id) VALUES (:title, :synopsis, :id)`, live)
	if err != nil {
		log.Println("Error InsertLives: ", err)
	}

	return nil
}

var StraatsSync straatsSync
