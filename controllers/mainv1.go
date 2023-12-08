package controllers

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"gopkg.in/yaml.v3"
)

func StartV1() {
	var sess *session.Session
	var svc *apigateway.APIGateway
	if os.Args[1] == "delete" {
		sess = SwitchCredentials(os.Args[2])
		svc = apigateway.New(sess)
		err := DeleteAllAPIsV1(svc)
		if err != nil {
			log.Println(err)
		}
		return
	}
	if len(os.Args) < 5 {
		log.Println("Usage: go run main.go <ApiGatewayVersion> <path> <vpcLinkID> <profile> <stageName>")
		return
	}

	path := os.Args[2]
	vpcLinkID := os.Args[3]
	profile := os.Args[4]
	stageName := os.Args[5]

	sess = SwitchCredentials(profile)
	svc = apigateway.New(sess)
	var currentDirectory string
	var dnsbase string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			currentDirectory = info.Name()
			dnsbase = currentDirectory
			return nil
		}

		if filepath.Ext(path) != ".json" {
			return nil
		}

		if filepath.Base(path) == "config.json" {
			return nil
		}

		jsonData, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var data map[string]interface{}
		err = json.Unmarshal(jsonData, &data)
		if err != nil {
			return err
		}

		DeleteSecurity(data)

		err = AddPrivateLinkToRoutes(data, vpcLinkID)
		if err != nil {
			return err
		}

		err = AddVersionToTitle(data, info.Name())
		if err != nil {
			return err
		}

		err = AddStageToRoutes(data, stageName)
		if err != nil {
			return err
		}

		yamlData, err := yaml.Marshal(data)
		if err != nil {
			return err
		}

		yamlPath := path[:len(path)-5] + ".yaml"
		err = os.WriteFile(yamlPath, yamlData, 0644)
		if err != nil {
			return err
		}
		log.Println("Converted", path, "to", yamlPath)

		err = ImportOrUpdateAPIV1(svc, yamlPath, data, stageName, dnsbase, info)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Println("Error:", err)
		return
	}
}
