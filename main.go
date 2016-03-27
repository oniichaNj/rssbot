package main

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"os"
	"time"
	"encoding/json"
	"crypto/tls"
	"strings"
	"github.com/SlyMarbo/rss"
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

	var feeds [](*rss.Feed)
	for _, rssfeed := range config.RSS {
		feed, err := rss.Fetch(rssfeed)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		feeds = append(feeds, feed)
	}
	fmt.Printf("%p\n", feeds[0])
	for _, f := range feeds {
		f.Update()
		for _, item := range f.Items {
			fmt.Printf("%s - %s\n", item.Title, item.Link)
			f.Unread--
		}
		fmt.Println("Unread counters: ", f.Unread)		
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
					for _, f := range feeds {
						f.Update()
						fmt.Println("Just updated - new unread count is ", f.Unread)
						if f.Unread > 0 {
							// at least one more new item
							for i := 0; i < int(f.Unread); i++ {
								conn.Privmsg(e.Arguments[0], fmt.Sprintf("%s - %s\n", f.Items[i].Title, f.Items[i].Link))
								
							}
							f.Unread = 0
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
				for _, f := range feeds {
					fmt.Println("Pre update - unread is", f.Unread)
					f.Update()
					fmt.Println("Post update - unread is", f.Unread)
					if f.Unread > 0 {
						// at least one more new item
						for i := 0; i < int(f.Unread); i++ {
							for _, channel := range config.Channels {
								conn.Privmsg(channel, fmt.Sprintf("%s - %s\n", f.Items[i].Title, f.Items[i].Link))
							}
						}
						f.Unread = 0
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
/*
func urlshorten(url string) (string, error) {
	//Turns out these guys rate limit aggressively.
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
*/
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
