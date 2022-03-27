package model

import "encoding/xml"

// OcrEl MiniOcr base element
type OcrEl struct {
	XMLName xml.Name `xml:"ocr"`
	Pages   []P      `xml:"p"`
}

// P MiniOcr page element
type P struct {
	XMLName    xml.Name `xml:"p"`
	Id         string   `xml:"xml:id,attr"`
	Dimensions string   `xml:"wh,attr"`
	Blocks     []B      `xml:"b"`
}

// B MiniOcr block element
type B struct {
	XMLName xml.Name `xml:"b"`
	Lines   []L      `xml:"l"`
}

// L MiniOcr line element
type L struct {
	XMLName xml.Name `xml:"l"`
	Words   []W      `xml:"w"`
}

// W MiniOcr word element
type W struct {
	XMLName       xml.Name `xml:"w"`
	CoorinateAttr string   `xml:"x,attr"`
	Content       string   `xml:",chardata"`
}
