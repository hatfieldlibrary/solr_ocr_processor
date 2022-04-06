package model

import (
	"encoding/xml"
)

type OcrxWord struct {
	XMLName xml.Name `xml:"span"`
	Class   string   `xml:"class,attr"`
	Id      string   `xml:"id,attr"`
	Title   string   `xml:"title,attr"`
	Lang    string   `xml:"lang,attr,omitempty"`
	Dir     string   `xml:"dir,attr,omitempty"`
	XmlLang string   `xml:"xml:lang,attr,omitempty"`
	Content string   `xml:",chardata"`
}
