import Foundation

struct Article: Identifiable, Sendable {
    let id: Int64
    let userId: Int64
    let promptRunJobId: Int64?
    let title: String
    let content: String
    let deletedAt: Date?
    let createdAt: Date
    let updatedAt: Date
}

extension Article: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case promptRunJobId = "prompt_run_job_id"
        case title
        case content
        case deletedAt = "deleted_at"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct ArticleCreateRequest: Sendable {
    let title: String
    let content: String
    let promptRunJobId: Int64?
}

extension ArticleCreateRequest: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case title
        case content
        case promptRunJobId = "prompt_run_job_id"
    }
}

struct ArticleGenerateRequest: Sendable {
    let transcriptionId: Int64
    let promptId: Int64
}

extension ArticleGenerateRequest: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case transcriptionId = "transcription_id"
        case promptId = "prompt_id"
    }
}

struct ArticleUpdateRequest: Sendable {
    let title: String?
    let content: String?
}

extension ArticleUpdateRequest: Codable {}

struct ArticleDraft: Sendable {
    var title: String = ""
    var content: String = ""

    init() {}

    init(article: Article) {
        title = article.title
        content = article.content
    }
}
