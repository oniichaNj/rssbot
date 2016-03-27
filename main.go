package main

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"net/http"
	"os"
	"time"
	"encoding/json"
	"crypto/tls"
	"strings"
	"github.com/SlyMarbo/rss"
	"io/ioutil"
)

type Config struct {
	Server             string
	SSL                bool
	InsecureSkipVerify bool
	Channels           []string
	Realname           string
	Nick               string
	Prefix             string
	RSS                []string
}

func main() {
	file, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	conn := irc.IRC(config.Nick, config.Realname)
	if config.SSL {
		conn.UseTLS = true
		if config.InsecureSkipVerify {
			conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}

	var feeds []rss.Feed
	for _, rssfeed := range config.RSS {
		feed, err := rss.Fetch(rssfeed)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		feeds = append(feeds, *feed)
	}

	for i, _ := range feeds {
		feeds[i].Update()
	}
	
	// Connect to IRC server.
	err = conn.Connect(config.Server)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Join IRC channels.
	// 001 is the IRC status for "Welcome". That means, the func is ran upon successful server connect.
	conn.AddCallback("001", func(e *irc.Event) {
		for _, channel := range config.Channels {
			fmt.Println("Joining", channel)
			conn.Join(channel)
		}
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		go func(e *irc.Event){
			if strings.HasPrefix(e.Message(), config.Prefix) {
				if stringInSlice(e.Arguments[0], config.Channels) {
					for i, _ := range feeds {
						feeds[i].Update()
					}
					// beware, magic numbers
					for _, feed := range feeds {
						if feed.Unread < 5 {
							for i, _ := range feed.Items {
								s, err := urlshorten(feed.Items[i].Link)
								if err != nil {
									fmt.Println("Couldn't shorten url: ", err)
								} else {
									conn.Privmsg(e.Arguments[0], feed.Items[i].Title + " - " + s)
								}
							}
						}
					}
				}
			}
		}(e)
	});

	ticker := time.NewTicker(3 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <- ticker.C:
				for i, _ := range feeds {
					feeds[i].Update()
				}
				for _, feed := range feeds {
					if feed.Unread < 10 {
						for i, _ := range feed.Items {
							for _, channel := range config.Channels {
								s, err := urlshorten(feed.Items[i].Link)
								if err != nil {
									fmt.Println("Couldn't shorten url: ", err)
								} else {
									conn.Privmsg(channel, feed.Items[i].Title + " - " + s)
								}
							}
						}
					}
				}		
			case <- quit:
				ticker.Stop()
				return
			}
		}
	}()
	conn.Loop()
}

func urlshorten(url string) (string, error) {
	//https://is.gd/create.php?format=simple&url=www.example.com
	resp, err := http.Get("https://is.gd/create.php?format=simple&url=" + url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
