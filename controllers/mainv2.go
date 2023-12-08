package controllers

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"gopkg.in/yaml.v2"
)

func StartV2() {
	//Declaração das variáveis
	var currentDirectory string
	var dnsbase string
	var yamlOutput []byte
	var yamlOutputcomponents []byte

	//Verifica se a opção "delete" foi selecionada
	if os.Args[2] == "delete" {
		sess := SwitchCredentials(os.Args[3])
		svc := apigatewayv2.New(sess)
		err := DeleteAllAPIs(svc)
		if err != nil {
			log.Println(err)
		}
		os.Exit(0)
	}

	//Verifica se há argumentos suficientes para executar a conversão
	if len(os.Args) < 5 {
		log.Println("Usage: go run main.go <apiVersion> <path> <vpcLinkID> <profile> <stageName>")
		return
	}

	//Obtém o caminho e o ID do link privado
	path := os.Args[2]
	vpcLinkID := os.Args[3]

	//Executa a função Walk para percorrer todos os arquivos no diretório especificado
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

		err = AddPrivateLinkToRoutes(data, vpcLinkID)
		if err != nil {
			return err
		}

		err = AddFilenameToPath(data, info.Name(), dnsbase)
		if err != nil {
			return err
		}

		//Tratamento de paths
		paths, ok := data["paths"]
		if !ok {
			log.Printf("Error: File %s does not contain paths section\n", path)
			return nil
		}

		//Converte o objeto "paths" em YAML sem o prefixo "paths:"
		yamlData, err := yaml.Marshal(paths)
		if err != nil {
			return err
		}

		//Remove o prefixo "paths:" de cada linha
		yamlString := string(yamlData)
		yamlString = strings.Replace(yamlString, "paths:", "", -1)
		yamlData = []byte(yamlString)

		//Concatena as conversões YAML no arquivo de destino
		yamlOutput = append(yamlOutput, yamlData...)

		log.Println("Converted", path, "to YAML")

		//Tratamento de components
		components, ok := data["components"]
		if !ok {
			log.Printf("Error: File %s does not contain paths section\n", path)
			return nil
		}

		//Converte o objeto "paths" em YAML sem o prefixo "paths:"
		yamlDatacomponents, err := yaml.Marshal(components)
		if err != nil {
			return err
		}

		//Remove o prefixo "paths:" de cada linha
		yamlStringcomponents := string(yamlDatacomponents)
		yamlStringcomponents = strings.Replace(yamlStringcomponents, "paths:", "", -1)
		yamlDatacomponents = []byte(yamlStringcomponents)

		//Concatena as conversões YAML no arquivo de destino
		yamlOutputcomponents = append(yamlOutputcomponents, yamlDatacomponents...)

		log.Println("Converted", components, "to YAML")

		return nil
	})

	if err != nil {
		log.Println("Error:", err)
		return
	}

	//Define o nome do arquivo de destino
	//destFilename := "output.yaml"

	//Cria ou abre o arquivo de destino
	resultFilePath := "paths.yaml"
	resultFilePathcomponents := "components.yaml"
	resultFile, err := os.OpenFile(resultFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	resultFilecomponents, err := os.OpenFile(resultFilePathcomponents, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer resultFile.Close()
	defer resultFilecomponents.Close()
	//Percorre novamente os arquivos para copiar os dados para o arquivo de destino
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
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

		err = AddFilenameToPath(data, info.Name(), dnsbase)
		if err != nil {
			return err
		}

		paths, ok := data["paths"]
		if !ok {
			log.Printf("Error: File %s does not contain paths section\n", path)
			return nil
		}

		//Converte o objeto "paths" em YAML sem o prefixo "paths:"
		yamlData, err := yaml.Marshal(paths)
		if err != nil {
			return err
		}

		//Remove o prefixo "paths:" de cada linha
		yamlString := string(yamlData)
		yamlString = strings.Replace(yamlString, "paths:", "", -1)
		yamlData = []byte(yamlString)

		//Escreve os dados no arquivo de destino
		if _, err = resultFile.Write(yamlData); err != nil {
			return err
		}
		//resultFile.WriteString("\n---\n")

		components, ok := data["components"]
		if !ok {
			log.Printf("Error: File %s does not contain paths section\n", components)
			return nil
		}

		//Converte o objeto "paths" em YAML sem o prefixo "paths:"
		yamlDatacomponents, err := yaml.Marshal(components)
		if err != nil {
			return err
		}

		//Remove o prefixo "paths:" de cada linha
		//yamlStringcomponents := string(yamlDatacomponents)
		//yamlStringcomponents = strings.Replace(yamlStringcomponents, "paths:", "", -1)
		yamlDatacomponents = []byte(yamlDatacomponents)

		//Escreve os dados no arquivo de destino
		if _, err = resultFile.Write(yamlDatacomponents); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Println("Error:", err)
		return
	}

	log.Println("Conversion completed, results saved in", resultFilePath, resultFilePathcomponents)
	ConvertToOpenAPI3(resultFilePath, resultFilePathcomponents)
}
