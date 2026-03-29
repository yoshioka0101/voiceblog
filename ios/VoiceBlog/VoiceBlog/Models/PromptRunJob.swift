import Foundation

struct PromptRunJob: Identifiable, Sendable {
    let id: Int64
    let transcriptionId: Int64
    let promptId: Int64
    let status: String
    let attemptCount: Int
    let nextRunAt: Date
    let errorMessage: String?
    let createdAt: Date
    let generatedTitle: String?
    let generatedContent: String?

    var isInProgress: Bool {
        status == "pending" || status == "running"
    }

    var isFailed: Bool {
        status == "failed"
    }

    var isCompleted: Bool {
        status == "completed"
    }

    var userFacingErrorMessage: String? {
        guard isFailed else {
            return nil
        }

        return "AIで下書きを生成できませんでした。しばらくしてからもう一度お試しください。"
    }
}

extension PromptRunJob: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case id
        case transcriptionId = "transcription_id"
        case promptId = "prompt_id"
        case status
        case attemptCount = "attempt_count"
        case nextRunAt = "next_run_at"
        case errorMessage = "error_message"
        case createdAt = "created_at"
        case generatedTitle = "generated_title"
        case generatedContent = "generated_content"
    }
}

struct PromptRunJobCreateRequest: Sendable {
    let transcriptionId: Int64
    let promptId: Int64
}

extension PromptRunJobCreateRequest: Codable {
    nonisolated enum CodingKeys: String, CodingKey {
        case transcriptionId = "transcription_id"
        case promptId = "prompt_id"
    }
}
