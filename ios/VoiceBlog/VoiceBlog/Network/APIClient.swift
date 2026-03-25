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
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    init(bundle: Bundle = .main) {
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let value = try container.decode(String.self)

            if let date = apiDateFormatterWithFractionalSeconds.date(from: value) {
                return date
            }
            if let date = apiDateFormatter.date(from: value) {
                return date
            }

            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Unsupported date format: \(value)")
        }

        let encoder = JSONEncoder()
        encoder.outputFormatting = [.sortedKeys]
        self.decoder = decoder
        self.encoder = encoder

        do {
            baseURLResult = .success(try bundle.apiBaseURL())
        } catch {
            baseURLResult = .failure(.invalidConfiguration(error.localizedDescription))
        }
    }

    private func performRequest(
        path: String,
        method: String = "GET",
        token: String? = nil,
        body: Data? = nil
    ) async throws -> Data {
        let baseURL = try baseURLResult.get()
        let normalizedPath = path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        let url = baseURL.appendingPathComponent(normalizedPath)
        var req = URLRequest(url: url, timeoutInterval: 30)
        req.httpMethod = method
        req.setValue("application/json", forHTTPHeaderField: "Accept")
        req.httpBody = body

        if body != nil {
            req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        }

        if let token {
            req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        let (data, response): (Data, URLResponse)
        do {
            print("[APIClient] sending \(method) \(url)")
            (data, response) = try await URLSession.shared.data(for: req)
            print("[APIClient] received response for \(method) \(url)")
        } catch {
            print("[APIClient] network error for \(method) \(url): \(error)")
            throw APIError.networkError(error)
        }

        guard let http = response as? HTTPURLResponse else {
            throw APIError.networkError(URLError(.badServerResponse))
        }

        switch http.statusCode {
        case 200..<300:
            return data
        case 401:
            throw APIError.unauthorized
        default:
            let body = String(data: data, encoding: .utf8) ?? ""
            throw APIError.serverError(statusCode: http.statusCode, body: body)
        }
    }

    private func request<T: Decodable>(
        path: String,
        method: String = "GET",
        token: String? = nil,
        body: Data? = nil
    ) async throws -> T {
        let data = try await performRequest(path: path, method: method, token: token, body: body)

        if T.self == EmptyResponse.self, data.isEmpty {
            return EmptyResponse() as! T
        }

        return try decoder.decode(T.self, from: data)
    }

    private func request<Body: Encodable, T: Decodable>(
        path: String,
        method: String,
        token: String,
        body: Body
    ) async throws -> T {
        let encoded = try encoder.encode(body)
        let responseData = try await performRequest(path: path, method: method, token: token, body: encoded)

        if T.self == EmptyResponse.self, responseData.isEmpty {
            return EmptyResponse() as! T
        }

        return try decoder.decode(T.self, from: responseData)
    }

    func getMe(token: String) async throws -> User {
        try await request(path: "/me", token: token)
    }

    func getListPrompts(token: String) async throws -> [Prompt] {
        try await request(path: "/prompts", token: token)
    }

    func createPrompt(request body: PromptCreateRequest, token: String) async throws -> Prompt {
        try await request(path: "/prompts", method: "POST", token: token, body: body)
    }

    func updatePrompt(id: Int64, request body: PromptUpdateRequest, token: String) async throws -> Prompt {
        try await request(path: "/prompts/\(id)", method: "PATCH", token: token, body: body)
    }

    func deletePrompt(id: Int64, token: String) async throws {
        let _: EmptyResponse = try await request(path: "/prompts/\(id)", method: "DELETE", token: token)
    }

    func createTranscription(request body: TranscriptionCreateRequest, token: String) async throws -> Transcription {
        try await request(path: "/transcriptions", method: "POST", token: token, body: body)
    }

    func getListArticles(token: String) async throws -> [Article] {
        try await request(path: "/articles", token: token)
    }

    func getArticle(id: Int64, token: String) async throws -> Article {
        try await request(path: "/articles/\(id)", token: token)
    }

    func createArticle(request body: ArticleCreateRequest, token: String) async throws -> Article {
        try await request(path: "/articles", method: "POST", token: token, body: body)
    }

    func updateArticle(id: Int64, request body: ArticleUpdateRequest, token: String) async throws -> Article {
        try await request(path: "/articles/\(id)", method: "PATCH", token: token, body: body)
    }

    func deleteArticle(id: Int64, token: String) async throws {
        let _: EmptyResponse = try await request(path: "/articles/\(id)", method: "DELETE", token: token)
    }

    func createPromptRunJob(request body: PromptRunJobCreateRequest, token: String) async throws -> PromptRunJob {
        try await request(path: "/prompt-run-jobs", method: "POST", token: token, body: body)
    }

    func getPromptRunJob(id: Int64, token: String) async throws -> PromptRunJob {
        try await request(path: "/prompt-run-jobs/\(id)", token: token)
    }
}

private struct EmptyResponse: Decodable, Sendable {
    nonisolated init() {}
}

private let apiDateFormatter: ISO8601DateFormatter = {
    let formatter = ISO8601DateFormatter()
    formatter.formatOptions = [.withInternetDateTime]
    return formatter
}()

private let apiDateFormatterWithFractionalSeconds: ISO8601DateFormatter = {
    let formatter = ISO8601DateFormatter()
    formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
    return formatter
}()
