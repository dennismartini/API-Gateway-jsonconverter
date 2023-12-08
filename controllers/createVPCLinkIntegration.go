package controllers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
)

func CreateApiGatewayV2IntegrationWithVpcLink(apiID, vpcLinkID string, svc *apigatewayv2.ApiGatewayV2) (*apigatewayv2.CreateIntegrationOutput, error) {

	// Crie o objeto Integration para a integração VPC link
	integration := &apigatewayv2.CreateIntegrationInput{
		Description:          aws.String("apisitesloja integration"),
		ApiId:                aws.String(apiID),
		ConnectionId:         aws.String(vpcLinkID),
		ConnectionType:       aws.String("VPC_LINK"),
		IntegrationType:      aws.String("HTTP_PROXY"),
		PayloadFormatVersion: aws.String("1.0"),
		IntegrationMethod:    aws.String("ANY"),
		IntegrationUri:       aws.String("arn:aws:elasticloadbalancing:us-east-1:123123123213:listener/net/abb39ec342asfasfasf4ed12d98u98usd/109dd4ea79648ae0/4185f7182771cb96"),
	}

	// Crie a integração VPC link no API Gateway v2
	createdintegration, err := svc.CreateIntegration(integration)
	if err != nil {
		return nil, err
	}

	return createdintegration, nil
}

// func updateApiGatewayV2MethodWithIntegration(svc *apigatewayv2.ApiGatewayV2, apiID, routeID, integrationID string) error {

// 	// Atualize o método no API Gateway v2
// 	_, err := svc.UpdateIntegrationResponse(&apigatewayv2.UpdateIntegrationResponseInput{
// 		ApiId:         aws.String(apiID),
// 		IntegrationId: aws.String(integrationID),
// 		//IntegrationResponseId: aws.String(integrationResponse.IntegrationStatusCode),
// 		ResponseParameters: integrationResponse.ResponseParameters,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	// Atualize o método no API Gateway v2
// 	_, err = svc.UpdateRouteResponse(&apigatewayv2.UpdateRouteResponseInput{
// 		ApiId:           aws.String(apiID),
// 		RouteId:         aws.String(routeID),
// 		RouteResponseId: aws.String(integrationResponse.IntegrationStatusCode),
// 		MethodResponse:  methodResponse,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
