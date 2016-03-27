package docker

import (
	"path"
	"strings"
)

type Image struct {
	Registry   string `json:"registry"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

func (this Image) ImageString() string {
	s := this.Repository
	if this.Tag != "" {
		s = s + ":" + this.Tag
	}
	return s
}

func (this Image) Url() string {
	return path.Join(this.Registry, this.ImageString())
}

func ParseImageUrl(url string) Image {
	image := Image{}
	delim1 := strings.Index(url, "://")
	if delim1 < 0 {
		delim1 = 0
	} else {
		delim1 += 3
	}
	tag_index := strings.LastIndex(url[delim1:], ":")
	if tag_index > -1 {
		tag_index += delim1
		image.Tag = url[tag_index+1:]
	} else {
		tag_index = len(url)
	}
	project := path.Base(url[0:tag_index])
	account := path.Base(path.Dir(url[0:tag_index]))
	delim2 := strings.Index(url, account)
	image.Registry = url[0 : delim2-1]
	image.Repository = path.Join(account, project)
	return image
}
