import AVFAudio
import CoreMedia
import Foundation
import Observation
import Speech

@MainActor
@Observable
final class SpeechCaptureViewModel {
    enum CaptureState: String {
        case idle
        case preparing
        case recording
        case analyzing
        case readyToSave
        case saving
        case saved
    }

    var state: CaptureState = .idle
    var transcriptText = ""
    var segments: [SegmentPayload] = []
    var elapsedSeconds: TimeInterval = 0
    var statusMessage = "音声開始を押すと録音を始めます。"
    var errorMessage: String?
    var savedTranscription: Transcription?

    private let auth: AuthManager
    private let recorder = SpeechAudioRecorder()
    private let transcriber = SpeechAnalyzerTranscriber()

    private var recordingStartedAt: Date?
    private var elapsedTask: Task<Void, Never>?

    init(auth: AuthManager) {
        self.auth = auth
    }

    var isBusy: Bool {
        switch state {
        case .preparing, .analyzing, .saving:
            return true
        default:
            return false
        }
    }

    var isRecording: Bool {
        state == .recording
    }

    var canStartRecording: Bool {
        !isBusy && !isRecording
    }

    var canStopRecording: Bool {
        state == .recording
    }

    var canSave: Bool {
        !isBusy && !transcriptText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && !segments.isEmpty && savedTranscription == nil
    }

    func startRecording() async {
        guard canStartRecording else {
            return
        }

        do {
            errorMessage = nil
            savedTranscription = nil
            transcriptText = ""
            segments = []
            elapsedSeconds = 0
            state = .preparing
            statusMessage = "マイク権限と音声認識の準備をしています。"

            try await requestPermissions()
            try await recorder.start()

            recordingStartedAt = Date()
            state = .recording
            statusMessage = "録音中です。話し終えたら停止してください。"
            startElapsedTimer()
        } catch {
            resetElapsedTimer()
            state = .idle
            statusMessage = "準備に失敗しました。"
            errorMessage = error.localizedDescription
        }
    }

    func stopRecording() async {
        guard canStopRecording else {
            return
        }

        resetElapsedTimer()
        state = .analyzing
        statusMessage = "録音を解析しています。"

        do {
            let audioFileURL = try await recorder.stop()
            let result = try await transcriber.transcribeAudioFile(at: audioFileURL) { [weak self] partialText, partialSegments in
                guard let self else {
                    return
                }

                Task { @MainActor in
                    self.transcriptText = partialText
                    self.segments = partialSegments
                }
            }

            transcriptText = result.fullText
            segments = result.segments
            state = .readyToSave
            statusMessage = switch result.engine {
            case .speechAnalyzer:
                "解析が完了しました。内容を確認して保存できます。"
            case .speechRecognizer:
                "標準の音声認識で解析しました。内容を確認して保存できます。"
            }
        } catch {
            state = .idle
            statusMessage = "解析に失敗しました。"
            errorMessage = error.localizedDescription
        }
    }

    func saveTranscription() async {
        guard canSave else {
            return
        }

        do {
            errorMessage = nil
            state = .saving
            statusMessage = "文字起こしを保存しています。"

            let token = try await auth.fetchIDToken()
            let created = try await APIClient.shared.createTranscription(
                request: TranscriptionCreateRequest(
                    fullText: transcriptText,
                    segmentsJson: segments
                ),
                token: token
            )

            savedTranscription = created
            state = .saved
            statusMessage = "文字起こしを保存しました。"
        } catch {
            state = .readyToSave
            statusMessage = "保存に失敗しました。"
            errorMessage = error.localizedDescription
        }
    }

    func discardResult() {
        guard !isRecording else {
            return
        }

        savedTranscription = nil
        transcriptText = ""
        segments = []
        elapsedSeconds = 0
        state = .idle
        statusMessage = "音声開始を押すと録音を始めます。"
        errorMessage = nil
    }

    func cleanup() {
        resetElapsedTimer()
        Task {
            await recorder.cancel()
        }
    }

    private func startElapsedTimer() {
        resetElapsedTimer()
        elapsedTask = Task { [weak self] in
            guard let self else {
                return
            }

            while !Task.isCancelled {
                if let recordingStartedAt {
                    elapsedSeconds = Date().timeIntervalSince(recordingStartedAt)
                }

                try? await Task.sleep(for: .milliseconds(250))
            }
        }
    }

    private func resetElapsedTimer() {
        elapsedTask?.cancel()
        elapsedTask = nil
        recordingStartedAt = nil
    }

    private func requestPermissions() async throws {
        let speechStatus = await requestSpeechRecognitionAuthorization()
        guard speechStatus == .authorized else {
            throw SpeechCaptureError.speechPermissionDenied
        }

        let microphoneGranted = await requestMicrophoneAuthorization()
        guard microphoneGranted else {
            throw SpeechCaptureError.microphonePermissionDenied
        }
    }

    private func requestSpeechRecognitionAuthorization() async -> SFSpeechRecognizerAuthorizationStatus {
        let currentStatus = SFSpeechRecognizer.authorizationStatus()
        if currentStatus != .notDetermined {
            return currentStatus
        }

        return await withCheckedContinuation { continuation in
            SFSpeechRecognizer.requestAuthorization { status in
                continuation.resume(returning: status)
            }
        }
    }

    private func requestMicrophoneAuthorization() async -> Bool {
        if AVAudioApplication.shared.recordPermission == .granted {
            return true
        }

        return await withCheckedContinuation { continuation in
            AVAudioApplication.requestRecordPermission { granted in
                continuation.resume(returning: granted)
            }
        }
    }
}

private actor SpeechAudioRecorder {
    private var audioEngine: AVAudioEngine?
    private var audioFile: AVAudioFile?
    private var audioFileURL: URL?
    private var writeError: Error?

    func start() async throws {
        try configureAudioSessionIfNeeded()

        let engine = AVAudioEngine()
        let inputNode = engine.inputNode
        let recordingFormat = inputNode.outputFormat(forBus: 0)
        let fileURL = FileManager.default.temporaryDirectory
            .appendingPathComponent(UUID().uuidString)
            .appendingPathExtension("caf")
        let file = try AVAudioFile(forWriting: fileURL, settings: recordingFormat.settings)

        writeError = nil
        inputNode.removeTap(onBus: 0)
        inputNode.installTap(onBus: 0, bufferSize: 2_048, format: recordingFormat) { [weak self] buffer, _ in
            guard let self else {
                return
            }

            do {
                try file.write(from: buffer)
            } catch {
                Task {
                    await self.storeWriteError(error)
                }
            }
        }

        engine.prepare()
        try engine.start()

        audioEngine = engine
        audioFile = file
        audioFileURL = fileURL
    }

    func stop() async throws -> URL {
        guard let engine = audioEngine, let fileURL = audioFileURL else {
            throw SpeechCaptureError.recordingNotStarted
        }

        engine.inputNode.removeTap(onBus: 0)
        engine.stop()
        audioEngine = nil
        audioFile = nil
        audioFileURL = nil

        try deactivateAudioSessionIfNeeded()

        if let writeError {
            throw writeError
        }

        return fileURL
    }

    func cancel() async {
        audioEngine?.inputNode.removeTap(onBus: 0)
        audioEngine?.stop()
        audioEngine = nil
        audioFile = nil

        if let audioFileURL {
            try? FileManager.default.removeItem(at: audioFileURL)
        }
        audioFileURL = nil
        writeError = nil

        try? deactivateAudioSessionIfNeeded()
    }

    private func storeWriteError(_ error: Error) {
        if writeError == nil {
            writeError = error
        }
    }

    private func configureAudioSessionIfNeeded() throws {
        #if canImport(UIKit)
        let session = AVAudioSession.sharedInstance()
        try session.setCategory(.playAndRecord, mode: .spokenAudio, options: [.defaultToSpeaker, .allowBluetooth])
        try session.setActive(true)
        #endif
    }

    private func deactivateAudioSessionIfNeeded() throws {
        #if canImport(UIKit)
        try AVAudioSession.sharedInstance().setActive(false, options: [.notifyOthersOnDeactivation])
        #endif
    }
}

private actor SpeechAnalyzerTranscriber {
    typealias ProgressHandler = @Sendable (_ partialText: String, _ partialSegments: [SegmentPayload]) -> Void

    func transcribeAudioFile(
        at fileURL: URL,
        progressHandler: ProgressHandler? = nil
    ) async throws -> SpeechAnalysisResult {
        do {
            return try await transcribeWithSpeechAnalyzer(at: fileURL, progressHandler: progressHandler)
        } catch SpeechCaptureError.modelUnavailable {
            return try await transcribeWithSpeechRecognizer(at: fileURL, progressHandler: progressHandler)
        }
    }

    private func transcribeWithSpeechAnalyzer(
        at fileURL: URL,
        progressHandler: ProgressHandler? = nil
    ) async throws -> SpeechAnalysisResult {
        let locale = try await bestLocale()
        let transcriber = SpeechTranscriber(
            locale: locale,
            preset: .timeIndexedTranscriptionWithAlternatives
        )
        let modules: [any SpeechModule] = [transcriber]

        try await prepareAssetsIfNeeded(for: modules)

        let audioFile = try AVAudioFile(forReading: fileURL)
        let analyzer = SpeechAnalyzer(modules: modules)

        var resultsByRange: [String: SegmentPayload] = [:]
        let resultsTask = Task { () throws -> Void in
            for try await result in transcriber.results {
                let text = String(result.text.characters).trimmingCharacters(in: .whitespacesAndNewlines)
                guard !text.isEmpty else {
                    continue
                }

                let segment = Self.makeSegment(from: result.range, text: text)
                resultsByRange[Self.rangeKey(for: result.range)] = segment
                let orderedSegments = Self.orderedSegments(from: resultsByRange)
                progressHandler?(Self.fullText(from: orderedSegments), orderedSegments)
            }
        }

        do {
            try await analyzer.start(inputAudioFile: audioFile, finishAfterFile: true)
            try await resultsTask.value
        } catch {
            resultsTask.cancel()
            throw error
        }

        let orderedSegments = Self.orderedSegments(from: resultsByRange)
        return SpeechAnalysisResult(
            fullText: Self.fullText(from: orderedSegments),
            segments: orderedSegments,
            engine: .speechAnalyzer
        )
    }

    private func transcribeWithSpeechRecognizer(
        at fileURL: URL,
        progressHandler: ProgressHandler? = nil
    ) async throws -> SpeechAnalysisResult {
        let locale = try bestSpeechRecognizerLocale()
        guard let recognizer = SFSpeechRecognizer(locale: locale) else {
            throw SpeechCaptureError.localeUnavailable
        }

        let request = SFSpeechURLRecognitionRequest(url: fileURL)
        request.shouldReportPartialResults = true

        return try await withCheckedThrowingContinuation { continuation in
            var hasResumed = false
            let recognitionTask = recognizer.recognitionTask(with: request) { result, error in
                if let result {
                    let segments = Self.makeSegments(from: result.bestTranscription.segments)
                    let fullText = result.bestTranscription.formattedString.trimmingCharacters(in: .whitespacesAndNewlines)
                    if !fullText.isEmpty {
                        progressHandler?(fullText, segments)
                    }

                    if result.isFinal, !hasResumed {
                        hasResumed = true
                        continuation.resume(returning: SpeechAnalysisResult(
                            fullText: fullText,
                            segments: segments,
                            engine: .speechRecognizer
                        ))
                    }
                }

                if let error, !hasResumed {
                    hasResumed = true
                    continuation.resume(throwing: error)
                }
            }

            if Task.isCancelled, !hasResumed {
                recognitionTask.cancel()
                hasResumed = true
                continuation.resume(throwing: CancellationError())
            }
        }
    }

    private func bestLocale() async throws -> Locale {
        let preferredLocales = [
            Locale(identifier: "ja-JP"),
            Locale.current
        ]

        for locale in preferredLocales {
            if let matchedLocale = await SpeechTranscriber.supportedLocale(equivalentTo: locale) {
                return matchedLocale
            }
        }

        if let fallback = await SpeechTranscriber.supportedLocales.first {
            return fallback
        }

        throw SpeechCaptureError.localeUnavailable
    }

    private func prepareAssetsIfNeeded(for modules: [any SpeechModule]) async throws {
        let status = await AssetInventory.status(forModules: modules)
        switch status {
        case .installed:
            return
        case .supported, .downloading:
            if let request = try await AssetInventory.assetInstallationRequest(supporting: modules) {
                try await request.downloadAndInstall()
            }
        case .unsupported:
            throw SpeechCaptureError.modelUnavailable
        @unknown default:
            throw SpeechCaptureError.modelUnavailable
        }
    }

    private func bestSpeechRecognizerLocale() throws -> Locale {
        let supportedLocales = SFSpeechRecognizer.supportedLocales()
        let preferredLocales = [
            Locale(identifier: "ja-JP"),
            Locale.current
        ]

        for locale in preferredLocales {
            if supportedLocales.contains(locale) {
                return locale
            }
        }

        if let fallback = supportedLocales.first {
            return fallback
        }

        throw SpeechCaptureError.localeUnavailable
    }

    private static func makeSegment(from range: CMTimeRange, text: String) -> SegmentPayload {
        let startMilliseconds = Self.milliseconds(from: range.start)
        let endMilliseconds = Self.milliseconds(from: CMTimeRangeGetEnd(range))

        return [
            "text": .string(text),
            "start_ms": .int(startMilliseconds),
            "end_ms": .int(max(endMilliseconds, startMilliseconds))
        ]
    }

    private static func makeSegments(from transcriptionSegments: [SFTranscriptionSegment]) -> [SegmentPayload] {
        transcriptionSegments
            .compactMap { segment in
                let text = segment.substring.trimmingCharacters(in: .whitespacesAndNewlines)
                guard !text.isEmpty else {
                    return nil
                }

                let startMilliseconds = Int((segment.timestamp * 1000).rounded())
                let endMilliseconds = Int(((segment.timestamp + segment.duration) * 1000).rounded())

                return [
                    "text": .string(text),
                    "start_ms": .int(startMilliseconds),
                    "end_ms": .int(max(endMilliseconds, startMilliseconds))
                ]
            }
    }

    private static func milliseconds(from time: CMTime) -> Int {
        guard time.isNumeric else {
            return 0
        }

        return Int((CMTimeGetSeconds(time) * 1000).rounded())
    }

    private static func rangeKey(for range: CMTimeRange) -> String {
        "\(milliseconds(from: range.start))-\(milliseconds(from: CMTimeRangeGetEnd(range)))"
    }

    private static func orderedSegments(from map: [String: SegmentPayload]) -> [SegmentPayload] {
        map.values.sorted { lhs, rhs in
            let lhsStart = lhs["start_ms"]?.intValue ?? 0
            let rhsStart = rhs["start_ms"]?.intValue ?? 0
            if lhsStart == rhsStart {
                let lhsEnd = lhs["end_ms"]?.intValue ?? 0
                let rhsEnd = rhs["end_ms"]?.intValue ?? 0
                return lhsEnd < rhsEnd
            }
            return lhsStart < rhsStart
        }
    }

    private static func fullText(from segments: [SegmentPayload]) -> String {
        segments
            .compactMap { $0["text"]?.stringValue?.trimmingCharacters(in: .whitespacesAndNewlines) }
            .filter { !$0.isEmpty }
            .joined(separator: " ")
    }
}

private struct SpeechAnalysisResult {
    let fullText: String
    let segments: [SegmentPayload]
    let engine: SpeechRecognitionEngine
}

private enum SpeechRecognitionEngine {
    case speechAnalyzer
    case speechRecognizer
}

private enum SpeechCaptureError: LocalizedError {
    case speechPermissionDenied
    case microphonePermissionDenied
    case localeUnavailable
    case modelUnavailable
    case recordingNotStarted

    var errorDescription: String? {
        switch self {
        case .speechPermissionDenied:
            return "音声認識の許可が必要です。設定から Speech Recognition を許可してください。"
        case .microphonePermissionDenied:
            return "マイクの許可が必要です。設定から Microphone を許可してください。"
        case .localeUnavailable:
            return "利用できる音声認識ロケールが見つかりません。"
        case .modelUnavailable:
            return "この端末では音声モデルを利用できません。"
        case .recordingNotStarted:
            return "録音が開始されていません。"
        }
    }
}

private extension JSONValue {
    nonisolated var intValue: Int? {
        if case .int(let value) = self {
            return value
        }
        if case .double(let value) = self {
            return Int(value)
        }
        return nil
    }

    nonisolated var stringValue: String? {
        if case .string(let value) = self {
            return value
        }
        return nil
    }
}
