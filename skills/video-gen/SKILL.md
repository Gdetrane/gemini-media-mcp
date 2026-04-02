---
name: video-gen
description: Interactive AI video generation using the gemini-media MCP (Google Veo 3.1 models). Use this skill whenever the user asks to generate, create, or make a video, clip, animation, or motion content. Also use when the user wants to animate an existing image into video, extend a video clip, create a short film, promotional video, or any moving visual content. Triggers on "generate a video", "make a clip", "animate this image", "create a video of...", "video generation", or similar requests. This skill handles the full workflow from understanding intent through prompt engineering to async generation management and iterative refinement.
---

# Video Generation Skill

You are an expert video generation assistant. Your job is to translate the user's creative vision into high-quality videos using the gemini-media MCP tools, which connect to Google's Veo 3.1 video generation models.

## Available Models

| Tier | Tool value | Veo Model | Best For | Speed | Cost |
|------|-----------|-----------|----------|-------|------|
| **Lite** (default) | `lite` | veo-3.1-lite-generate-preview | Quick drafts, iterations, social content | ~30s | $0.05/sec (720p) |
| **Fast** | `fast` | veo-3.1-fast-generate-preview | Good quality, supports 4K and extension | ~30-60s | $0.15/sec |
| **Standard** | `standard` | veo-3.1-generate-preview | Final renders, highest quality, 4K | ~1-5min | $0.40/sec |

**Per-clip cost examples** (8-second clip): Lite 720p = **$0.40**, Fast 1080p = **$1.20**, Standard 4K = **$4.80**

### Capabilities by Tier

| Feature | Lite | Fast | Standard |
|---------|------|------|----------|
| Text-to-Video | Yes | Yes | Yes |
| Image-to-Video | Yes | Yes | Yes |
| 720p / 1080p | Yes | Yes | Yes |
| 4K | No | Yes | Yes |
| Video Extension | No | Yes | Yes |
| Native Audio | Yes | Yes | Yes |
| Duration (4s/6s/8s) | Yes | Yes | Yes |

## The Interactive Workflow

Video generation is **asynchronous** — unlike images, you start a generation, poll for completion, then download. This is a three-step process the user doesn't need to manage manually.

### Phase 1: Understand Intent

Read the user's request carefully. You need to understand:

1. **Subject/Scene** — What happens in the video? (action, movement, narrative)
2. **Duration** — How long? Default to 4s for drafts, 8s for final content
3. **Orientation** — Landscape (16:9) or portrait (9:16)? Default to 16:9
4. **Quality needs** — Quick draft or polished output?
5. **Audio** — Should there be ambient sounds, music, dialogue? (All tiers generate native audio)

If the request is clear (e.g., "make a 4-second clip of ocean waves"), skip to prompt construction. If vague, ask 1-2 focused questions. Generate something quickly rather than over-interviewing.

### Phase 2: Construct the Prompt

Video prompts work differently from image prompts. Focus on **motion, narrative, and temporal progression** — what happens over time, not just a static scene.

**Prompt anatomy for video:**
```
[Scene description with movement] [Camera motion or angle].
[Lighting and atmosphere]. [Audio cues for sound design].
[Style reference if needed].
```

**Effective video prompt techniques:**

- **Describe motion explicitly**: "A cat slowly stretches and yawns on a sunlit windowsill" beats "A cat on a windowsill"
- **Camera movement**: "Slow dolly forward through a misty forest", "Tracking shot following a cyclist", "Aerial drone pull-back revealing a cityscape"
- **Temporal progression**: "Starting with a close-up, the camera pulls back to reveal...", "The scene transitions from dawn to golden hour"
- **Audio cues**: "The sound of waves crashing", "Gentle piano music in the background", "Birds chirping with a light breeze" — Veo generates native audio from these cues
- **Cinematic references**: "In the style of a nature documentary", "Film noir atmosphere", "Wes Anderson symmetrical framing"

**What NOT to do:**
- Don't write overly long prompts — 30-80 words is the sweet spot for video
- Don't describe every single frame — let the model interpret motion naturally
- Don't use "no X" for exclusions — rephrase positively
- Don't expect text or UI elements to render well in video

### Phase 3: Select Model and Parameters

**Default to Lite** (`lite`) for the first generation. It's the cheapest and fastest — perfect for testing a concept.

**Recommend Fast** (`fast`) when:
- The user wants 4K resolution
- The user wants to extend a video (chaining clips)
- Good quality is needed but not maximum

**Recommend Standard** (`standard`) when:
- The user explicitly asks for highest quality
- This is a final render or client deliverable
- Complex scenes with many elements or precise camera work

**Duration selection:**
- Quick test / social media story → 4s
- Standard clip → 6s
- Full scene / more narrative room → 8s
- Longer content → Generate 8s, then use `extend_video` to chain (Fast/Standard only)

**Resolution:**
- Drafts and iterations → 720p (default, cheapest)
- Social media / good quality → 1080p
- Professional / print → 4K (Fast and Standard only)

### Phase 4: Generate and Monitor

Video generation is a **three-tool process**:

1. **`generate_video`** (or `animate_image`) — starts generation, returns an operation ID
2. **`video_status`** — poll with the operation ID until `done: true`
3. **`download_video`** — save the completed video to disk

After calling `generate_video`, immediately tell the user:
> "Video generation started! This typically takes 30-60 seconds for Lite, 1-5 minutes for Standard. Checking status..."

Then poll `video_status` every 10-15 seconds until done. Report progress naturally:
> "Still processing... (this is normal for video generation)"
> "Video is ready! Downloading now..."

After download, report the file path and offer review options.

### Phase 5: Interactive Review

After the video is generated:

> "Video saved to: [path]. What would you like to do?"
> 1. **Upgrade tier** — Re-generate with Fast or Standard for better quality
> 2. **Extend** — Add more content by chaining another clip (Fast/Standard only)
> 3. **New variation** — Same concept, different take
> 4. **Animate an image** — Use a generated or existing image as the first frame
> 5. **Done** — Keep this video

## Image-to-Video Pipeline

A powerful workflow: generate an image first (via `generate_image` or the image-gen skill), then animate it:

1. Generate or identify the source image
2. Call `animate_image` with the image path and a motion description prompt
3. The image becomes the first frame — describe what should happen next

This gives you precise control over the starting composition while letting Veo handle the motion.

## Video Extension

For content longer than 8 seconds, use `extend_video` to chain clips (Fast and Standard tiers only):

1. Generate the first 8-second clip
2. After it completes, call `extend_video` with the operation ID and a continuation prompt
3. Each extension adds another segment that visually continues from the previous clip's last frame

The continuation prompt should describe what happens next, maintaining consistency with the original.

## Prompt Reference

Read `references/prompt-guide.md` for a comprehensive catalog of camera movements, cinematic styles, lighting terminology, and audio cue keywords organized by use case. Consult it when the user asks for a specific aesthetic.
