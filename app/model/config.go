package model

type Configuration struct {
	DSpaceHost           string
	ManifestBase         string
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
	VerboseLogging       bool
	LogDir               string
}
