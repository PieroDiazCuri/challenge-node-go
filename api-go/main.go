package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/jwt/v2"
	"gonum.org/v1/gonum/mat"
)

// MatrixRequest represents the input matrix
type MatrixRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

// QRResponse represents the QR factorization result
type QRResponse struct {
	Q [][]float64 `json:"q"`
	R [][]float64 `json:"r"`
}

// StatisticsRequest represents the data sent to Node.js API
type StatisticsRequest struct {
	Q        [][]float64 `json:"q"`
	R        [][]float64 `json:"r"`
	Original [][]float64 `json:"original"`
}

// ErrorResponse represents error responses
type ErrorResponse struct {
	Error string `json:"error"`
}

// JWT secret key
var jwtSecret = []byte("your-secret-key")

func main() {
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	// JWT middleware for protected routes
	jwtMiddleware := jwt.New(jwt.Config{
		SigningKey: jwtSecret,
	})

	// Routes
	api := app.Group("/api")

	// Public health check
	api.Get("/health", func(contexto *fiber.Ctx) error {
		return contexto.JSON(fiber.Map{
			"status":  "healthy",
			"service": "go-api",
		})
	})

	// Protected QR factorization endpoint
	api.Post("/qr-factorization", jwtMiddleware, manejarFactorizacionQR)

	// Generate JWT token endpoint (for testing)
	api.Post("/login", manejarLogin)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

// manejarLogin genera un token JWT para pruebas
func manejarLogin(contexto *fiber.Ctx) error {
	token := jwt.NewToken(jwt.Config{
		SigningKey: jwtSecret,
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return contexto.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to generate token",
		})
	}

	return contexto.JSON(fiber.Map{
		"token": tokenString,
	})
}

// manejarFactorizacionQR realiza la factorizacion QR y envia el resultado a Node.js
func manejarFactorizacionQR(contexto *fiber.Ctx) error {
	var solicitudMatriz MatrixRequest
	if err := contexto.BodyParser(&solicitudMatriz); err != nil {
		return contexto.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate matrix
	if len(solicitudMatriz.Matrix) == 0 {
		return contexto.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Matrix cannot be empty",
		})
	}

	cantidadFilas := len(solicitudMatriz.Matrix)
	cantidadColumnas := len(solicitudMatriz.Matrix[0])

	// Check if matrix is rectangular
	for indiceFila := 1; indiceFila < cantidadFilas; indiceFila++ {
		if len(solicitudMatriz.Matrix[indiceFila]) != cantidadColumnas {
			return contexto.Status(http.StatusBadRequest).JSON(ErrorResponse{
				Error: "Matrix must be rectangular",
			})
		}
	}

	// Convert to gonum matrix
	matrizOrigen := mat.NewDense(cantidadFilas, cantidadColumnas, nil)
	for indiceFila := 0; indiceFila < cantidadFilas; indiceFila++ {
		for indiceColumna := 0; indiceColumna < cantidadColumnas; indiceColumna++ {
			matrizOrigen.Set(indiceFila, indiceColumna, solicitudMatriz.Matrix[indiceFila][indiceColumna])
		}
	}

	// Perform QR factorization
	matrizOrtogonal := mat.NewDense(cantidadFilas, cantidadFilas, nil)
	matrizTriangularSuperior := mat.NewDense(cantidadFilas, cantidadColumnas, nil)

	// Use QR decomposition
	descomposicionQR := new(mat.QR)
	descomposicionQR.Factorize(matrizOrigen)

	// Extract Q and R matrices
	descomposicionQR.QTo(matrizOrtogonal)
	descomposicionQR.RTo(matrizTriangularSuperior)

	// Convert back to slices
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

	// Send to Node.js API for statistics
	solicitudEstadisticas := StatisticsRequest{
		Q:        matrizQ,
		R:        matrizR,
		Original: solicitudMatriz.Matrix,
	}

	respuestaEstadisticas, err := enviarAApiNodeJS(contexto, solicitudEstadisticas)
	if err != nil {
		log.Printf("Error sending to Node.js API: %v", err)
		// Return QR results even if statistics fail
		return contexto.JSON(QRResponse{
			Q: matrizQ,
			R: matrizR,
		})
	}

	// Return combined results
	return contexto.JSON(fiber.Map{
		"qr": QRResponse{
			Q: matrizQ,
			R: matrizR,
		},
		"statistics": respuestaEstadisticas,
	})
}

// enviarAApiNodeJS envia los datos a la API de Node.js para calcular estadisticas
func enviarAApiNodeJS(contexto *fiber.Ctx, solicitudEstadisticas StatisticsRequest) (map[string]interface{}, error) {
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
