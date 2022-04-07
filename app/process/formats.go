package process

import (
	"strings"
)

var MiniOcrMatchers = [1]string{"<ocr>"}
var AltoMatchers = [5]string{"<alto", ":alto", "<Description>", "<Layout>", "<Page"}
var HocrMatchers = [11]string{"ocr_document", "ocr_page", "ocr_carea", "ocrx_block", "ocr_chapter", "ocr_section",
	"ocr_subsection", "ocr_par", "ocr_line", "ocrx_line", "ocrx_word"}

type Format int64

const (
	MiniocrFormat Format = iota
	AltoFormat
	HocrFormat
	UnknownFormat
)

func (f Format) String() string {
	switch f {
	case MiniocrFormat:
		return "minocr"
	case AltoFormat:
		return "alto"
	case HocrFormat:
		return "hocr"
	}
	return "unknown"
}

func GetOcrFormat(chunk string) Format {
	for i := 0; i < len(AltoMatchers); i++ {
		if strings.Contains(chunk, AltoMatchers[i]) {
			return AltoFormat
		}
	}
	for i := 0; i < len(HocrMatchers); i++ {
		if strings.Contains(chunk, HocrMatchers[i]) {
			return HocrFormat
		}
	}
	for i := 0; i < len(MiniOcrMatchers); i++ {
		if strings.Contains(chunk, MiniOcrMatchers[i]) {
			return MiniocrFormat
		}
	}

	return UnknownFormat
}
