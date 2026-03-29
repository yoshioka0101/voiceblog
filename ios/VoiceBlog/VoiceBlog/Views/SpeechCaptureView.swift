import SwiftUI

struct SpeechCaptureView: View {
    let auth: AuthManager

    @State private var viewModel: SpeechCaptureViewModel
    @State private var navigateToPromptRun = false

    init(auth: AuthManager) {
        self.auth = auth
        _viewModel = State(initialValue: SpeechCaptureViewModel(auth: auth))
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                heroCard
                transcriptCard

            }
            .padding(20)
        }
        .background(
            LinearGradient(
                colors: [Color.teal.opacity(0.12), Color.orange.opacity(0.08), Color.clear],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .navigationTitle("音声開始")
        .homeNavigationToolbar()
        .alert("エラー", isPresented: isShowingError) {
            Button("閉じる", role: .cancel) {
                viewModel.errorMessage = nil
            }
        } message: {
            Text(viewModel.errorMessage ?? "")
        }
        .onDisappear {
            viewModel.cleanup()
        }
        .navigationDestination(isPresented: $navigateToPromptRun) {
            if let transcription = viewModel.savedTranscription {
                PromptRunComposerView(auth: auth, transcription: transcription)
            }
        }
        .onChange(of: viewModel.savedTranscription?.id) {
            if viewModel.savedTranscription != nil {
                navigateToPromptRun = true
            }
        }
    }

    private var isShowingError: Binding<Bool> {
        Binding(
            get: { viewModel.errorMessage != nil },
            set: { newValue in
                if !newValue {
                    viewModel.errorMessage = nil
                }
            }
        )
    }

    private var heroCard: some View {
        AppSurface(accent: viewModel.isRecording ? .red : .teal) {
            Label("音声から文字起こし", systemImage: viewModel.isRecording ? "waveform.circle.fill" : "mic.circle.fill")
                .font(.title3.weight(.semibold))

            Text(viewModel.statusMessage)
                .font(.subheadline)
                .foregroundStyle(.secondary)

            HStack(spacing: 8) {
                AppTag(title: statusLabel, tint: statusTint)
                AppTag(title: elapsedLabel, tint: .blue)
            }

            VStack(spacing: 12) {
                Button {
                    Task {
                        if viewModel.isRecording {
                            await viewModel.stopRecording()
                        } else {
                            await viewModel.startRecording()
                        }
                    }
                } label: {
                    Label(
                        viewModel.isRecording ? "停止して解析する" : "音声開始",
                        systemImage: viewModel.isRecording ? "stop.circle.fill" : "record.circle"
                    )
                }
                .buttonStyle(AppPrimaryButtonStyle(tint: viewModel.isRecording ? .red : .teal))
                .disabled(viewModel.isBusy && !viewModel.isRecording)

                if viewModel.state == .saving {
                    HStack {
                        ProgressView()
                        Text("保存しています…")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                } else {
                    HStack(spacing: 12) {
                        Button {
                            Task {
                                await viewModel.saveTranscription()
                            }
                        } label: {
                            Label("保存", systemImage: "square.and.arrow.down")
                        }
                        .buttonStyle(AppSecondaryButtonStyle(tint: .teal))
                        .disabled(!viewModel.canSave)

                        Button {
                            viewModel.discardResult()
                        } label: {
                            Label("やり直す", systemImage: "arrow.counterclockwise")
                        }
                        .buttonStyle(AppSecondaryButtonStyle(tint: .orange))
                        .disabled(viewModel.isBusy || viewModel.isRecording)
                    }
                }
            }
        }
    }

    private var transcriptCard: some View {
        AppSurface(accent: .blue) {
            Text("文字起こし結果")
                .font(.headline)

            if viewModel.transcriptText.isEmpty {
                Text("録音して停止すると、ここに文字起こし結果が入ります。")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            } else {
                TextEditor(text: $viewModel.transcriptText)
                    .frame(minHeight: 220)
                    .scrollContentBackground(.hidden)
                    .background(Color.clear)

                Text("\(viewModel.transcriptText.count) 文字")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
    }

    private var statusLabel: String {
        switch viewModel.state {
        case .idle:
            return "待機中"
        case .preparing:
            return "準備中"
        case .recording:
            return "録音中"
        case .analyzing:
            return "解析中"
        case .readyToSave:
            return "保存前"
        case .saving:
            return "保存中"
        case .saved:
            return "保存済み"
        }
    }

    private var statusTint: Color {
        switch viewModel.state {
        case .recording:
            return .red
        case .analyzing, .preparing:
            return .orange
        case .saved:
            return .green
        default:
            return .teal
        }
    }

    private var elapsedLabel: String {
        let totalSeconds = Int(viewModel.elapsedSeconds.rounded(.down))
        let minutes = totalSeconds / 60
        let seconds = totalSeconds % 60
        return String(format: "%02d:%02d", minutes, seconds)
    }
}
