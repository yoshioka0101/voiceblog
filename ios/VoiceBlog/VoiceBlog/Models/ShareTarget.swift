import Foundation

struct ShareTarget: Identifiable, Sendable {
    let id: Int64
    let articleId: Int64
    let provider: String
    let externalId: String
    let externalUrl: String
    let publishedAt: Date?
    let createdAt: Date
    let updatedAt: Date
}

extension ShareTarget: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case id
        case articleId = "article_id"
        case provider
        case externalId = "external_id"
        case externalUrl = "external_url"
        case publishedAt = "published_at"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct PublishRequest: Sendable {
    let provider: String
}

extension PublishRequest: Codable {}
