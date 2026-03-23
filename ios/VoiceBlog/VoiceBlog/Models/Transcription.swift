import Foundation

struct Transcription: Codable, Identifiable, Sendable {
    let id: Int64
    let userId: Int64
    let fullText: String
    let segmentsJson: [SegmentPayload]
    let createdAt: Date
    let updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case fullText = "full_text"
        case segmentsJson = "segments_json"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct TranscriptionCreateRequest: Codable, Sendable {
    let fullText: String
    let segmentsJson: [SegmentPayload]

    enum CodingKeys: String, CodingKey {
        case fullText = "full_text"
        case segmentsJson = "segments_json"
    }
}
