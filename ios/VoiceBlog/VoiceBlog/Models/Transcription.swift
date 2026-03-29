import Foundation

struct Transcription: Identifiable, Sendable {
    let id: Int64
    let userId: Int64
    let fullText: String
    let segmentsJson: [SegmentPayload]
    let createdAt: Date
    let updatedAt: Date
}

extension Transcription: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case fullText = "full_text"
        case segmentsJson = "segments_json"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct TranscriptionCreateRequest: Sendable {
    let fullText: String
    let segmentsJson: [SegmentPayload]
}

extension TranscriptionCreateRequest: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case fullText = "full_text"
        case segmentsJson = "segments_json"
    }
}
