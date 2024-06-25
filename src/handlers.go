package main

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
)

func HandlerTCP(w dns.ResponseWriter, req *dns.Msg) {
	totalRequestsTCP.Inc()
	Handler(w, req)
}

func HandlerUDP(w dns.ResponseWriter, req *dns.Msg) {
	totalRequestsUDP.Inc()
	Handler(w, req)
}

func Handler(w dns.ResponseWriter, req *dns.Msg) {
	defer w.Close()

	question := req.Question[0]

	cachedReq := cache.Get(question.Qtype, question.Name)
	if cachedReq != nil {
		totalCacheHits.Inc()

		response := &dns.Msg{}
		response.SetReply(req)
		response.Answer = append(response.Answer, cachedReq)

		err := w.WriteMsg(response)
		if err != nil {
			log.Printf("error: %v", err.Error())
			return
		}
		totalRequestsSuccess.Inc()
		return
	}

	if (question.Qtype == dns.TypeA || question.Qtype == dns.TypeAAAA) && blackList.Contains(question.Name) {
		response := &dns.Msg{}
		response.SetReply(req)

		head := dns.RR_Header{
			Name:   question.Name,
			Rrtype: question.Qtype,
			Class:  dns.ClassINET,
			Ttl:    uint32(config.UpdateInterval.Seconds()),
		}

		var line dns.RR
		if question.Qtype == dns.TypeA {
			line = &dns.A{
				Hdr: head,
				A:   net.ParseIP(config.BlockAddress4),
			}
		} else {
			line = &dns.AAAA{
				Hdr:  head,
				AAAA: net.ParseIP(config.BlockAddress6),
			}
		}
		response.Answer = append(response.Answer, line)

		err := w.WriteMsg(response)
		if err != nil {
			log.Printf("error: %v", err.Error())
			return
		}
		log.Println("blocked", question.Name)
		totalRequestsBlocked.Inc()
		return
	}

	resp, err := LookupRouter(req)
	if err != nil {
		resp = &dns.Msg{}
		resp.SetRcode(req, dns.RcodeServerFailure)
		log.Println("fail", question.Name)
		totalRequestsFailed.Inc()
	} else {
		totalRequestsSuccess.Inc()
		if len(resp.Answer) > 0 {
			cache.Set(question.Qtype, question.Name, resp.Answer[0])
		}
	}

	err = w.WriteMsg(resp)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return
	}
}

func Lookup(req *dns.Msg, nameservers []string) (*dns.Msg, error) {
	c := &dns.Client{
		Net:          "udp",
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}

	qName := req.Question[0].Name

	res := make(chan *dns.Msg, 1)
	var wg sync.WaitGroup
	L := func(nameserver string) {
		defer wg.Done()
		r, _, err := c.Exchange(req, nameserver)
		if err != nil {
			log.Printf("%s socket error on %s", qName, nameserver)
			log.Printf("error:%s", err.Error())
			return
		}
		if r != nil && r.Rcode != dns.RcodeSuccess {
			if r.Rcode == dns.RcodeServerFailure {
				return
			}
		}
		select {
		case res <- r:
		default:
		}
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Start lookup on each nameserver top-down, in every second
	for _, nameserver := range nameservers {
		wg.Add(1)
		go L(nameserver)
		// but exit early, if we have an answer
		select {
		case r := <-res:
			return r, nil
		case <-ticker.C:
			continue
		}
	}

	// wait for all the namservers to finish
	wg.Wait()
	select {
	case r := <-res:
		return r, nil
	default:
		return nil, errors.New("can't resolve ip for" + qName)
	}
}

func LookupRouter(req *dns.Msg) (*dns.Msg, error) {
	qName := req.Question[0].Name
	debugRequest(qName)
	if corpdomain.MatchExclude(qName) || !corpdomain.Match(qName) {
		totalRequestsToPublicDNS.Inc()
		return Lookup(req, config.Nameservers)
	}
	totalRequestsToCorpDNS.Inc()
	return Lookup(req, config.CorpNameservers)
}

func debugRequest(name string) {
	if corpdomain.MatchExclude(name) {
		log.Printf("qName: %s routed to public DNS due to corporate exclude domain", name)
		return
	}
	if corpdomain.Match(name) {
		log.Printf("qName: %s routed to corp DNS", name)
		return
	}
	log.Printf("qName: %s routed to public DNS", name)
}
