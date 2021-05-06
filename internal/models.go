package internal

type Configuration struct {
	DSpaceHost string
	Collections []string
	SolrUrl string
	SolrCore string
	XmlFileLocation string
	LogDir string
}

type SolrPost struct {
	Id          string `json:"id"`
	ManifestUrl string `json:"manifest_url"`
	OcrText     string `json:"ocr_text"`
}

type Manifest struct {
	Context string `json:"@context"`
	Type string `json:"@type"`
	Id string `json:"@id"`
	Label string `json:"label"`
	Metadata []Metadata `json:"metadata"`
	Service Service `json:"service,omitempty"`
	SeeAlso SeeAlso `json:"SeeAlso,omitempty"`
	Sequences []Sequence `json:"sequences"`
	Thumbnail Thumbnail `json:"thumbnail,omitempty"`
	ViewingHint string `json:"viewingHint,omitempty"`
	Related Related `json:"related,omitempty"`
}
type Metadata struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Service struct {
	Context string `json:"@context"`
	Id string `json:"@id"`
	Profile string `json:"profile"`
	Protocol string `json:"protocol"`
}

type SeeAlso struct {
	Id string `json:"@id"`
	Type string `json:"@type"`
	Label string `json:"label"`
}

type Sequence struct {
	Id string `json:"@id"`
	Type string `json:"@type"`
	Canvases []Canvas `json:"canvases"`
}

type Canvas struct {
	Id string `json:"@id"`
	Type string `json:"@type"`
	Label string `json:"label"`
	Thumbnail Thumbnail `json:"thumbnail"`
	Images []Image `json:"images"`
	Width int `json:"width,int"`
	Height int `json:"height,int"`
}

type Thumbnail struct {
	Id string `json:"@id"`
	Type string `json:"@type"`
	Label string `json:"label"`
	Service Service `json:"service"`
	Format string `json:"format"`
}

type Image struct {
	Type string `json:"@type"`
	Motivation string `json:"motivation"`
	Resource Resource `json:"resource"`
	On string `json:"on"`
}

type Resource struct {
	Id string `json:"@id"`
	Type string `json:"@type"`
	Service Service `json:"service"`
	Format string `json:"format"`
}

type Related struct {
	Id string `json:"@id"`
	Label string `json:"label"`
	Format string `json:"format"`
}

type ResourceAnnotationList struct {
	Context string `json:"@context"`
	Id string `json:"@id"`
	Type string `json:"@type"`
	Resources []ResourceAnnotation `json:"resources"`
}
type ResourceAnnotation struct {
	Id string `json:"@id"`
	Type string `json:"@type"`
	Motivation string `json:"motivation"`
	Resource ResourceAnnotationResource `json:"resource"`
}
type ResourceAnnotationResource struct {
	Id string `json:"@id"`
	Label string `json:"label"`
	Format string `json:"format"`
}

