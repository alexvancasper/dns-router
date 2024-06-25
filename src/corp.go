package main

import (
	"strings"
)

type CorpDomain struct {
	suffix  string
	exclude string
}

func NewCorpDomain() *CorpDomain {
	return &CorpDomain{
		suffix: "",
	}
}

func (b *CorpDomain) Set(domain string) bool {
	domain = strings.Trim(domain, " ")
	if len(domain) == 0 {
		return false
	}

	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}
	b.suffix = domain

	return true
}

func (b *CorpDomain) SetExclude(domain string) bool {
	domain = strings.Trim(domain, " ")
	if len(domain) == 0 {
		return false
	}

	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}
	b.exclude = domain

	return true
}

func (b *CorpDomain) Match(domain string) bool {
	return strings.HasSuffix(domain, b.suffix)
}

func (b *CorpDomain) MatchExclude(domain string) bool {
	return strings.HasSuffix(domain, b.exclude)
}

func (b *CorpDomain) GetExclude() string {
	return b.exclude
}
