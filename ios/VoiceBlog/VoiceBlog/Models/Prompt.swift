import Foundation

struct Prompt: Identifiable, Sendable {
    let id: Int64
    let userId: Int64?
    let name: String
    let body: String
    let isActive: Bool
    let isDefault: Bool
    let createdAt: Date
    let updatedAt: Date

    var isSystemPrompt: Bool {
        userId == nil
    }

}

extension Prompt: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case name
        case body
        case isActive = "is_active"
        case isDefault = "is_default"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct PromptCreateRequest: Sendable {
    let name: String
    let body: String
    let isActive: Bool?
}

extension PromptCreateRequest: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case name
        case body
        case isActive = "is_active"
    }
}

struct PromptUpdateRequest: Sendable {
    let name: String?
    let body: String?
    let isActive: Bool?
}

extension PromptUpdateRequest: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case name
        case body
        case isActive = "is_active"
    }
}
