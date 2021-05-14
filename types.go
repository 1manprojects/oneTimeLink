package main

import (
	"html/template"
	"time"
)

/*SecretType ...*/
type SecretType int

/*Enum user ...*/
const (
	Text SecretType = 0
	File SecretType = 1
)

/*Secret ...*/
type Secret struct {
	//Secret Data for File
	data []byte
	//Type Text or File
	ofType SecretType
	//number of times valid
	counter int
	//Name of Secret
	name string
	//Password for Secret
	pass string
	//Two factor verification for User
	twoFa string
	//If Link is active or not
	isActive bool
	//If the Link has been visited but not jet retrived
	visited bool
	//Time of creation
	createdOn time.Time
	//How many minutes link should be valid
	validFor int
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
	Type     string
	Url      template.URL
	Count    int
	Name     string
	TwoFa    string
	State    string
	TimeLeft string
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

type secretGetPageData struct {
	Secret string
	Tfa    string
	Logo   string
	Footer template.HTML
	Pass   template.HTML
}

type secretAuthData struct {
	Secret string
	Logo   string
	Footer template.HTML
}

type secretPreviewData struct {
	Secret template.HTML
	Url    string
	Logo   string
	Footer template.HTML
}
type secretPageHTMLData struct {
	Secret template.HTML
	Logo   string
	Footer template.HTML
}

type boolResponse struct {
	Result bool `json:"result"`
}

type idRequest struct {
	Id  string `json:"Id"`
	Tfa string `json:"Tfa"`
}
