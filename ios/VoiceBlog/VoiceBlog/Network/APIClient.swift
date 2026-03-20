import Foundation

enum APIError: LocalizedError {
    case unauthorized
    case serverError(statusCode: Int, body: String)
    case networkError(Error)

    var errorDescription: String? {
        switch self {
        case .unauthorized:
            return "認証に失敗しました"
        case .serverError(let code, let body):
            return "サーバーエラー (\(code)): \(body)"
        case .networkError(let error):
            return error.localizedDescription
        }
    }
}

actor APIClient {
    static let shared = APIClient()

    // TODO: 環境に合わせて変更
    private let baseURL = URL(string: "http://localhost:8080")!

    private func request<T: Decodable>(
        path: String,
        method: String = "GET",
        token: String? = nil
    ) async throws -> T {
        let url = baseURL.appendingPathComponent(path)
        var req = URLRequest(url: url)
        req.httpMethod = method
        req.setValue("application/json", forHTTPHeaderField: "Accept")

        if let token {
            req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        let (data, response) = try await URLSession.shared.data(for: req)

        guard let http = response as? HTTPURLResponse else {
            throw APIError.networkError(URLError(.badServerResponse))
        }

        switch http.statusCode {
        case 200..<300:
            return try JSONDecoder().decode(T.self, from: data)
        case 401:
            throw APIError.unauthorized
        default:
            let body = String(data: data, encoding: .utf8) ?? ""
            throw APIError.serverError(statusCode: http.statusCode, body: body)
        }
    }

    func fetchMe(token: String) async throws -> User {
        try await request(path: "/me", token: token)
    }
}
