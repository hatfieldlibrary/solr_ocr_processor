package model

import "encoding/xml"

type Alto struct {
	XMLName     xml.Name    `xml:"alto"`
	Xmlns       string      `xml:"xmlns,attr,omitempty"`
	Description Description `xml:"Description"`
	Styles      Styles      `xml:"Styles"`
	Layout      Layout      `xml:"Layout"`
}
type Description struct {
	Xmlns                  string                 `xml:"xmlns,attr,omitempty"`
	MeasurementUnit        string                 `xml:"MeasurementUnit"`
	SourceImageInformation SourceImageInformation `xml:"sourceImageInformation"`
}

type SourceImageInformation struct {
	FileName string `xml:"fileName"'`
}

type Styles struct {
	Xmlns          string           `xml:"xmlns,attr,omitempty"`
	ParagraphStyle []ParagraphStyle `xml":ParagraphStyle"`
}

type ParagraphStyle struct {
	Id        string `xml:"ID,attr"`
	ALIGN     string `xml:"ALIGN,attr"`
	LEFT      string `xml:"LEFT,attr"`
	RIGHT     string `xml:"RIGHT,attr"`
	FIRSTLINE string `xml:"FIRSTLINE,attr"`
	LINESPACE string `xml:"LINESPACE,attr"`
}

type Layout struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	Page  Page   `xml:"Page"`
}

// Page Alto page element
type Page struct {
	Xmlns         string     `xml:"xmlns,attr,omitempty"`
	Id            string     `xml:"ID,attr"`
	PhysicalImgNr string     `xml:"PHYSICAL_IMG_NR,attr"`
	Height        string     `xml:"HEIGHT,attr"`
	Width         string     `xml:"WIDTH,attr"`
	PrintSpace    PrintSpace `xml:"PrintSpace"`
}

// PrintSpace Alto composed block element (not always present, contains TextBlock elements)
type PrintSpace struct {
	Xmlns         string          `xml:"xmlns,attr,omitempty"`
	TextBlock     []TextBlock     `xml:"TextBlock"`
	ComposedBlock []ComposedBlock `xml:"ComposedBlock"`
}

// ComposedBlock Alto composed block element (not always present, contains TextBlock elements)
type ComposedBlock struct {
	Xmlns     string      `xml:"xmlns,attr,omitempty"`
	TextBlock []TextBlock `xml:"TextBlock"`
}

// TextBlock Alto text block element
type TextBlock struct {
	Xmlns    string     `xml:"xmlns,attr,omitempty"`
	TextLine []TextLine `xml:"TextLine"`
}

// TextLine Alto line element
type TextLine struct {
	Xmlns  string   `xml:"xmlns,attr,omitempty"`
	String []String `xml:"String"`
}

// String Alto element containing word information
type String struct {
	Xmlns   string `xml:"xmlns,attr,omitempty"`
	CONTENT string `xml:"CONTENT,attr"`
	HEIGHT  string `xml:"HEIGHT,attr"`
	WIDTH   string `xml:"WIDTH,attr"`
	VPOS    string `xml:"VPOS,attr"`
	HPOS    string `xml:"HPOS,attr"`
}
