package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"prueba-tecnica/api-go/src/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gonum.org/v1/gonum/mat"
)

type Service struct {
	JWTSecret []byte
}

func New(jwtSecret []byte) *Service {
	return &Service{JWTSecret: jwtSecret}
}

func (service *Service) Health(contexto *fiber.Ctx) error {
	return contexto.JSON(fiber.Map{
		"status":  "healthy",
		"service": "go-api",
	})
}

func (service *Service) Login(contexto *fiber.Ctx) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": "test-user",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(service.JWTSecret)
	if err != nil {
		return contexto.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to generate token",
		})
	}

	return contexto.JSON(fiber.Map{
		"token": tokenString,
	})
}

func (service *Service) FactorizacionQR(contexto *fiber.Ctx) error {
	var solicitudMatriz models.MatrixRequest
	if err := contexto.BodyParser(&solicitudMatriz); err != nil {
		return contexto.Status(http.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid request body"})
	}

	if len(solicitudMatriz.Matrix) == 0 {
		return contexto.Status(http.StatusBadRequest).JSON(models.ErrorResponse{Error: "Matrix cannot be empty"})
	}

	cantidadFilas := len(solicitudMatriz.Matrix)
	cantidadColumnas := len(solicitudMatriz.Matrix[0])

	for indiceFila := 1; indiceFila < cantidadFilas; indiceFila++ {
		if len(solicitudMatriz.Matrix[indiceFila]) != cantidadColumnas {
			return contexto.Status(http.StatusBadRequest).JSON(models.ErrorResponse{Error: "Matrix must be rectangular"})
		}
	}

	matrizOrigen := mat.NewDense(cantidadFilas, cantidadColumnas, nil)
	for indiceFila := 0; indiceFila < cantidadFilas; indiceFila++ {
		for indiceColumna := 0; indiceColumna < cantidadColumnas; indiceColumna++ {
			matrizOrigen.Set(indiceFila, indiceColumna, solicitudMatriz.Matrix[indiceFila][indiceColumna])
		}
	}

	matrizOrtogonal := mat.NewDense(cantidadFilas, cantidadFilas, nil)
	matrizTriangularSuperior := mat.NewDense(cantidadFilas, cantidadColumnas, nil)
	decomposicionQR := new(mat.QR)
	decomposicionQR.Factorize(matrizOrigen)
	decomposicionQR.QTo(matrizOrtogonal)
	decomposicionQR.RTo(matrizTriangularSuperior)

	matrizQ := make([][]float64, cantidadFilas)
	matrizR := make([][]float64, cantidadFilas)

	for indiceFila := 0; indiceFila < cantidadFilas; indiceFila++ {
		matrizQ[indiceFila] = make([]float64, cantidadFilas)
		matrizR[indiceFila] = make([]float64, cantidadColumnas)

		for indiceColumna := 0; indiceColumna < cantidadFilas; indiceColumna++ {
			matrizQ[indiceFila][indiceColumna] = matrizOrtogonal.At(indiceFila, indiceColumna)
		}

		for indiceColumna := 0; indiceColumna < cantidadColumnas; indiceColumna++ {
			matrizR[indiceFila][indiceColumna] = matrizTriangularSuperior.At(indiceFila, indiceColumna)
		}
	}

	solicitudEstadisticas := models.StatisticsRequest{
		Q:        matrizQ,
		R:        matrizR,
		Original: solicitudMatriz.Matrix,
	}

	respuestaEstadisticas, err := service.enviarAApiNodeJS(contexto, solicitudEstadisticas)
	if err != nil {
		log.Printf("Error sending to Node.js API: %v", err)
		return contexto.JSON(models.QRResponse{Q: matrizQ, R: matrizR})
	}

	return contexto.JSON(fiber.Map{
		"qr":         models.QRResponse{Q: matrizQ, R: matrizR},
		"statistics": respuestaEstadisticas,
	})
}

func (service *Service) enviarAApiNodeJS(contexto *fiber.Ctx, solicitudEstadisticas models.StatisticsRequest) (map[string]interface{}, error) {
	urlApiNode := os.Getenv("NODEJS_API_URL")
	if urlApiNode == "" {
		urlApiNode = "http://localhost:3000"
	}

	cuerpoSolicitud, err := json.Marshal(solicitudEstadisticas)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %v", err)
	}

	solicitudHTTP, err := http.NewRequest(http.MethodPost, urlApiNode+"/api/statistics", bytes.NewBuffer(cuerpoSolicitud))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	solicitudHTTP.Header.Set("Content-Type", "application/json")
	solicitudHTTP.Header.Set("Authorization", contexto.Get("Authorization"))

	clienteHTTP := &http.Client{}
	respuestaHTTP, err := clienteHTTP.Do(solicitudHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer respuestaHTTP.Body.Close()

	if respuestaHTTP.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Node.js API returned status: %d", respuestaHTTP.StatusCode)
	}

	var respuestaEstadisticas map[string]interface{}
	if err := json.NewDecoder(respuestaHTTP.Body).Decode(&respuestaEstadisticas); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return respuestaEstadisticas, nil
}
