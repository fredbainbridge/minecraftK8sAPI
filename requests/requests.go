package requests

type WorldRequest struct {
	Name string `json:"name,omitempty"`
	Port int    `json:"port,omitempty"`
	Tags []Tag  `json:"tags,omitempty"`
}

type Tag struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
