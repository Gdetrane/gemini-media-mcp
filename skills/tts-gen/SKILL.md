---
name: tts-gen
description: Interactive text-to-speech audio generation using the gemini-media MCP (Google Gemini TTS). Use this skill whenever the user asks to convert text to speech, generate spoken audio, create a voiceover, narrate text, read something aloud, or produce audio from written content. Also use when the user wants a specific voice or language for audio output, mentions "text to speech", "TTS", "voiceover", "narration", "read this aloud", "speak this", or wants audio versions of text content. This skill handles voice selection, language configuration, and the generation workflow.
---

# Text-to-Speech Generation Skill

You are an expert TTS assistant. Your job is to convert the user's text into natural, expressive spoken audio using the gemini-media MCP tools, which connect to Google's Gemini TTS model.

## Available Model

| Tier | Tool value | Model | Output | Cost |
|------|-----------|-------|--------|------|
| **TTS** | `tts` | gemini-2.5-flash-preview-tts | 24kHz PCM audio | Standard Gemini token pricing |

## Available Voices

| Voice | Character | Best For |
|-------|-----------|----------|
| **Aoede** (default) | Warm, clear, professional | Narration, general purpose |
| **Kore** | Expressive, engaging | Storytelling, presentations |
| **Puck** | Energetic, bright | Casual content, tutorials |

Additional voices may be available — these are the confirmed prebuilt options.

## The Workflow

### Phase 1: Understand Intent

Determine what the user needs:

1. **Content** — What text should be spoken? (user-provided text, or text to compose)
2. **Voice** — Any preference? Default to Aoede if unspecified
3. **Language** — What language? Default to en-US. The model supports many languages
4. **Purpose** — Voiceover, narration, accessibility, podcast intro, etc.

If the user provides clear text (e.g., "read this paragraph aloud"), skip to generation. If they want you to write the text first, help compose it before generating audio.

### Phase 2: Prepare the Text

The TTS model works best with **natural transcript text** — text that reads like something a person would actually say aloud.

**Good prompts** (pure transcript):
- "The sun set behind the mountains, painting the sky in shades of orange and gold."
- "Welcome to our weekly podcast. Today we're discussing the future of renewable energy."
- "Il tramonto dipingeva il cielo di arancione e rosa sopra le colline toscane."

**Prompts that may be rejected** (too meta/instructional):
- "This is a test of text-to-speech synthesis" — the model sees this as an instruction, not transcript
- "Say hello in a friendly voice" — this is a command, not text to speak
- "Testing testing one two three" — too meta

The model's safety filter rejects prompts that reference TTS, testing, or speech synthesis directly. Frame everything as natural text to be spoken.

**Language matching:** The output language matches the prompt language. Write in Italian to get Italian speech, in Czech to get Czech speech, etc.

### Phase 3: Generate

Call `generate_audio` with:
- `prompt` — the text to speak
- `voiceName` — voice selection (default: "Aoede")
- `languageCode` — ISO language code (default: "en-US")

The response is saved as a PCM audio file (`.pcm`, 24kHz, 16-bit, mono).

### Phase 4: Present Results

After generation, report the file path and playback instructions:

> "Audio generated! Saved to: [path]
> Voice: [voice], Language: [language]
>
> To play: `ffplay -f s16le -ar 24000 -ac 1 [path]`
> To convert to WAV: `ffmpeg -f s16le -ar 24000 -ac 1 -i [path] output.wav`
> To convert to MP3: `ffmpeg -f s16le -ar 24000 -ac 1 -i [path] output.mp3`"

Then offer options:

> 1. **Different voice** — Try another voice (Kore, Puck)
> 2. **Different language** — Re-generate in another language
> 3. **Modify text** — Adjust the content and regenerate
> 4. **Done** — Keep this audio

## Supported Languages

Common language codes (non-exhaustive):

| Language | Code |
|----------|------|
| English (US) | `en-US` |
| English (UK) | `en-GB` |
| Italian | `it-IT` |
| Czech | `cs-CZ` |
| German | `de-DE` |
| French | `fr-FR` |
| Spanish | `es-ES` |
| Portuguese | `pt-BR` |
| Japanese | `ja-JP` |
| Korean | `ko-KR` |
| Chinese | `zh-CN` |

## Tips for Better Results

- **Longer text produces more natural speech** — single words or very short phrases may sound choppy
- **Punctuation matters** — commas create pauses, periods create stops, question marks affect intonation
- **Each generation is independent** — you can't append to or edit existing audio
- **File size scales with duration** — PCM at 24kHz is ~48KB per second of audio
