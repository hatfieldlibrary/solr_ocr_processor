package model

type Configuration struct {
	DSpaceHost           string
	Collections          []string
	SolrUrl              string
	SolrCore             string
	ConvertToMiniOcr     bool
	IndexType            string
	EscapeUtf8           bool
	XmlFileLocation      string
	HttpPort             string
	IpWhitelist          []string
	InputImageResolution int
	LogDir               string
}
