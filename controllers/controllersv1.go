package controllers

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
)

// A function to see if an api is already deployed
func IsApiDeployedV1(svc *apigateway.APIGateway, name string) (bool, error) {
	input := &apigateway.GetRestApisInput{}
	result, err := svc.GetRestApis(input)
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

func AddPrivateLinkToRoutesV1(data map[string]interface{}, vpcLinkID string) error {
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
				"type":           "HTTP",
				"uri":            "http://abb39ec342asfasfasf4ed12d98u98usd-109dd4ea79648ae0.elb.us-east-1.amazonaws.com",
				"connectionType": "VPC_LINK",
				"connectionId":   vpcLinkID,
				"httpMethod":     method,
			}
		}
	}

	return nil
}

func AddStageToRoutesV1(data map[string]interface{}, stageName string) error {
	if _, ok := data["x-amazon-apigateway-deployment"]; ok {
		return fmt.Errorf("error: 'x-amazon-apigateway-deployment' key already exists in JSON data")
	}

	data["x-amazon-apigateway-deployment"] = map[string]interface{}{
		"description": "Deployment description",
		"stageName":   stageName,
	}

	return nil
}

func AddVersionToTitleV1(data map[string]interface{}, fileName string) error {
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

func SwitchCredentialsV1(profile string) *session.Session {
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

func CreateApiMappingV1(svc *apigateway.APIGateway, domainName string, stage string, apiId string) error {
	input := &apigateway.CreateBasePathMappingInput{
		BasePath:   aws.String("(none)"),
		DomainName: aws.String(domainName),
		RestApiId:  aws.String(apiId),
		Stage:      aws.String(stage),
	}
	_, err := svc.CreateBasePathMapping(input)
	if err != nil {
		return err
	}

	log.Println("API Mapping created successfully")
	return nil
}

func DeleteAllAPIsV1(svc *apigateway.APIGateway) error {
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

func ImportOrUpdateAPIV1(svc *apigateway.APIGateway, file string, data map[string]interface{}, stageName string, dnsbase string, filename fs.FileInfo) error {

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

	apiInput := &apigateway.GetRestApisInput{Limit: aws.Int64(500)}
	result, err := svc.GetRestApis(apiInput)
	if err != nil {
		return err
	}

	var found bool
	for _, api := range result.Items {
		if *api.Name == title {
			log.Println("Updating " + *api.Name)
			found = true
			_, err = svc.PutRestApi(&apigateway.PutRestApiInput{
				Body:           yamlData,
				RestApiId:      api.Id,
				Mode:           aws.String("overwrite"),
				FailOnWarnings: aws.Bool(false),
			})
			if err != nil {
				log.Println(err)
			}
			// create a deployment
			log.Println("Deployng", *api.Name)

			deploymentInput := &apigateway.CreateDeploymentInput{
				RestApiId: aws.String(*api.Id),
				StageName: aws.String(stageName),
			}

			deploymentResult, err := svc.CreateDeployment(deploymentInput)
			if err != nil {
				return err
			}

			//Create stage if doesn't exist
			stageInput := &apigateway.CreateStageInput{
				DeploymentId: deploymentResult.Id,
				RestApiId:    api.Id,
				StageName:    aws.String(stageName),
			}
			_, err = svc.CreateStage(stageInput)
			if err != nil {
				log.Println(err)
			}

			// Create api mapping
			s := strings.TrimSuffix(filename.Name(), ".json")
			log.Print(dnsbase + "/" + s)
			createApiMappingInput := &apigateway.CreateBasePathMappingInput{
				DomainName: aws.String("api.develop.asdasddasd.com.br"),
				BasePath:   aws.String(dnsbase + "/" + s),
				RestApiId:  api.Id,
				Stage:      aws.String(stageName),
			}

			_, err = svc.CreateBasePathMapping(createApiMappingInput)
			if err != nil {
				log.Println(err)
			}

			log.Println("Updated API", *api.Name)
			break
		}
	}

	if !found {
		log.Println("Importing " + file)
		importedRestAPI, err := svc.ImportRestApi(&apigateway.ImportRestApiInput{
			Body:           yamlData,
			FailOnWarnings: aws.Bool(false),
		})
		if err != nil {
			return err
		}

		deploymentInput := &apigateway.CreateDeploymentInput{
			RestApiId: aws.String(*importedRestAPI.Id),
			StageName: aws.String(stageName),
		}

		deploymentResult, err := svc.CreateDeployment(deploymentInput)
		if err != nil {
			return err
		}

		stageInput := &apigateway.CreateStageInput{
			DeploymentId: deploymentResult.Id,
			RestApiId:    importedRestAPI.Id,
			StageName:    aws.String(stageName),
		}
		_, err = svc.CreateStage(stageInput)
		if err != nil {
			return err
		}

		createApiMappingInput := &apigateway.CreateBasePathMappingInput{
			DomainName: aws.String("api.develop.asdasdasdasd.com.br"),
			BasePath:   aws.String(dnsbase + "/" + filename.Name()),
			RestApiId:  importedRestAPI.Id,
			Stage:      aws.String(stageName),
		}

		_, err = svc.CreateBasePathMapping(createApiMappingInput)
		if err != nil {
			log.Println(err)
		}

		log.Println("Updated API", *importedRestAPI.Name)
		// create a deployment
		// log.Println("Deployng", *importedRestAPI.Name)
		// deploymentInput = &apigateway.CreateDeploymentInput{
		// 	RestApiId: aws.String(*importedRestAPI.Id),
		// 	StageName: aws.String(stageName),
		// }
		// _, err = svc.CreateDeployment(deploymentInput)
		// if err != nil {
		// 	return err
		// }
		if err != nil {
			log.Println(err)
		}

		log.Println("Imported", file, "to API")
	}

	// log.Println("API Mapping created successfully")
	return nil
}

func DeleteSecurityV1(data map[string]interface{}) {
	delete(data, "securitySchemes")
	delete(data, "security")
}
