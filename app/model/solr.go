package model

import "encoding/xml"

type SolrField struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Field   string   `xml:",cdata"`
}

type SolrPostXml struct {
	XMLName xml.Name `xml:"doc"`
	Fields  []SolrField
}

type SolrResponse struct {
	ResponseHeader struct {
		Status int `json:"status"`
		QTime  int `json:"QTime"`
	} `json:"responseHeader"`
	Response struct {
		NumFound      int    `json:"numFound"`
		Start         int    `json:"start"`
		NumFoundExact bool   `json:"numFoundExact"`
		Docs          []Docs `json:"docs"`
	} `json:"response"`
}

type Docs struct {
	ManifestUrl []string `json:"manifest_url,omitempty"`
	OcrText     string   `json:"ocr_text,omitempty"`
}

type SolrCreatePost struct {
	Id          string `json:"id"`
	ManifestUrl string `json:"manifest_url"`
	OcrText     string `json:"ocr_text"`
}

type SolrDeletePost struct {
	Delete Delete `json:"delete"`
}

type Delete struct {
	Query string `json:"query"`
}
