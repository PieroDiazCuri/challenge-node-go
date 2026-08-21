import React, { useState } from "react";
import axios from "axios";
import "./App.css";

interface MatrixData {
  q: number[][];
  r: number[][];
}

interface Statistics {
  maxValue: number;
  minValue: number;
  average: number;
  sum: number;
  totalValues: number;
  diagonalMatrices: boolean[];
  matrixNames: string[];
}

interface ApiResponse {
  qr: MatrixData;
  statistics: Statistics;
}

function App() {
  const [matrix, setMatrix] = useState<string>("");
  const [result, setResult] = useState<ApiResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");
  const [token, setToken] = useState<string>("");

  const generateToken = async () => {
    try {
      const response = await axios.post("http://localhost:8080/api/login");
      setToken(response.data.token);
      setError("");
    } catch (err) {
      setError("Failed to generate token");
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      // Parse matrix input
      const matrixData = JSON.parse(matrix);
      
      if (!Array.isArray(matrixData) || matrixData.length === 0) {
        throw new Error("Invalid matrix format");
      }

      const response = await axios.post(
        "http://localhost:8080/api/qr-factorization",
        { matrix: matrixData },
        {
          headers: {
            "Authorization": `Bearer ${token}`,
            "Content-Type": "application/json",
          },
        }
      );

      setResult(response.data);
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 py-8">
      <div className="max-w-6xl mx-auto px-4">
        <h1 className="text-4xl font-bold text-center text-gray-800 mb-8">
          Matrix QR Factorization
        </h1>

        <div className="bg-white rounded-lg shadow-lg p-6 mb-8">
          <h2 className="text-2xl font-semibold mb-4">Generate JWT Token</h2>
          <button
            onClick={generateToken}
            className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
          >
            Generate Token
          </button>
          {token && (
            <div className="mt-4 p-3 bg-green-100 rounded">
              <p className="text-sm text-green-800">Token: {token.substring(0, 20)}...</p>
            </div>
          )}
        </div>

        <div className="bg-white rounded-lg shadow-lg p-6 mb-8">
          <h2 className="text-2xl font-semibold mb-4">Enter Matrix</h2>
          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label className="block text-gray-700 text-sm font-bold mb-2">
                Matrix (JSON format):
              </label>
              <textarea
                value={matrix}
                onChange={(e) => setMatrix(e.target.value)}
                placeholder="[[1, 2, 3], [4, 5, 6], [7, 8, 9]]"
                className="w-full h-32 p-3 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
            <button
              type="submit"
              disabled={loading || !token}
              className="bg-green-500 hover:bg-green-700 disabled:bg-gray-400 text-white font-bold py-2 px-4 rounded"
            >
              {loading ? "Processing..." : "Calculate QR Factorization"}
            </button>
          </form>
        </div>

        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-8">
            {error}
          </div>
        )}

        {result && (
          <div className="space-y-8">
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-2xl font-semibold mb-4">QR Factorization Results</h2>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div>
                  <h3 className="text-lg font-semibold mb-2">Original Matrix</h3>
                  <div className="bg-gray-100 p-3 rounded">
                    <pre className="text-sm">{JSON.stringify(result.qr.q, null, 2)}</pre>
                  </div>
                </div>
                
                <div>
                  <h3 className="text-lg font-semibold mb-2">Q Matrix</h3>
                  <div className="bg-gray-100 p-3 rounded">
                    <pre className="text-sm">{JSON.stringify(result.qr.q, null, 2)}</pre>
                  </div>
                </div>
                
                <div>
                  <h3 className="text-lg font-semibold mb-2">R Matrix</h3>
                  <div className="bg-gray-100 p-3 rounded">
                    <pre className="text-sm">{JSON.stringify(result.qr.r, null, 2)}</pre>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-2xl font-semibold mb-4">Statistics</h2>
              
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="bg-blue-100 p-4 rounded">
                  <h3 className="font-semibold">Max Value</h3>
                  <p className="text-2xl font-bold">{result.statistics.maxValue}</p>
                </div>
                
                <div className="bg-green-100 p-4 rounded">
                  <h3 className="font-semibold">Min Value</h3>
                  <p className="text-2xl font-bold">{result.statistics.minValue}</p>
                </div>
                
                <div className="bg-yellow-100 p-4 rounded">
                  <h3 className="font-semibold">Average</h3>
                  <p className="text-2xl font-bold">{result.statistics.average.toFixed(4)}</p>
                </div>
                
                <div className="bg-purple-100 p-4 rounded">
                  <h3 className="font-semibold">Sum</h3>
                  <p className="text-2xl font-bold">{result.statistics.sum.toFixed(2)}</p>
                </div>
              </div>
              
              <div className="mt-6">
                <h3 className="text-lg font-semibold mb-2">Diagonal Matrix Check</h3>
                <div className="grid grid-cols-3 gap-4">
                  {result.statistics.matrixNames.map((name, index) => (
                    <div key={name} className="bg-gray-100 p-3 rounded">
                      <p className="font-semibold">{name}</p>
                      <p className={result.statistics.diagonalMatrices[index] ? "text-green-600" : "text-red-600"}>
                        {result.statistics.diagonalMatrices[index] ? "Diagonal" : "Not Diagonal"}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
