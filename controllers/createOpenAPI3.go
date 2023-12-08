package controllers

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Route struct {
	Path   string     `yaml:"-"`
	Get    *Operation `yaml:"get,omitempty"`
	Post   *Operation `yaml:"post,omitempty"`
	Put    *Operation `yaml:"put,omitempty"`
	Delete *Operation `yaml:"delete,omitempty"`
	Patch  *Operation `yaml:"patch,omitempty"`
}

type Operation struct {
	OperationId string               `yaml:"operationId,omitempty"`
	Description string               `yaml:"-"`
	Parameters  []*Parameter         `yaml:"parameters,omitempty"`
	RequestBody *RequestBody         `yaml:"requestBody,omitempty"`
	Responses   map[string]*Response `yaml:"responses,omitempty"`
}

type Parameter struct {
	In       string  `yaml:"in,omitempty"`
	Name     string  `yaml:"name,omitempty"`
	Required bool    `yaml:"required,omitempty"`
	Schema   *Schema `yaml:"schema,omitempty"`
}

type RequestBody struct {
	Content map[string]*MediaType `yaml:"content,omitempty"`
}

type Response struct {
	Description string                `yaml:"description,omitempty"`
	Content     map[string]*MediaType `yaml:"content,omitempty"`
}

type MediaType struct {
	Schema *Schema `yaml:"schema,omitempty"`
}

type Schema struct {
	Type  string `json:"type,omitempty"`
	Items *Items `json:"items,omitempty"`
	Ref   string `json:"$ref,omitempty"`
}

type Items struct {
	Type string `json:"type,omitempty"`
	Ref  string `json:"$ref,omitempty"`
}

type Components struct {
	Schemas map[string]interface{} `yaml:"schemas,omitempty"`
}

func ConvertToOpenAPI3(filepath string, filepathcomponents string) {
	// Lê o arquivo YAML
	content, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Erro ao ler o arquivo YAML: %v\n", err)
		return
	}

	contentcomponents, err := os.ReadFile(filepathcomponents)
	if err != nil {
		fmt.Printf("Erro ao ler o arquivo YAML: %v\n", err)
		return
	}

	// Faz o unmarshal do conteúdo YAML em um objeto Go
	var routes map[string]Route
	var components Components

	err = yaml.Unmarshal(content, &routes)
	if err != nil {
		fmt.Printf("Erro ao fazer o unmarshal do YAML: %v\n", err)
		return
	}

	err = yaml.Unmarshal(contentcomponents, &components)
	if err != nil {
		fmt.Printf("Erro ao fazer o unmarshal do YAML: %v\n", err)
		return
	}

	// Cria o objeto OpenAPI 3.0 com as rotas
	openapi := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   os.Args[7] + " " + os.Args[6],
			"version": "1.0.0",
		},
		"paths": map[string]interface{}{},
		"components": map[string]interface{}{
			"schemas": components.Schemas,
		},
	}

	for path, route := range routes {
		pathItem := map[string]interface{}{}

		if route.Get != nil {
			operation := createOperation(route.Get)
			pathItem["get"] = operation
		}

		if route.Post != nil {
			operation := createOperation(route.Post)
			pathItem["post"] = operation
		}

		if route.Put != nil {
			operation := createOperation(route.Put)
			pathItem["put"] = operation
		}

		if route.Delete != nil {
			operation := createOperation(route.Delete)
			pathItem["delete"] = operation
		}

		if route.Patch != nil {
			operation := createOperation(route.Patch)
			pathItem["patch"] = operation
		}

		openapi["paths"].(map[string]interface{})[path] = pathItem
	}

	// Escreve o resultado em um novo arquivo YAML
	result, err := yaml.Marshal(&openapi)
	if err != nil {
		fmt.Printf("Erro ao fazer o marshal do YAML: %v\n", err)
		return
	}

	err = os.WriteFile("openapi"+os.Args[6]+".yaml", result, 0644)
	if err != nil {
		fmt.Printf("Erro ao escrever o arquivo YAML: %v\n", err)
		return
	}

	fmt.Println("Arquivo YAML gerado com sucesso")
}

func createOperation(op *Operation) map[string]interface{} {
	operation := map[string]interface{}{
		"operationId": op.OperationId,
	}

	if op.Parameters != nil && len(op.Parameters) > 0 {
		parameters := []map[string]interface{}{}
		for _, param := range op.Parameters {
			parameter := map[string]interface{}{
				"in":       param.In,
				"name":     param.Name,
				"required": param.Required,
			}

			if param.Schema != nil {
				parameter["schema"] = map[string]interface{}{
					"$ref": param.Schema.Ref,
				}
			}

			parameters = append(parameters, parameter)
		}
		operation["parameters"] = parameters
	}

	if op.RequestBody != nil {
		content := map[string]interface{}{}
		for k, v := range op.RequestBody.Content {
			mediaType := map[string]interface{}{}
			if v.Schema != nil {
				mediaType["schema"] = map[string]interface{}{
					"$ref": v.Schema.Ref,
				}
			}
			content[k] = mediaType
		}
		operation["requestBody"] = map[string]interface{}{
			"content": content,
		}
	}

	if op.Responses != nil && len(op.Responses) > 0 {
		responses := map[string]interface{}{}
		for k, v := range op.Responses {
			content := map[string]interface{}{}
			for k2, v2 := range v.Content {
				mediaType := map[string]interface{}{}
				if v2.Schema != nil {
					mediaType["schema"] = map[string]interface{}{
						"$ref": v2.Schema.Ref,
					}
				}
				content[k2] = mediaType
			}

			response := map[string]interface{}{
				"content":     content,
				"description": v.Description,
			}

			responses[k] = response
		}
		operation["responses"] = responses
	}

	return operation
}
