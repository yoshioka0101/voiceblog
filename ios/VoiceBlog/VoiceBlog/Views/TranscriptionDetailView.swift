import SwiftUI

struct TranscriptionDetailView: View {
    let transcription: Transcription

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                AppSurface(accent: .teal) {
                    Text("保存内容の概要")
                        .font(.headline)

                    HStack(spacing: 8) {
                        AppTag(title: "\(transcription.segmentsJson.count) セグメント", tint: .teal)
                        AppTag(title: "\(transcription.fullText.count) 文字", tint: .blue)
                    }

                    LabeledContent("作成", value: transcription.createdAt.formatted(date: .abbreviated, time: .shortened))
                    LabeledContent("更新", value: transcription.updatedAt.formatted(date: .abbreviated, time: .shortened))
                    LabeledContent("ユーザーID", value: String(transcription.userId))
                }

                AppSurface(accent: .blue) {
                    Text("本文")
                        .font(.headline)

                    Text(transcription.fullText)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .textSelection(.enabled)
                }

                AppSurface(accent: .orange) {
                    Text("segments_json")
                        .font(.headline)

                    Text(JSONCoding.prettyPrintedString(from: transcription.segmentsJson))
                        .font(.system(.body, design: .monospaced))
                        .textSelection(.enabled)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }
            }
            .padding(20)
        }
        .background(
            LinearGradient(
                colors: [Color.teal.opacity(0.08), Color.clear],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .navigationTitle("文字起こし詳細")
    }
}
