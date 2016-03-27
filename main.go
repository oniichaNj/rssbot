package main

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"os"
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
				for i, _ := range feeds {
					feeds[i].Update()
				}
				for _, feed := range feeds {
					for i, _ := range feed.Items {
						fmt.Println(feed.Items[i].Title + " - " + feed.Items[i].Link)
					}
				}
			}
		//event.Message() contains the message
		//event.Nick Contains the sender
		//event.Arguments[0] Contains the channel
		}(e)
	});

	conn.Loop()
}
