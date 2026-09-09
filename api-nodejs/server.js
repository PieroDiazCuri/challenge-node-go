const express = require("express");
const cors = require("cors");
const helmet = require("helmet");
const jwt = require("jsonwebtoken");
const app = express();
const PORT = process.env.PORT || 3000;

// JWT secret
const JWT_SECRET = process.env.JWT_SECRET || "your-secret-key";

const apiKey = process.env.API_KEY;
console.log("API_KEY desde Azure:", apiKey);


// Middleware
app.use(helmet());
app.use(
  cors({
    origin: process.env.CORS_ORIGIN || "http://localhost:5173",
  })
);
app.use(express.json({ limit: "10mb" }));

// JWT middleware
const autenticarToken = (peticion, respuesta, siguiente) => {
  const encabezadoAutorizacion = peticion.headers["authorization"];
  const token = encabezadoAutorizacion?.split(" ")[1];

  if (!token) {
    return respuesta.status(401).json({ error: "Access token required" });
  }

  jwt.verify(token, JWT_SECRET, (err, user) => {
    if (err) {
      return respuesta.status(403).json({ error: "Invalid token" });
    }
    peticion.user = user;
    siguiente();
  });
};

// Health check endpoint
app.get("/api/health", (peticion, respuesta) => {
  respuesta.json({
    status: "healthy",
    service: "nodejs-api",
    timestamp: new Date().toISOString(),
  });
});

// Statistics calculation endpoint
app.post("/api/statistics", autenticarToken, (peticion, respuesta) => {
  try {
    const { q, r, original } = peticion.body;

    if (!q || !r || !original) {
      return respuesta.status(400).json({
        error: "Missing required data: q, r, and original matrices are required",
      });
    }

    // Combine las matices para calcular las estadisticas.
    const todasLasMatrices = [q, r, original];
    const todosLosValores = [];

    // Extract all values from matrices
    todasLasMatrices.forEach((matriz) => {
      if (Array.isArray(matriz)) {
        matriz.forEach((fila) => {
          if (Array.isArray(fila)) {
            fila.forEach((valor) => {
              if (typeof valor === "number" && !Number.isNaN(valor)) {
                todosLosValores.push(valor);
              }
            });
          }
        });
      }
    });

    if (todosLosValores.length === 0) {
      return respuesta.status(400).json({
        error: "No valid numeric values found in matrices",
      });
    }

    // Calcule las estadisticas.
    const valorMaximo = Math.max(...todosLosValores);
    const valorMinimo = Math.min(...todosLosValores);
    const sumaTotal = todosLosValores.reduce((acumulado, valor) => acumulado + valor, 0);
    const promedio = sumaTotal / todosLosValores.length;

    // Diagonal matrices.
    const matricesDiagonales = todasLasMatrices.map((matriz) => {
      if (!Array.isArray(matriz) || matriz.length === 0) return false;
      
      const cantidadFilas = matriz.length;
      const cantidadColumnas = matriz[0].length;
      
      // La matriz debe ser cuadrada para ser diagonal
      if (cantidadFilas !== cantidadColumnas) return false;
      
      for (let indiceFila = 0; indiceFila < cantidadFilas; indiceFila++) {
        for (let indiceColumna = 0; indiceColumna < cantidadColumnas; indiceColumna++) {
          if (indiceFila !== indiceColumna && Math.abs(matriz[indiceFila][indiceColumna]) > 1e-10) {
            return false;
          }
        }
      }
      return true;
    });

    const statisticsResponsePayload = {
      statistics: {
        maxValue: valorMaximo,
        minValue: valorMinimo,
        average: Number.parseFloat(promedio.toFixed(6)),
        sum: Number.parseFloat(sumaTotal.toFixed(6)),
        totalValues: todosLosValores.length,
        diagonalMatrices: matricesDiagonales,
        matrixNames: ["Q", "R", "Original"],
      },
      matrices: {
        q,
        r,
        original,
      },
    };

    respuesta.json(statisticsResponsePayload);
  } catch (error) {
    console.error("Error calculating statistics:", error);
    respuesta.status(500).json({
      error: "Internal server error during statistics calculation",
    });
  }
});

// Generate JWT token endpoint (for testing)
app.post("/api/login", (peticion, respuesta) => {
  const token = jwt.sign({ user: "test-user" }, JWT_SECRET, {
    expiresIn: "1h",
  });
  respuesta.json({ token });
});

// Error handling middleware
app.use((error, peticion, respuesta, siguiente) => {
  console.error(error.stack);
  respuesta.status(500).json({ error: "Something went wrong!" });
});

// 404 handler
app.use("*", (peticion, respuesta) => {
  respuesta.status(404).json({ error: "Endpoint not found" });
});

app.listen(PORT, () => {
  console.log(`Node.js API server running on port ${PORT}`);
  console.log(`Health check: http://localhost:${PORT}/api/health`);
});
