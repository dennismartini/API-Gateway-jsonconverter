package main

import (
	"flag"
	"jsontoyamlconverte/controllers"
)

func main() {
	// if os.Args[1] == "v1" {
	// 	controllers.StartV1()
	// } else if os.Args[1] == "v2" {
	// 	controllers.StartV2()
	// } else if os.Args[1] == "v3" {
	accountId := flag.String("account-id", "", "ID da conta AWS (ex: 123456789012)")
	lbarnuri := flag.String("lb-arn-uri", "", "ARN do Load Balancer (ex: arn:aws:elasticloadbalancing:us-east-1:123456789012:listener/app/my-load-balancer/50dc6c495c0c9188/3c5f9d6c12a1a6d1)")
	actionPtr := flag.String("action", "convert", "Ação")
	sourcePathPtr := flag.String("source", "", "Caminho de origem")
	indexnumberPtr := flag.String("index", "", "Indice para geração dos arquivos")
	vpcLinkPtr := flag.String("vpclinkid", "", "ID da VPC Link (ex: abcdef0)")
	awsprofilePtr := flag.String("aws-profile", "", "Perfil AWS a ser carregado (ex: default, dev)")
	stageNamePtr := flag.String("stage-name", "", "Nome do Estágio (ex: dev, prod)")
	apigatewaynamePtr := flag.String("apigateway-name", "", "Nome do API Gateway (ex: api-dev, api-prod, api-loja)")
	targetFilePathPtr := flag.String("targetfile", "", "Arquivo de destino")
	domainNamePtr := flag.String("domain-name", "", "Nome do domínio (ex: api.example.com)")
	deployPtr := flag.String("deploy", "false", "Deploy (ex: true, false)")
	serverURLPtr := flag.String("server-url", "", "URL do servidor (ex: http://localhost:8080)")
	FailOnWarnings := flag.Bool("fail-on-warnings", true, "Falhar em caso de warning (ex: true, false)")
	flag.Parse()
	if action := *actionPtr; action == "convert" {
		controllers.StartV3(*lbarnuri, *accountId, *sourcePathPtr, *indexnumberPtr, *vpcLinkPtr, *awsprofilePtr, *stageNamePtr, *apigatewaynamePtr, *targetFilePathPtr, domainNamePtr, deployPtr, *serverURLPtr, *FailOnWarnings)
	}
	// }
}

// // }
// package main

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"strings"

// 	"github.com/go-openapi/loads"
// 	"github.com/go-openapi/spec"
// 	"gopkg.in/yaml.v3"
// )

// // Uma função para ler o nome de todas as pastas de um caminho e salvar em um slice
// func ReadDirNames(dir string) ([]string, error) {
// 	f, err := os.Open(dir)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer f.Close()
// 	names, err := f.Readdirnames(-1)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return names, nil
// }

// func getJSONFolders(rootFolder string) ([]string, error) {
// 	var folders []string
// 	err := filepath.Walk(rootFolder, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if info.IsDir() {
// 			_, err := os.Stat(filepath.Join(path, "*.json"))
// 			if !os.IsNotExist(err) {
// 				folders = append(folders, path)
// 			}
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return folders, nil
// }

// // yamlPath := "newapifile.yaml"
// // err = os.WriteFile(yamlPath, yamlBytes, 0644)
// // if err != nil {
// // 	log.Println(err)
// // }

// func main() {
// 	swagger := &loads.Document{}
// 	updatedPaths := make(map[string]spec.PathItem)
// 	folders, err := getJSONFolders(os.Args[1])
// 	if err != nil {
// 		panic(err)
// 	}
// 	for _, folder := range folders {
// 		err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
// 			if err != nil {
// 				return err
// 			}
// 			if !info.IsDir() && filepath.Ext(path) == ".json" {
// 				data, err := os.ReadFile(path)
// 				if err != nil {
// 					return err
// 				}
// 				doc, err := loads.Analyzed(data, "")
// 				if err != nil {
// 					return err
// 				}
// 				log.Println(string(data))
// 				for path, pathItem := range doc.Spec().Paths.Paths {
// 					updatedPath := fmt.Sprintf("%s/%s/%s", folder, strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), path)
// 					updatedPaths[updatedPath] = pathItem
// 				}
// 			}
// 			return nil
// 		})
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	swagger.Spec().Paths.Paths = updatedPaths
// 	yamlBytes, err := yaml.Marshal(swagger.Spec())
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = os.WriteFile("documentation.yaml", yamlBytes, 0644)
// 	if err != nil {
// 		panic(err)
// 	}
// }
