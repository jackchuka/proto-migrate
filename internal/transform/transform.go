package transform

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/emicklei/proto"
	"github.com/jackchuka/proto-migrate/internal/config"
	"github.com/jackchuka/proto-migrate/internal/loader"
)

type Rule interface {
	ID() string
	Apply(file *loader.ProtoFile) (changed bool, err error)
}

type Registry struct {
	rules map[string]func(cfg config.Rule) Rule
}

var defaultRegistry = &Registry{
	rules: make(map[string]func(cfg config.Rule) Rule),
}

func init() {
	RegisterRule("package", func(cfg config.Rule) Rule {
		return &PackageRule{From: cfg.From, To: cfg.To}
	})
	RegisterRule("service", func(cfg config.Rule) Rule {
		return &ServiceRule{From: cfg.From, To: cfg.To}
	})
	RegisterRule("import", func(cfg config.Rule) Rule {
		return &ImportRule{From: cfg.From, To: cfg.To}
	})
	RegisterRule("option", func(cfg config.Rule) Rule {
		return &OptionRule{From: cfg.From, To: cfg.To}
	})
	RegisterRule("regexp", func(cfg config.Rule) Rule {
		return &RegexpRule{Pattern: cfg.Pattern, Replace: cfg.Replace}
	})
}

func RegisterRule(kind string, factory func(cfg config.Rule) Rule) {
	defaultRegistry.rules[kind] = factory
}

func CreateRule(cfg config.Rule) (Rule, error) {
	factory, ok := defaultRegistry.rules[cfg.Kind]
	if !ok {
		return nil, fmt.Errorf("unknown rule kind: %s", cfg.Kind)
	}
	return factory(cfg), nil
}

type PackageRule struct {
	From string
	To   string
}

func (r *PackageRule) ID() string {
	return fmt.Sprintf("package.rename:%s->%s", r.From, r.To)
}

func (r *PackageRule) Apply(file *loader.ProtoFile) (bool, error) {
	changed := false
	newContent := file.Content

	proto.Walk(file.Proto,
		proto.WithPackage(func(p *proto.Package) {
			if p.Name == r.From {
				oldLine := fmt.Sprintf("package %s;", p.Name)
				newLine := fmt.Sprintf("package %s;", r.To)
				newContent = strings.Replace(newContent, oldLine, newLine, 1)
				changed = true
			}
		}),
	)

	if changed {
		file.Content = newContent
		return true, nil
	}
	return false, nil
}

type ServiceRule struct {
	From string
	To   string
}

func (r *ServiceRule) ID() string {
	return fmt.Sprintf("service.rename:%s->%s", r.From, r.To)
}

func (r *ServiceRule) Apply(file *loader.ProtoFile) (bool, error) {
	changed := false
	newContent := file.Content

	proto.Walk(file.Proto,
		proto.WithService(func(s *proto.Service) {
			if s.Name == r.From {
				oldPattern := fmt.Sprintf(`service\s+%s\s*{`, regexp.QuoteMeta(s.Name))
				newService := fmt.Sprintf("service %s {", r.To)
				re := regexp.MustCompile(oldPattern)
				newContent = re.ReplaceAllString(newContent, newService)
				changed = true
			}
		}),
	)

	if changed {
		file.Content = newContent
		return true, nil
	}
	return false, nil
}

type ImportRule struct {
	From string
	To   string
}

func (r *ImportRule) ID() string {
	return fmt.Sprintf("import.rewrite:%s->%s", r.From, r.To)
}

func (r *ImportRule) Apply(file *loader.ProtoFile) (bool, error) {
	changed := false
	newContent := file.Content

	proto.Walk(file.Proto,
		proto.WithImport(func(i *proto.Import) {
			if strings.Contains(i.Filename, r.From) {
				oldImport := fmt.Sprintf(`import "%s";`, i.Filename)
				newFilename := strings.ReplaceAll(i.Filename, r.From, r.To)
				newImport := fmt.Sprintf(`import "%s";`, newFilename)
				newContent = strings.Replace(newContent, oldImport, newImport, 1)
				changed = true
			}
		}),
	)

	if changed {
		file.Content = newContent
		return true, nil
	}
	return false, nil
}

type OptionRule struct {
	From string
	To   string
}

func (r *OptionRule) ID() string {
	return fmt.Sprintf("option.update:%s->%s", r.From, r.To)
}

func (r *OptionRule) Apply(file *loader.ProtoFile) (bool, error) {
	changed := false
	newContent := file.Content

	patterns := []struct {
		option string
		regex  string
	}{
		{"go_package", `option\s+go_package\s*=\s*"([^"]+)"`},
		{"java_package", `option\s+java_package\s*=\s*"([^"]+)"`},
		{"swift_prefix", `option\s+swift_prefix\s*=\s*"([^"]+)"`},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		matches := re.FindAllStringSubmatch(newContent, -1)
		for _, match := range matches {
			if len(match) > 1 && strings.Contains(match[1], r.From) {
				oldValue := match[1]
				newValue := strings.ReplaceAll(oldValue, r.From, r.To)
				oldOption := match[0]
				newOption := fmt.Sprintf(`option %s = "%s"`, p.option, newValue)
				newContent = strings.Replace(newContent, oldOption, newOption, 1)
				changed = true
			}
		}
	}

	if changed {
		file.Content = newContent
		return true, nil
	}
	return false, nil
}

type RegexpRule struct {
	Pattern string
	Replace string
}

func (r *RegexpRule) ID() string {
	return fmt.Sprintf("regexp:%s->%s", r.Pattern, r.Replace)
}

func (r *RegexpRule) Apply(file *loader.ProtoFile) (bool, error) {
	re, err := regexp.Compile(r.Pattern)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %w", err)
	}

	newContent := re.ReplaceAllString(file.Content, r.Replace)
	if newContent != file.Content {
		file.Content = newContent
		return true, nil
	}
	return false, nil
}
