# test-node-react-go


# PASO 1: Navegar al Proyecto
Apply
Run

# PASO 2: Levantar con Docker (RECOMENDADO)
Apply
Run
Resultado: Se levantan 3 servicios:
�� Frontend: http://localhost:5173
🔧 API Go: http://localhost:8080
 API Node.js: http://localhost:3000

# PASO 3: Probar las APIs
A. Generar Token JWT
Apply
Run
Respuesta: {"token": "eyJhbGciOiJIUzI1NiIs..."}
B. Probar Factorización QR
Apply
Run
C. Probar Health Checks
Apply
Run

# PASO 4: Usar el Frontend
Abrir navegador: http://localhost:5173
Generar Token: Click en "Generate Token"
Ingresar matriz: [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
Calcular: Click en "Calculate QR Factorization"
Ver resultados: Matrices Q, R + estadísticas

�� ALTERNATIVA: Sin Docker (Desarrollo Local)
Si no tienes Docker:
Apply
Run
** bash

#API Go
cd api-go
go mod download
go run main.go

#API Node.js (nueva terminal)
cd api-nodejs
npm install
npm start

#Frontend (nueva terminal)
cd frontend
npm install
npm run dev

*************************************************
image.png