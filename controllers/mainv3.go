package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
)

func ChangeTitle(data map[string]interface{}, apigatewayname string) error {
	info, ok := data["info"].(map[string]interface{})
	if !ok {
		log.Println("error: 'info' key not found in JSON data")
	}
	info["title"] = apigatewayname
	return nil
}

func StartV3(lbarnuri string, accountId string, sourcePathPtr string, indexnumberPtr string, vpcLinkPtr string, awsprofilePtr string, stageNamePtr string, apigatewaynamePtr string, targetFilePathPtr string, domainNamePtr *string, deploy *string, serverURL string, FailOnWarnings bool) {
	inputDir := sourcePathPtr
	outputFile := targetFilePathPtr

	// Cria a lista de inputs
	inputs := []map[string]interface{}{}
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Ignora diretórios
			return nil
		}

		if info.Name() == "config.json" {
			// Ignora o arquivo de saída
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".json") {
			// Ignora arquivos que não são JSON
			return nil
		}

		// Obter o nome do diretório pai
		parentDir := filepath.Base(filepath.Dir(path))
		routeversion := info.Name()[:len(info.Name())-5]
		input := map[string]interface{}{
			"inputFile": path,
			"pathModification": map[string]string{
				"prepend": fmt.Sprintf("/%s/%s", parentDir, routeversion),
			},
		}
		inputs = append(inputs, input)
		//var inputroot = map[string]interface{}{}
		if routeversion == "v1" {
			inputroot := map[string]interface{}{
				"inputFile": path,
				"pathModification": map[string]string{
					"prepend": fmt.Sprintf("/%s", parentDir),
				},
			}
			inputs = append(inputs, inputroot)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	// Cria o arquivo de configuração de entrada
	output := map[string]interface{}{
		"inputs": inputs,
		"output": outputFile,
	}
	config, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		panic(err)
	}

	// Grava o arquivo de configuração de entrada
	err = os.WriteFile("config"+indexnumberPtr+".json", config, os.ModePerm)
	if err != nil {
		panic(err)
	}
	fmt.Println("Arquivo de configuração criado com sucesso!")
	//cmd := exec.Command("npm", "install", "openapi-merge-cli")
	cmd := exec.Command("npx", "openapi-merge-cli", "--config", "config"+indexnumberPtr+".json")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Erro ao executar o comando:", err)
		return
	}
	fmt.Println(string(out))

	yamlData := jsonToAWSChanges(lbarnuri, accountId, vpcLinkPtr, awsprofilePtr, stageNamePtr, apigatewaynamePtr, indexnumberPtr, targetFilePathPtr, serverURL)
	// Deploy to AWS
	if *deploy == "true" {
		//Start AWS API Gateway Session
		sess := SwitchCredentials(awsprofilePtr)
		svc := apigatewayv2.New(sess)

		//The api-gateway exists on AWS?
		restAPIs, err := svc.GetApis(&apigatewayv2.GetApisInput{})
		if err != nil {
			panic(err)
		}
		var importedRestAPI *apigatewayv2.ImportApiOutput
		var updatedRestAPI *apigatewayv2.ReimportApiOutput
		apiid, exist := restApiExist(apigatewaynamePtr, restAPIs)
		if exist {
			//Reimport API
			log.Println("API Reimporting: ", apigatewaynamePtr)
			updatedRestAPI, err = svc.ReimportApi(&apigatewayv2.ReimportApiInput{
				ApiId:          aws.String(apiid),
				Body:           aws.String(string(yamlData)),
				FailOnWarnings: aws.Bool(FailOnWarnings),
			})
			if err != nil {
				panic(err)
			}
			log.Println("API Reimported: ", *updatedRestAPI.ApiId)
		} else if !exist {
			//Import API
			log.Println("API Importing: ", apigatewaynamePtr)
			importedRestAPI, err = svc.ImportApi(&apigatewayv2.ImportApiInput{
				Body: aws.String(string(yamlData)),
				//Basepath:       aws.String(""),
				FailOnWarnings: aws.Bool(FailOnWarnings),
			})
			if err != nil {
				panic(err)
			}
			stageInput := &apigatewayv2.CreateStageInput{
				ApiId:      importedRestAPI.ApiId,
				AutoDeploy: aws.Bool(true),
				//DeploymentId: deploymentResult.DeploymentId,
				StageName: aws.String(stageNamePtr),
			}
			_, err = svc.CreateStage(stageInput)
			if err != nil {
				panic(err)
			}

			//API Mapping
			createApiMappingInput := &apigatewayv2.CreateApiMappingInput{
				DomainName: domainNamePtr,
				ApiId:      importedRestAPI.ApiId,
				Stage:      aws.String(stageNamePtr),
				//ApiMappingKey: aws.String(dnsbase + "/" + s),
			}

			_, err = svc.CreateApiMapping(createApiMappingInput)
			if err != nil {
				log.Println(err)
			}
		}

		// createApiGatewayV2IntegrationWithVpcLink(*importedRestAPI.ApiId, vpcLinkPtr, svc)

		//Create Stage

	}
}

func restApiExist(apigatewaynamePtr string, restAPIsList *apigatewayv2.GetApisOutput) (string, bool) {
	for _, restAPI := range restAPIsList.Items {
		if *restAPI.Name == apigatewaynamePtr {
			return *restAPI.ApiId, true
		}
	}
	return "", false
}
