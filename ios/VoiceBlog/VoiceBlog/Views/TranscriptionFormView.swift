import SwiftUI

struct TranscriptionFormView: View {
    @Environment(\.dismiss) private var dismiss

    let onSubmit: @Sendable (TranscriptionCreateRequest) async throws -> Void

    @State private var fullText = ""
    @State private var segmentsText = Self.defaultSegmentsText
    @State private var isSaving = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationStack {
            Form {
                Section("使い方") {
                    Text("SpeechAnalyzer 本実装前の MVP として、文字起こし本文と `segments_json` を手動で確認しながら保存します。")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                Section("保存内容") {
                    LabeledContent("本文文字数", value: "\(fullText.count)")
                    LabeledContent("segments_json", value: segmentsPreviewText)
                }

                Section("本文") {
                    TextEditor(text: $fullText)
                        .frame(minHeight: 180)
                }

                Section("segments_json") {
                    TextEditor(text: $segmentsText)
                        .font(.system(.body, design: .monospaced))
                        .frame(minHeight: 240)

                    Text("JSON 配列をそのまま編集します。Go 側の `[]map[string]any` は Swift 側で `JSONValue` に変換しています。")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
            .navigationTitle("文字起こし作成")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("閉じる") {
                        dismiss()
                    }
                }

                ToolbarItem(placement: .confirmationAction) {
                    if isSaving {
                        ProgressView()
                    } else {
                        Button("保存") {
                            Task {
                                await submit()
                            }
                        }
                        .disabled(isSaveDisabled)
                    }
                }
            }
            .navigationTitle("文字起こし作成")
            .alert("エラー", isPresented: isShowingError) {
                Button("閉じる", role: .cancel) {
                    errorMessage = nil
                }
            } message: {
                Text(errorMessage ?? "")
            }
        }
    }

    private static let defaultSegmentsText = JSONCoding.prettyPrintedString(from: [
        [
            "text": JSONValue.string(""),
            "start_ms": JSONValue.int(0),
            "end_ms": JSONValue.int(0)
        ]
    ])

    private var isShowingError: Binding<Bool> {
        Binding(
            get: { errorMessage != nil },
            set: { newValue in
                if !newValue {
                    errorMessage = nil
                }
            }
        )
    }

    private var isSaveDisabled: Bool {
        fullText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || parsedSegments == nil
    }

    private var parsedSegments: [SegmentPayload]? {
        try? JSONCoding.decodeSegments(from: segmentsText)
    }

    private var segmentsPreviewText: String {
        if let parsedSegments {
            return "\(parsedSegments.count) 件"
        }
        return "JSON が不正です"
    }

    private func submit() async {
        isSaving = true
        defer { isSaving = false }

        do {
            let request = TranscriptionCreateRequest(
                fullText: fullText,
                segmentsJson: try JSONCoding.decodeSegments(from: segmentsText)
            )
            try await onSubmit(request)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
