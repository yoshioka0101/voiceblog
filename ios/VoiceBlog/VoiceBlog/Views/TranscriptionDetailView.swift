import SwiftUI

struct TranscriptionDetailView: View {
    let transcription: Transcription

    var body: some View {
        List {
            Section("本文") {
                Text(transcription.fullText)
            }

            Section("segments_json") {
                Text(JSONCoding.prettyPrintedString(from: transcription.segmentsJson))
                    .font(.system(.body, design: .monospaced))
                    .textSelection(.enabled)
            }

            Section("メタデータ") {
                LabeledContent("作成", value: transcription.createdAt.formatted(date: .abbreviated, time: .shortened))
                LabeledContent("更新", value: transcription.updatedAt.formatted(date: .abbreviated, time: .shortened))
                LabeledContent("ユーザーID", value: String(transcription.userId))
            }
        }
        .navigationTitle("文字起こし詳細")
    }
}
