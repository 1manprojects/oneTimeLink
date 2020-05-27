package main

import "html/template"

/*SecretType ...*/
type SecretType int

/*Enum user ...*/
const (
	Text SecretType = 0
	File SecretType = 1
)

/*Secret ...*/
type Secret struct {
	data    []byte
	ofType  SecretType
	counter int
	name    string
	pass    string
}

/*Config user ...*/
type Config struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Url      string `json:"url"`
	Logo     string `json:"logo"`
	Privacy  string `json:"privacy"`
	Mail     string `json:"mail"`
}

/*ActiveLink active links ...*/
type ActiveLink struct {
	Type  string
	Url   template.URL
	Count int
	Name  string
}

type pageData struct {
	ActiveLinks []ActiveLink
	Logo        string
	Footer      template.HTML
}

type invalidPageData struct {
	ActiveLinks []ActiveLink
	Logo        string
	Footer      template.HTML
	Message     string
}

type secretPageData struct {
	Secret string
	Logo   string
	Footer template.HTML
}

type secretGetPageData struct {
	Secret string
	Logo   string
	Footer template.HTML
	Pass   template.HTML
}

type secretPreviewData struct {
	Secret template.HTML
	Url    string
	Logo   string
	Footer template.HTML
}

type getSecretData struct {
	Secret string
	Logo   string
	Footer template.HTML
	PASS   template.HTML
}

type secretPageHTMLData struct {
	Secret template.HTML
	Logo   string
	Footer template.HTML
}
