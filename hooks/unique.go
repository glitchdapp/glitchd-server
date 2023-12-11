package hooks

import (
	"github.com/99designs/gqlgen/plugin/modelgen"
)

func UniqueConstraint(b *modelgen.ModelBuild) *modelgen.ModelBuild {
	for _, model := range b.Models {
		for _, field := range model.Fields {
			if field.Name == "id" {
				field.Tag += ` bun:"unique"`
			}
		}
	}

	return b
}
