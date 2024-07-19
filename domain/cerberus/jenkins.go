package cerberus

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Jenkins struct {
	BaseUrl  string
	Token    string
	Username string
	Password string
}

func (j *Jenkins) BlockIp(ip string) error {
	// Crear los parámetros en la URL
	data := url.Values{}
	data.Set("token", j.Token)
	data.Set("ips", ip)

	// Crear el cliente HTTP
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Crear la petición HTTP POST
	reqUrl := j.BaseUrl + "/job/RiemannAlertIps/buildWithParameters"
	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println("Request failed:", err)
		return err
	}

	// Añadir la autenticación básica a la petición
	req.SetBasicAuth(j.Username, j.Password)

	// Añadir el tipo de contenido
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Hacer la petición
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request failed:", err)
		return err
	}
	defer resp.Body.Close()

	// Leer la respuesta
	log.Println("Response:", resp.StatusCode)
	return nil
}

func (j *Jenkins) BlockAsn(asn string) error {
	host := "https://www.tropipay.com" // TODO: Use the real host
	ratio := "0"                       // TODO: Use the real ratio
	total := "0"                       // TODO: Use the real total
	failed := "0"                      // TODO: Use the real failed

	// Crear los parámetros en la URL
	data := url.Values{}
	data.Set("token", j.Token)
	data.Set("host", host)
	data.Set("asn", asn)
	data.Set("ratio", ratio)
	data.Set("total", total)
	data.Set("failed", failed)

	// Crear el cliente HTTP
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Crear la petición HTTP POST
	reqUrl := j.BaseUrl + "/job/RiemannAlertASN/buildWithParameters"
	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println("Request failed:", err)
		return err
	}

	// Añadir la autenticación básica a la petición
	req.SetBasicAuth(j.Username, j.Password)

	// Añadir el tipo de contenido
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Hacer la petición
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request failed:", err)
		return err
	}
	defer resp.Body.Close()

	// Leer la respuesta
	log.Println("Response:", resp.StatusCode)
	return nil
}
