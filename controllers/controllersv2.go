package controllers

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"gopkg.in/yaml.v2"
)

func IsApiDeployed(svc *apigatewayv2.ApiGatewayV2, name string) (bool, error) {
	input := &apigatewayv2.GetApisInput{}
	result, err := svc.GetApis(input)
	if err != nil {
		return false, err
	}

	for _, api := range result.Items {
		if *api.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func AddPrivateLinkToContent(data map[string]interface{}, vpcLinkID string, lbarnuri string) error {
	components, ok := data["components"].(map[string]interface{})
	if !ok {
		log.Println("error: 'components' key not found in JSON data")
	}

	components["x-amazon-apigateway-integrations"] = map[string]interface{}{
		"integration1": map[string]interface{}{
			"type":                 "http_proxy",
			"uri":                  lbarnuri,
			"connectionType":       "VPC_LINK",
			"connectionId":         vpcLinkID,
			"payloadFormatVersion": "1.0",
			"httpMethod":           "ANY",
			// "$ref": "#/components/x-amazon-apigateway-integrations/integration1",
		}}

	return nil
}

func AddPrivateLinkToRoutes(data map[string]interface{}, vpcLinkID string) error {
	paths, ok := data["paths"].(map[string]interface{})
	if !ok {
		log.Println("error: 'paths' key not found in JSON data")
	}

	for route := range paths {
		routeData, ok := paths[route].(map[string]interface{})
		if !ok {
			return fmt.Errorf("error: route data not found")
		}

		for method := range routeData {
			methodData, ok := routeData[method].(map[string]interface{})
			if !ok {
				return fmt.Errorf("error: method data not found")
			}

			methodData["x-amazon-apigateway-integration"] = map[string]interface{}{
				// "type":                 "http_proxy",
				// "uri":                  "arn:aws:elasticloadbalancing:us-east-1:123123123213:listener/net/abb39ec342asfasfasf4ed12d98u98usd/109dd4ea79648ae0/4185f7182771cb96",
				// "connectionType":       "VPC_LINK",
				// "connectionId":         vpcLinkID,
				// "httpMethod":           method,
				// "payloadFormatVersion": "1.0",
				"$ref": "#/components/x-amazon-apigateway-integrations/integration1",
			}
		}
	}

	return nil
}

func PathCorrection(data map[string]interface{}) error {
	paths, ok := data["paths"].(map[string]interface{})
	if !ok {
		log.Println("error: 'paths' key not found in JSON data")
	}

	//if path name from paths ends with "/" then remove "/" from path name
	for route := range paths {
		if strings.HasSuffix(route, "/") {
			routeData, ok := paths[route].(map[string]interface{})
			if !ok {
				return fmt.Errorf("error: route data not found")
			}

			delete(paths, route)
			paths[strings.TrimSuffix(route, "/")] = routeData
		}
	}
	return nil
}

func AddStageToRoutes(data map[string]interface{}, stageName string) error {
	if _, ok := data["x-amazon-apigateway-deployment"]; ok {
		return fmt.Errorf("error: 'x-amazon-apigateway-deployment' key already exists in JSON data")
	}

	data["x-amazon-apigateway-deployment"] = map[string]interface{}{
		"description": "Deployment description",
		"stageName":   stageName,
	}

	return nil
}

func AddFilenameToPath(data map[string]interface{}, fileName string, folderName string) error {
	// Obtém a versão do arquivo a partir do nome do arquivo
	fileVersion := strings.TrimSuffix(fileName, ".json")

	// Obtém o objeto "paths" do arquivo OpenAPI
	paths, ok := data["paths"].(map[string]interface{})
	if !ok {
		return errors.New("error: 'paths' key not found in JSON data")
	}

	// Para cada entrada em "paths", atualiza o caminho
	for path, pathItem := range paths {
		// Converte o valor de "pathItem" para um mapa
		pathItemMap, ok := pathItem.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error: invalid path item for path %s", path)
		}

		// Atualiza o caminho com o nome da pasta e a versão do arquivo
		newPath := fmt.Sprintf("/%s/%s/%s", folderName, fileVersion, strings.TrimLeft(path, "/"))
		delete(paths, path)          // Remove a entrada antiga
		paths[newPath] = pathItemMap // Adiciona a entrada atualizada
	}

	return nil
}

// A function to append data[paths] to new openapi file
func AppendPaths(data map[string]interface{}, newFile map[string]interface{}) error {
	// Obtém o objeto "paths" do arquivo OpenAPI
	paths, ok := data["paths"].(map[string]interface{})
	if !ok {
		log.Println("error: 'paths' key not found in JSON data")
	}

	// Obtém o objeto "paths" do arquivo OpenAPI
	newPaths, ok := newFile["paths"].(map[string]interface{})
	if !ok {
		log.Println("error: 'paths' key not found in JSON data")
	}

	// Para cada entrada em "paths", atualiza o caminho
	for path, pathItem := range newPaths {
		// Converte o valor de "pathItem" para um mapa
		pathItemMap, ok := pathItem.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error: invalid path item for path %s", path)
		}

		paths[path] = pathItemMap // Adiciona a entrada atualizada
	}

	return nil
}

// A function to remove components from the OpenAPI file
func RemoveComponents(data map[string]interface{}) error {
	// Obtém o objeto "components" do arquivo OpenAPI
	components, ok := data["components"].(map[string]interface{})
	if !ok {
		log.Println("error: 'components' key not found in JSON data")
	}

	// Remove o objeto "components" do arquivo OpenAPI
	delete(data, "components")

	// Obtém o objeto "paths" do arquivo OpenAPI
	paths, ok := data["paths"].(map[string]interface{})
	if !ok {
		return errors.New("error: 'paths' key not found in JSON data")
	}

	// Para cada entrada em "paths", remove os componentes
	for _, pathItem := range paths {
		// Converte o valor de "pathItem" para um mapa
		pathItemMap, ok := pathItem.(map[string]interface{})
		if !ok {
			return errors.New("error: invalid path item")
		}

		// Para cada método, remove os componentes
		for _, method := range pathItemMap {
			// Converte o valor de "method" para um mapa
			methodMap, ok := method.(map[string]interface{})
			if !ok {
				return errors.New("error: invalid method")
			}

			// Obtém o objeto "requestBody" do método
			requestBody, ok := methodMap["requestBody"].(map[string]interface{})
			if !ok {
				continue
			}

			// Obtém o objeto "content" do objeto "requestBody"
			content, ok := requestBody["content"].(map[string]interface{})
			if !ok {
				continue
			}

			// Obtém o objeto "application/json" do objeto "content"
			json, ok := content["application/json"].(map[string]interface{})
			if !ok {
				continue
			}

			// Obtém o objeto "schema" do objeto "application/json"
			schema, ok := json["schema"].(map[string]interface{})
			if !ok {
				continue

			}

			// Obtém o objeto "$ref" do objeto "schema"
			ref, ok := schema["$ref"].(string)
			if !ok {
				continue
			}

			// Obtém o nome do componente a partir do objeto "$ref"
			componentName := strings.TrimPrefix(ref, "#/components/schemas/")
			if componentName == ref {
				continue
			}

			// Remove o componente do objeto "components"
			delete(components, componentName)
		}
	}

	return nil
}

func AddVersionToTitle(data map[string]interface{}, fileName string) error {
	info, ok := data["info"].(map[string]interface{})
	if !ok {
		log.Println("error: 'info' key not found in JSON data")
	}

	title, ok := info["title"].(string)
	if !ok {
		log.Println("error: 'title' key not found in 'info' data")
		return nil
	}

	fileVersion := strings.TrimSuffix(fileName, ".json")
	info["title"] = title + " " + fileVersion

	return nil
}

func SwitchCredentials(profile string) *session.Session {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},
		Profile: profile,
	})
	if err != nil {
		log.Println("error creating session:", err)
		return nil
	}
	return sess
}

func CreateApiMapping(svc *apigatewayv2.ApiGatewayV2, domainName string, stage string, apiId string) error {
	input := &apigatewayv2.CreateApiMappingInput{
		DomainName: aws.String(domainName),
		ApiId:      aws.String(apiId),
		Stage:      aws.String(stage),
	}
	_, err := svc.CreateApiMapping(input)
	if err != nil {
		return err
	}

	log.Println("API Mapping created successfully")

	return nil
}

func DeleteAllAPIsv1(svc *apigateway.APIGateway) error {
	apiInput := &apigateway.GetRestApisInput{Limit: aws.Int64(500)}
	result, err := svc.GetRestApis(apiInput)
	if err != nil {
		return err
	}

	for _, api := range result.Items {
		_, err := svc.DeleteRestApi(&apigateway.DeleteRestApiInput{
			RestApiId: api.Id,
		})
		if err != nil {
			return err
		}

		log.Println("Deleted API", *api.Name)
		log.Println("Waiting 30 seconds")
		time.Sleep(31 * time.Second)
	}
	return nil
}

func DeleteAllAPIs(svc *apigatewayv2.ApiGatewayV2) error {
	apiInput := &apigatewayv2.GetApisInput{MaxResults: aws.String("500")}
	result, err := svc.GetApis(apiInput)
	if err != nil {
		return err
	}

	for _, api := range result.Items {
		_, err := svc.DeleteApi(&apigatewayv2.DeleteApiInput{
			ApiId: api.ApiId,
		})
		if err != nil {
			return err
		}

		log.Println("Deleted API", *api.Name)
		log.Println("Waiting 30 seconds")
		time.Sleep(31 * time.Second)
	}

	return nil
}

func ImportOrUpdateAPI(svc *apigatewayv2.ApiGatewayV2, file string, data map[string]interface{}, stageName string, dnsbase string, filename fs.FileInfo) error {

	info, ok := data["info"].(map[string]interface{})
	if !ok {
		log.Println("error: 'info' key not found in JSON data on importing or updated API")
	}

	title, ok := info["title"].(string)
	if !ok {
		log.Println("error: 'title' key not found in 'info' data on importing or updated API")
		return nil
	}

	yamlData, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	apiInput := &apigatewayv2.GetApisInput{MaxResults: aws.String("500")}
	result, err := svc.GetApis(apiInput)
	if err != nil {
		return err
	}

	var found bool

	for _, api := range result.Items {
		if *api.Name == title {
			log.Println("Updating " + *api.Name)
			found = true
			yamlDataStr := string(yamlData)
			yamlDataStrPtr := &yamlDataStr
			_, err = svc.ReimportApi(&apigatewayv2.ReimportApiInput{
				ApiId: api.ApiId,
				Body:  yamlDataStrPtr,
			})
			if err != nil {
				log.Println(err)
			}
			// create a deployment
			log.Println("Deployng", *api.Name)
			//Create stage if doesn't exist
			stageInput := &apigatewayv2.CreateStageInput{
				ApiId: api.ApiId,
				//DeploymentId: deploymentResult.DeploymentId,
				StageName: aws.String(stageName),
			}
			_, err = svc.CreateStage(stageInput)
			if err != nil {
				log.Println(err)
			}

			s := strings.TrimSuffix(filename.Name(), ".json")
			log.Print(dnsbase + "/" + s)
			createApiMappingInput := &apigatewayv2.CreateApiMappingInput{
				DomainName:    aws.String("api.develop.asdasdasdasd.com.br"),
				ApiId:         api.ApiId,
				Stage:         aws.String(stageName),
				ApiMappingKey: aws.String(dnsbase + "/" + s),
			}

			_, err = svc.CreateApiMapping(createApiMappingInput)
			if err != nil {
				log.Println(err)
			}

			deploymentInput := &apigatewayv2.CreateDeploymentInput{
				ApiId:       api.ApiId,
				Description: aws.String("Deployment description"),
				StageName:   aws.String(stageName),
			}

			_, err := svc.CreateDeployment(deploymentInput)
			if err != nil {
				return err
			}

			// Create api mapping

			log.Println("Updated API", *api.Name)
			break
		}
	}
	if !found {
		yamlDataStr := string(yamlData)
		yamlDataStrPtr := &yamlDataStr
		log.Println("Importing " + file)
		importedRestAPI, err := svc.ImportApi(&apigatewayv2.ImportApiInput{
			Body: yamlDataStrPtr,
		})
		if err != nil {
			return err
		}

		stageInput := &apigatewayv2.CreateStageInput{
			ApiId: importedRestAPI.ApiId,
			//DeploymentId: deploymentResult.DeploymentId,
			StageName: aws.String(stageName),
		}
		_, err = svc.CreateStage(stageInput)
		if err != nil {
			return err
		}

		createApiMappingInput := &apigatewayv2.CreateApiMappingInput{
			DomainName:    aws.String("api.develop.asdasdasd.com.br"),
			ApiId:         importedRestAPI.ApiId,
			Stage:         aws.String(stageName),
			ApiMappingKey: aws.String(dnsbase + "/" + filename.Name()),
		}

		_, err = svc.CreateApiMapping(createApiMappingInput)
		if err != nil {
			log.Println(err)
		}

		deploymentInput := &apigatewayv2.CreateDeploymentInput{
			ApiId:       importedRestAPI.ApiId,
			Description: aws.String("Deployment description"),
			StageName:   aws.String(stageName),
		}

		_, err = svc.CreateDeployment(deploymentInput)
		if err != nil {
			return err
		}

		log.Println("Updated API", *importedRestAPI.Name)
		if err != nil {
			log.Println(err)
		}

		log.Println("Imported", file, "to API")
	}

	// log.Println("API Mapping created successfully")
	return nil
}

func DeleteSecurity(data map[string]interface{}) map[string]interface{} {
	delete(data["components"].(map[string]interface{}), "securitySchemes")
	delete(data, "security")
	return data
}

func ExtractPathsToSingleFile() {
	// Pasta raiz que contém as subpastas com os arquivos YAML do OpenAPI 3
	rootDir := os.Args[2]

	// Cria um mapa para armazenar os caminhos únicos
	paths := make(map[string]bool)

	// Percorre todas as pastas e subpastas
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignora pastas que não contêm arquivos YAML
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		// Lê o arquivo YAML
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Extrai os caminhos do arquivo YAML
		var doc map[interface{}]interface{}
		err = yaml.Unmarshal(data, &doc)
		if err != nil {
			return err
		}

		pathsMap, ok := doc["paths"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid yaml: %s", path)
		}

		for key := range pathsMap {
			paths[key] = true
		}

		// for _, path := range pathsList {
		// 	pathMap, ok := path.(map[interface{}]interface{})
		// 	if !ok {
		// 		continue
		// 	}

		// 	for key := range pathMap {
		// 		paths[key.(string)] = true
		// 	}
		// }

		return nil
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	// Escreve todos os caminhos no arquivo "paths.txt"
	file, err := os.Create("paths.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	for path := range paths {
		fmt.Fprintln(file, path)
	}

	fmt.Println("Caminhos extraídos com sucesso!")
}
