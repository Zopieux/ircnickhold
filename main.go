package main

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type AuthedEvent struct{}
type GotNicksEvent struct{}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	ircobj := irc.IRC(fmt.Sprintf("anonymous%04d", 1000+rand.Int63n(8999)), "user")
	ircobj.UseTLS = true
	ircobj.UseSASL = true
	ircobj.SASLLogin = os.Getenv("IRC_NICK")
	ircobj.SASLPassword = os.Getenv("IRC_PSWD")

	events := make(chan interface{}, 2)
	initialized := false
	var nicks []string

	ircobj.AddCallback("903", func(e *irc.Event) {
		events <- AuthedEvent{}
	})

	ircobj.AddCallback("NOTICE", func(e *irc.Event) {
		if len(e.Arguments) == 2 && strings.HasPrefix(e.Arguments[1], "Nicks") {
			parts := strings.SplitN(e.Arguments[1], ": ", 2)
			nicks = strings.Split(parts[1], " ")[1:]
			events <- GotNicksEvent{}
		}
	})

	err := ircobj.Connect(os.Getenv("IRC_SERVER"))
	if err != nil {
		log.Fatal(err)
	}
	go ircobj.Loop()

Main:
	for {
		select {
		case raw := <-events:
			switch raw.(type) {
			case AuthedEvent:
				time.Sleep(time.Second * 5)
				log.Printf("Authed, getting nick list")
				ircobj.Privmsgf("NickServ", "info %s", os.Getenv("IRC_NICK"))

			case GotNicksEvent:
				initialized = true
				log.Printf("Will cycle through: %v", nicks)
				time.Sleep(time.Second * 5)
				break Main
			}
		case <-time.After(time.Second * 40):
			if !initialized {
				log.Fatal("Timeout waiting for event")
			}
		}
	}

	for {
		for _, nick := range nicks {
			log.Printf("/nick %v", nick)
			ircobj.Nick(nick)
			time.Sleep(time.Duration(60+rand.Int63n(30)) * time.Minute)
		}
	}
}
