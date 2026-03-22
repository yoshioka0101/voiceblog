import Foundation

enum APIError: LocalizedError {
    case unauthorized
    case invalidConfiguration(String)
    case serverError(statusCode: Int, body: String)
    case networkError(Error)

    var errorDescription: String? {
        switch self {
        case .unauthorized:
            return "認証に失敗しました"
        case .invalidConfiguration(let message):
            return message
        case .serverError(let code, let body):
            return "サーバーエラー (\(code)): \(body)"
        case .networkError(let error):
            return error.localizedDescription
        }
    }
}

actor APIClient {
    static let shared = APIClient()

    private let baseURLResult: Result<URL, APIError>

    init(bundle: Bundle = .main) {
        do {
            baseURLResult = .success(try bundle.apiBaseURL())
        } catch {
            baseURLResult = .failure(.invalidConfiguration(error.localizedDescription))
        }
    }

    private func request<T: Decodable>(
        path: String,
        method: String = "GET",
        token: String? = nil
    ) async throws -> T {
        let baseURL = try baseURLResult.get()
        let normalizedPath = path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        let url = baseURL.appendingPathComponent(normalizedPath)
        var req = URLRequest(url: url)
        req.httpMethod = method
        req.setValue("application/json", forHTTPHeaderField: "Accept")

        if let token {
            req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        let (data, response): (Data, URLResponse)
        do {
            (data, response) = try await URLSession.shared.data(for: req)
        } catch {
            throw APIError.networkError(error)
        }

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
