package controllers

import (
	"encoding/json"
	"os"
	"strings"
)

type SecuritySchemesStruct struct {
	JwtAuthorizer struct {
		Type                        string `json:"type"`
		Name                        string `json:"name"`
		In                          string `json:"in"`
		XAmazonApigatewayAuthorizer struct {
			IdentitySource                 string `json:"identitySource"`
			AuthorizerURI                  string `json:"authorizerUri"`
			AuthorizerPayloadFormatVersion string `json:"authorizerPayloadFormatVersion"`
			AuthorizerResultTTLInSeconds   int    `json:"authorizerResultTtlInSeconds"`
			Type                           string `json:"type"`
			EnableSimpleResponses          bool   `json:"enableSimpleResponses"`
		} `json:"x-amazon-apigateway-authorizer"`
	} `json:"jwt-authorizer"`
}

// Convert entire SecuritySchemesValues variable content to json using SecuritySchemesStruct type and add to components section
func AddSecuritySchemesToComponents(accountId string, data map[string]interface{}) error {
	var SecuritySchemesValues string = `
	{     
		"jwt-authorizer": {
		"type": "apiKey",
		"name": "Authorization",
		"in": "header",
		"x-amazon-apigateway-authorizer": {
		  "identitySource": "$request.header.Authorization",
		  "authorizerUri": "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:` + accountId + `:function:IdentityAuthorizer/invocations",
		  "authorizerPayloadFormatVersion": "1.0",
		  "authorizerResultTtlInSeconds": 300,
		  "type": "request",
		  "enableSimpleResponses": false
		}
	  }
	}
	`
	var securitySchemes SecuritySchemesStruct
	err := json.Unmarshal([]byte(SecuritySchemesValues), &securitySchemes)
	if err != nil {
		panic("Error unmarshalling JSON: " + err.Error())
	}

	components := data["components"].(map[string]interface{})
	components["securitySchemes"] = securitySchemes
	return nil
}

// A function to add '"security" : [ {          "jwt-authorizer" : [ ]        } ],' to each method for each route
func AddAuthorizerToRoutes(data map[string]interface{}) error {
	paths := data["paths"].(map[string]interface{})
	for _, path := range paths {
		path := path.(map[string]interface{})
		for _, method := range path {
			method := method.(map[string]interface{})
			security := []map[string]interface{}{
				{
					"jwt-authorizer": []interface{}{},
				},
			}
			method["security"] = security
		}
	}
	return nil
}

// A function to lower case all paths in the paths section ignoring text inside curly braces
func LowerCasePaths(data map[string]interface{}) error {
	paths := data["paths"].(map[string]interface{})
	for key, path := range paths {
		delete(paths, key)
		newKey := strings.Builder{}
		isInBrackets := false
		for _, c := range key {
			if c == '{' {
				isInBrackets = true
			}
			if !isInBrackets {
				newKey.WriteRune(rune(strings.ToLower(string(c))[0]))
			} else {
				newKey.WriteRune(c)
			}
			if c == '}' {
				isInBrackets = false
			}
		}
		paths[newKey.String()] = path
	}
	return nil
}

// func LowerCasePaths(data map[string]interface{}) error {
// 	paths := data["paths"].(map[string]interface{})
// 	for key, path := range paths {
// 		delete(paths, key)
// 		paths[strings.ToLower(key)] = path
// 	}
// 	return nil
// }

func ChangeServersURL(data map[string]interface{}, url string) map[string]interface{} {
	servers := data["servers"].([]interface{})
	server := servers[0].(map[string]interface{})

	// Atualize o valor da chave "url"
	server["url"] = url
	return data
}

func jsonToAWSChanges(lbarnuri string, accountId string, vpclink string, profile string, stagename string, apigatewayname string, indexnumber string, targetfilepath string, serverURL string) []byte {
	jsonData, err := os.ReadFile(targetfilepath)
	if err != nil {
		panic("Error reading file: " + err.Error())
	}

	var data map[string]interface{}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		panic("Error unmarshalling JSON: " + err.Error())
	}

	//Change Title Name
	ChangeTitle(data, apigatewayname)

	data = ChangeServersURL(data, serverURL)
	//abb39ec342asfasfasf4ed12d98u98usd-109dd4ea79648ae0.elb.us-east-1.amazonaws.com
	//"https://"+vpclink+".execute-api.us-east-1.amazonaws.com/"+stagename
	data = DeleteSecurity(data)
	err = AddPrivateLinkToContent(data, vpclink, lbarnuri)
	if err != nil {
		panic("Error adding private link to content: " + err.Error())
	}

	err = LowerCasePaths(data)
	if err != nil {
		panic("Error lower casing paths: " + err.Error())
	}

	err = PathCorrection(data)
	if err != nil {
		panic("Error correcting path: " + err.Error())
	}

	err = AddPrivateLinkToRoutes(data, vpclink)
	if err != nil {
		panic("Error adding private link to routes: " + err.Error())
	}

	err = AddAuthorizerToRoutes(data)
	if err != nil {
		panic("Error adding Authorizer to routes: " + err.Error())
	}

	err = AddSecuritySchemesToComponents(accountId, data)
	if err != nil {
		panic("Error adding security schemes to components: " + err.Error())
	}

	dataMarshalled, err := json.Marshal(data)
	if err != nil {
		panic("Error marshalling JSON: " + err.Error())
	}
	os.WriteFile(targetfilepath, dataMarshalled, 0644)
	return dataMarshalled
}
