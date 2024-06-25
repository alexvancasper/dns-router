package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type BlackList struct {
	data map[string]struct{}
}

func NewBlackList() *BlackList {
	return &BlackList{
		data: make(map[string]struct{}),
	}
}

func (b *BlackList) Add(server string) bool {
	server = strings.Trim(server, " ")
	if len(server) == 0 {
		return false
	}

	if !strings.HasSuffix(server, ".") {
		server += "."
	}
	b.data[server] = struct{}{}

	return true
}

func (b *BlackList) AddList(servers []string) (count int) {
	for _, server := range servers {
		if b.Add(server) {
			count++
		}
	}

	return
}

func (b *BlackList) Contains(server string) bool {
	_, ok := b.data[server]
	return ok
}

func UpdateList() *BlackList {
	list := NewBlackList()

	for _, v := range config.Blocklist {
		ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer ctxCancel()
		resp, err := http.NewRequestWithContext(ctx, http.MethodGet, v, nil)
		if err != nil {
			log.Printf("[black]: could not request blacklist: %s", err.Error())
			continue
		}
		res, err := http.DefaultClient.Do(resp)
		if err != nil {
			log.Printf("[black]: error making http request: %s", err.Error())
			continue
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			log.Println("[black] Status code of", v, "!= 200")
			continue
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println("[black] Can't read body of", v)
			continue
		}
		servers := parseServers(data)
		cnt := list.AddList(servers)
		log.Println("[black] Loaded", cnt, "servers from", v)
	}

	return list
}

func listUpdater() {
	for {
		time.Sleep(config.UpdateInterval)
		blackList = UpdateList()
	}
}

func parseServers(data []byte) []string {
	data2 := strings.ReplaceAll(string(data), "\r", "")
	servers := strings.Split(data2, "\n")
	result := make([]string, 0, len(servers))
	r := regexp.MustCompile(`0\.0\.0\.0\s(.*)`)
	for _, line := range servers {
		if r.MatchString(line) {
			domain := r.ReplaceAllString(line, "$1")
			if domain == "0.0.0.0" {
				continue
			}
			result = append(result, domain)
		}
	}
	return result
}
