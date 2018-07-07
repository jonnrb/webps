package fe

import (
	"html/template"
	"io"
	"sort"
	"strings"

	"go.jonnrb.io/webps/fe/assets"
	"go.jonnrb.io/webps/pb"
)

var tmpl = template.Must(
	template.New("").Parse(string(assets.MustAsset("tmpl/index.html"))))

type Page struct {
	Sections []*Section
}

func (p Page) Render(w io.Writer) error {
	return tmpl.Execute(w, p)
}

type Section struct {
	Name       string
	Containers []*Container
}

type Container struct {
	Name, Image, Status string
}

func PageFromList(list *webpspb.ListResponse) Page {
	projs := make(map[string]*Section)
	for _, c := range list.Container {
		proj, ok := c.DockerComposeLabels["com.docker.compose.project"]
		if !ok {
			proj = "(none)"
		}

		var s *Section
		s, ok = projs[proj]
		if !ok {
			s = &Section{Name: proj}
			projs[proj] = s
		}

		s.Containers = append(s.Containers, &Container{
			Name:   strings.TrimPrefix(c.Name, "/"),
			Image:  c.Image,
			Status: c.Status,
		})
	}

	n := projs["(none)"]
	if n != nil {
		delete(projs, "(none)")
	}

	var keys []string
	for k := range projs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var p Page
	for _, k := range keys {
		p.Sections = append(p.Sections, projs[k])
	}
	if n != nil {
		p.Sections = append(p.Sections, n)
	}
	return p
}
