# Documentação do Conversor de API Management YAML para API Gateway JSON

## Visão Geral
Este software converte arquivos YAML do Azure API Management para JSON compatível com AWS API Gateway, facilitando a migração entre plataformas.

## Pré-requisitos
- Go (Golang)
- Node.js

## Configuração Inicial

### Baixar Dependências
Antes de executar o software, é necessário baixar as dependências. Execute os seguintes comandos para instalar as dependências necessárias:

```bash
# Dependências Golang
go get -u [lista de pacotes Golang]

# Dependências Node.js
npm install [lista de pacotes Node.js]
Estrutura do Projeto
O projeto é dividido em duas partes principais:

Função Principal (main): Processa argumentos da linha de comando e inicia a conversão.
Função jsonToAWSChanges: Realiza a conversão de formato e modifica os dados.
Uso
Utilize o comando go run com os parâmetros apropriados:

Desenvolvimento:
```bash
Copy code
go run .\main.go -action="convert" -source="..\azure\dev\api-manager-sites" -index="1" -vpclinkid="r6hwarningsuvf" -aws-profile="teste-dev" -stage-name="dev" -apigateway-name="api-sites-loja" -targetfile="outputapisitesloja.json" -domain-name="api-sites-loja.develop.testetii.com.br" -deploy="false" -server-url="abb39ec342asfasfasf4ed12d98u98usd-109dd4ea79648ae0.elb.us-east-1.amazonaws.com" -fail-on-warnings="false" -account-id="1625346125436512" -lb-arn-uri="arn:aws:elasticloadbalancing:us-east-1:1625346125436512:listener/net/abb39ec342asfasfasf4ed12d98u98usd/109dd4ea79648ae0/4185f7182771cb96"
```bash
##Configurações de Comando
As flags de linha de comando incluem:

account-id: ID da conta AWS.
lb-arn-uri: ARN do Load Balancer AWS.
action: "convert" para a conversão.
source: Caminho de origem dos YAMLs.
index: Índice para geração dos arquivos.
vpclinkid: ID do VPC Link AWS.
aws-profile: Perfil AWS.
stage-name: Nome do estágio (dev, hml).
apigateway-name: Nome do API Gateway AWS.
targetfile: Caminho do arquivo de destino.
domain-name: Nome do domínio.
deploy: Se deve realizar o deploy.
server-url: URL do servidor.
fail-on-warnings: Se deve falhar em caso de warnings.