package main

/*
This is a minimal sample application, demonstrating how to set up an RSS feed
for regular polling of new channels/items.
Build & run with:
 $ go run example.go
*/

package main

import (
 	"github.com/SurgeNews/SurgeServer/scrapper"
	"crypto/tls"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	longTimeForm  = "2006-01-02T15:04:05.000-07:00"
	shortTimeForm = "2006-01-02"
)

var (
	x bool
	m sync.Mutex
	c *sync.Cond

	ticker *time.Ticker
	timer  *time.Timer
	l      bool
	o      bool
)

func main() {
	x = true
	c = sync.NewCond(&m)
	l = true
	o = false


	// set the periodic call to reader()
	ticker = time.NewTicker(time.Second * 110)
	go func() {
		for t := range ticker.C {
			//@tochange reader := new(Reader)
			go reader.read(t)
		}
	}()

	// set the timer for token expiry
	timer = time.NewTimer(time.Second * time.Duration(phClient.ExpirySec-60))
	// set application main loop
	for l {
		select {
		case <-timer.C: // time-out condition
			m.Lock()
			x = false
			timer.Stop()
			// set timer for next token expiry
			timer = time.NewTimer(time.Second * time.Duration(phClient.ExpirySec-60))
			x = true
			c.Broadcast()
			m.Unlock()
		}
	}
}

type Reader struct {
	timeStamp    int32
	burstControl chan bool
	wg           sync.WaitGroup
	feed         *mgo.Collection
	url          *mgo.Collection
}

func (self *Reader) read(t time.Time) {
	// wait if token renewal is in process
	if !x {
		m.Lock()
		m.Unlock()
	}

	if time.Since(t).Seconds() > 30 {
		return // cancel a read operation is it has to wait more than 30 sec
	}

	if o {
		return // cancel a overlapping reads
	}
	o = true

	data, err := phClient.PostsOfTheDay()
	if err != nil {
		log.Printf("\n\tWarning1: %+v\n", err)
		o = false
		return
	}
	self.timeStamp = int32(time.Now().Unix())
	session, err := mgo.Dial("mongodb://localhost/smoothies")
	session.SetMode(mgo.Monotonic, true)
	self.feed = session.DB("smoothies").C("feed")
	self.url = session.DB("smoothies").C("url")
	if err != nil {
		panic(err)
	}

	self.burstControl = make(chan bool, 10)
	for index, feedPost := range data.Posts {
		feedEntry := feedPost
		//log.Println(" reading -", feedEntry.Id, feedEntry.Created_at, feedEntry.Day, index)
		post := new(models.Post)
		if feedEntry.Id != 0 && feedEntry.Created_at != "" && feedEntry.Day != "" {
			rank := index + 1
			self.feed.Find(bson.M{"ref_id": "ph-" + strconv.Itoa(feedEntry.Id)}).One(post)
			self.wg.Add(1)
			self.burstControl <- true
			if post.RefId == "" {
				go self.createPost(&feedEntry, rank, post)
			} else {
				go self.modifyPost(rank, post)
			}
		}
	}

	err = self.feed.EnsureIndex(models.PostReadIndex)
	if err != nil {
		log.Printf("\n\tWarning2: %+v\n", err)
	}

	err = self.feed.EnsureIndex(models.PostLookupIndex)
	if err != nil {
		log.Printf("\n\tWarning3: %+v\n", err)
	}

	err = self.feed.EnsureIndex(models.PostDiffIndex)
	if err != nil {
		log.Printf("\n\tWarning4: %+v\n", err)
	}

	self.wg.Wait()
	o = false
	defer session.Close()
	log.Printf("\n\nInfo: Read completed in %s\n", time.Since(t))
}

func (self *Reader) modifyPost(rank int, post *models.Post) {
	log.Println(" check to modify -", post.RefId, rank)
	if post.Rank != rank {
		log.Println("modifing record -", post.RefId, post.Rank, " to ", rank)
		colQuerier := bson.M{"ref_id": post.RefId}
		change := bson.M{"$set": bson.M{"secondery_index": rank, "epoch_time_modified": self.timeStamp}}
		err := self.feed.Update(colQuerier, change)
		if err != nil {
			log.Printf("\n\tWarning5: %+v\n", err)
		}
	}
	<-self.burstControl
	self.wg.Done()
}

func (self *Reader) createPost(feedEntry *ph.PostResponceBody, rank int, post *models.Post) {
	var err error
	post.RefId = "ph-" + strconv.Itoa(feedEntry.Id)
	post.Name = feedEntry.Name
	post.Title = feedEntry.Tagline
	post.Newslink = feedEntry.Discussion_url
	post.Date, _ = time.Parse(shortTimeForm, feedEntry.Day)
	createTS, _ := time.Parse(longTimeForm, feedEntry.Created_at)
	post.CTS = int32(createTS.Unix())
	post.MTS = self.timeStamp
	post.Rank = rank
	post.ScreenImg = feedEntry.Screenshot_url.Small
	// log.Println(" adding -", post.RefId, post.Rank)
	post.Permalink, err = self.getUrl(feedEntry.Redirect_url)
	if err == nil || post.Permalink != "" {
		err = self.feed.Insert(post)
		if err != nil {
			log.Printf("\n\tWarning6: %+v\n", err)
		}
	} else {
		log.Printf("\n\tWarning7: %+v\n", err)
	}
	<-self.burstControl
	self.wg.Done()
}

func (self *Reader) getUrl(link string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: 20 * time.Second,
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Get(link)
	if err != nil {
		return link, err
	}

	url := new(models.URL)
	url.Address = resp.Request.URL.String()
	url.PrepareHash()

	_, err = self.url.Upsert(bson.M{"hash": url.Hash}, url)
	if err != nil {
		return "", err
	}

	err = self.feed.EnsureIndex(models.URLReadIndex)
	if err != nil {
		log.Printf("\n\tWarning8: %+v\n", err)
	}

	return url.Hash, err
}