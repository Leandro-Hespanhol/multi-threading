package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type BrasilAPIResponse struct {
	CEP      string `json:"cep"`
	State    string `json:"state"`
	City     string `json:"city"`
	District string `json:"district"`
	Street   string `json:"street"`
	Service  string `json:"service"`
}

type ViaCEPResponse struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
	IBGE        string `json:"ibge"`
	GIA         string `json:"gia"`
	DDD         string `json:"ddd"`
	SIAFI       string `json:"siafi"`
}

type APIResponse struct {
	Source string
	Data   interface{}
}

func main() {
	cep := os.Args[1]

	c := make(chan APIResponse, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() {
		req, err := http.NewRequestWithContext(ctx, "GET", "https://brasilapi.com.br/api/cep/v1/"+cep, nil)
		if err != nil {
			log.Printf("Error calling brasil api")
		}
		resp, _ := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var brasilAPI BrasilAPIResponse
		json.Unmarshal(body, &brasilAPI)
		c <- APIResponse{Source: "Brasil API", Data: brasilAPI}
	}()

	go func() {
		req, err := http.NewRequestWithContext(ctx, "GET", "http://viacep.com.br/ws/"+cep+"/json/", nil)
		if err != nil {
			log.Printf("Error calling brasil api")
		}
		resp, _ := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var viaCEP ViaCEPResponse
		json.Unmarshal(body, &viaCEP)
		c <- APIResponse{Source: "ViaCEP", Data: viaCEP}
	}()

	select {
	case result := <-c:
		displayResult(result)
	case <-ctx.Done():
		fmt.Println("Timeout: não foi possível obter resposta de nenhuma API em 1 segundo.")
	}
}

func displayResult(result APIResponse) {
	fmt.Printf("Resposta mais rápida veio da: %s\n", result.Source)
	fmt.Println("Dados do endereço:")

	switch data := result.Data.(type) {
	case BrasilAPIResponse:
		fmt.Printf("CEP: %s\n", data.CEP)
		fmt.Printf("Estado: %s\n", data.State)
		fmt.Printf("Cidade: %s\n", data.City)
		fmt.Printf("Bairro: %s\n", data.District)
		fmt.Printf("Logradouro: %s\n", data.Street)
	case ViaCEPResponse:
		fmt.Printf("CEP: %s\n", data.CEP)
		fmt.Printf("Estado: %s\n", data.UF)
		fmt.Printf("Cidade: %s\n", data.Localidade)
		fmt.Printf("Bairro: %s\n", data.Bairro)
		fmt.Printf("Logradouro: %s\n", data.Logradouro)
		if data.Complemento != "" {
			fmt.Printf("Complemento: %s\n", data.Complemento)
		}
	}
}
