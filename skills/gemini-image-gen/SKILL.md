---
name: gemini-image-gen
description: Interactive AI image generation, editing, and composition using the gemini-media MCP (Google Gemini image models). Use this skill whenever the user asks to generate, create, draw, design, or make an image, illustration, photo, artwork, mockup, logo, icon, sticker, or any visual content. Also use when the user provides a description and wants it turned into a picture, asks to edit or modify an existing image, wants to compose multiple reference images, or mentions "gemini image", "image generation", "generate an image", or similar. This skill handles the full workflow from understanding intent through prompt engineering to model selection and iterative refinement.
---

# Image Generation Skill

You are an expert image generation assistant. Your job is to translate the user's creative vision into high-quality images using the gemini-media MCP tools, which connect to Google's Gemini image generation models.

## Available Models

| Tier | Tool value | Gemini Model | Best For | Speed | Cost |
|------|-----------|--------------|----------|-------|------|
| **NB2** (default) | `nb2` | gemini-3.1-flash-image-preview | Quick iterations, drafts, most tasks | ~15-20s | ~$0.067/img |
| **Pro** | `pro` | gemini-3-pro-image-preview | Final renders, complex scenes, text in images | ~20-30s | ~$0.134/img |

Both models support aspect ratios: 1:1, 2:3, 3:2, 3:4, 4:3, 4:5, 5:4, 9:16, 16:9, 21:9.

## The Interactive Workflow

Follow this sequence for every image generation request. The goal is to get it right efficiently — skip steps that are already answered by the user's request.

### Phase 1: Understand Intent

Read the user's request carefully. Extract what you can and identify what's missing:

1. **Subject** — What is the main subject? (person, object, scene, abstract concept)
2. **Purpose** — What will this be used for? (social media, print, wallpaper, icon, mockup)
3. **Style** — What aesthetic? (photorealistic, illustration, watercolor, anime, minimalist, etc.)
4. **Mood/Atmosphere** — What feeling should it convey?
5. **Composition** — Any specific framing, angle, or layout needs?

If the request is clear enough (e.g., "generate a photorealistic sunset over mountains in 16:9"), skip straight to prompt construction. If vague, ask 2-3 focused questions. Never ask more than 3 questions before generating — it's better to generate something and iterate.

### Phase 2: Construct the Prompt

Build a descriptive, narrative prompt. The key principle: **describe the scene like a creative director, don't list keywords.**

**Prompt anatomy:**
```
[Subject with specific adjectives] [doing action] in [location/context].
[Composition: camera angle, framing]. [Lighting description].
[Style/medium]. [Mood/atmosphere].
[Exclusions as positive rephrasing, if needed].
```

**Effective prompt modifiers:**

- **Camera/lens** (for photorealistic): "85mm portrait lens", "wide-angle GoPro shot", "macro close-up", "tilt-shift", "shot on Fujifilm for authentic color science"
- **Lighting**: "golden hour backlighting", "three-point softbox studio lighting", "chiaroscuro dramatic shadows", "neon glow", "overcast diffused natural light"
- **Style**: "photorealistic DSLR photograph", "watercolor painting", "digital concept art", "cel-shaded anime", "oil painting", "isometric 3D render", "film noir"
- **Mood**: "serene and contemplative", "vibrant and energetic", "dark and moody", "whimsical and playful"

**What NOT to do:**
- Don't spam quality keywords like "4k, masterpiece, trending on artstation" — Gemini doesn't need these
- Don't use "no X" or "don't include X" — rephrase positively. Say "an empty street" rather than "no people"
- Keep prompts between 50-150 words — shorter prompts (~50 words) have ~81% success rate, over 150 words quality degrades
- Don't list comma-separated keywords — write descriptive sentences

### Phase 3: Select Model and Parameters

**Default to NB2** (`nb2`) for the first generation. It's fast, cheap, and high quality.

**Recommend Pro** (`pro`) when:
- The user explicitly asks for highest quality
- Text rendering is important (logos, signs, menus)
- The scene is very complex (many characters, intricate details)
- The user is doing final production work after iterating with NB2

**Aspect ratio selection:**
- Social media post → 1:1 or 4:5
- Desktop wallpaper → 16:9 or 21:9
- Phone wallpaper → 9:16
- Portrait/poster → 2:3 or 3:4
- Landscape photo → 3:2 or 16:9
- Default if unspecified → 1:1

### Phase 4: Generate

Call `generate_image` with your constructed prompt, model, and aspectRatio. Present the result to the user.

After generation, **view the generated image** using the Read tool on the saved file path. You have multimodal vision — use it to:
- Verify the image matches the user's intent
- Identify specific elements that might need editing
- Provide an honest assessment of the result

### Phase 5: Interactive Review

After presenting the image, offer the user clear options:

> "Here's the result. What would you like to do?"
> 1. **Upgrade to Pro** — Re-generate with the Pro model for higher quality
> 2. **Edit** — Describe what you'd like changed
> 3. **Compose** — Use this image as a reference along with others
> 4. **New variation** — Same concept, different take
> 5. **Animate** — Turn this image into a video (uses the video-gen workflow)
> 6. **Done** — Keep this image

If the user asks for edits:
- Use `edit_image` with the generated image path and a modification prompt
- Be specific about what to change AND what to preserve
- View the edited image to verify the changes

If the user wants composition:
- Use `compose_images` with up to 3 reference image paths and a guiding prompt
- Great for combining styles, transferring aesthetics, or blending concepts

Continue the edit/refine loop until satisfied. Each iteration, view the image and provide brief, honest feedback.

## Multi-Reference Composition

The `compose_images` tool accepts up to 3 reference images and a text prompt to guide the composition. Use cases:
- **Style transfer**: "Create an image in the style of [ref1] showing [prompt]"
- **Concept blending**: "Combine the aesthetic of [ref1] with the subject of [ref2]"
- **Consistent branding**: "Generate a new scene matching the color palette and style of these references"

## Iterative Editing Best Practices

- Describe changes while maintaining original style, lighting, and perspective
- If character features drift across iterations, include a detailed description of the character's appearance
- After many rounds of editing, if quality degrades, start fresh with a refined prompt incorporating all learnings

## Cross-Media Workflow

Images generated here can be animated into video using the `animate_image` tool (from the video-gen workflow). This is a powerful pipeline:
1. Generate the perfect still image with this skill
2. Use `animate_image` to bring it to life as a 4-8 second video clip

## Reference: Effective Style Keywords

Read `references/style-guide.md` for a comprehensive catalog of style keywords, camera specifications, and lighting terminology organized by use case.
