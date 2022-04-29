package main

import (
	"encoding/base64"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"image"
	"image/png"
	"bytes"
	"embed"
	"text/template"
	"log"
	"fmt"
    "encoding/json"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

//go:embed og-template.svg
var f embed.FS

type requestData struct {
	Params map[string]string
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	w, h := 1200, 630

	// Load SVG template
	in, _ := f.ReadFile("og-template.svg")
	tmpl, err := template.New("template").Parse(string(in))
	if err != nil { return nil, err }

	// Populate SVG template
	var out bytes.Buffer
	err = tmpl.Execute(&out, requestData{request.QueryStringParameters})
	logParams, _ := json.Marshal(request.QueryStringParameters)
	log.Println(fmt.Sprintf("populating template with data: %s", string(logParams)))
	if err != nil { return nil, err }

	// Load SVG data into rasteriser
	icon, err := oksvg.ReadIconStream(&out)
	if err != nil { return nil, err }
	icon.SetTarget(0, 0, float64(w), float64(h))
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())), 1)

	// Convert to PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, rgba)
	if err != nil {
	  return nil, err
	}

	// Return PNG image body
	return &events.APIGatewayProxyResponse{
		StatusCode:        200,
		Headers:           map[string]string{"Content-Type": "image/png"},
		MultiValueHeaders: http.Header{"Set-Cookie": {"Ding", "Ping"}},
		Body:              base64.StdEncoding.EncodeToString(buf.Bytes()),
		IsBase64Encoded:   true,
	}, nil
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(handler)
}
