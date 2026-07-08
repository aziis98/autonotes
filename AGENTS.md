# OCR Conversion Workflow

This directory handles the manual/agent-assisted transcription mapping between handwritten/scanned source images and digital `.note` files.

Your role as an agent is to systematically convert all unmapped OCR images into detailed, highly precise `<note>` structures.

Never use the browser subagent tool. Never use the browser subagent tool.

IMPORTANT: Always prefer using the language present in the source note photos for all note content, including rewordings, summaries, titles, and mathematical terminology. Do not translate the notes to English if the source is in another language.

## Available Tools

We use a CLI in `main.go` built with Cobra (`./converter` once built):

- `go run ./tool/cmd/autonotes status`: Lists all images inside `src/*/images/` directories that have not been assigned or mapped to any `.note` file.
- `go run ./tool/cmd/autonotes serve`: Starts a local server with live-reload. This command automatically recompiles the project whenever a `.note` or template file is saved.
- `go run ./tool/cmd/autonotes query`: Allows searching and filtering through the transcribed notes.

## Task Workflow

1. **Find Work**: Run the command `go run ./tool/cmd/autonotes status`. This gives you paths to images that have not been transcribed yet (e.g. `src/ist-geom/images/photo1.jpg`). You MUST run this command periodically to ensure no images are missed!

2. **Automatic Rebuild**: You do NOT need to manually run `build` or `sync`. Since the `serve` command is running in the background, saving any `.note` file will automatically trigger a rebuild and refresh the preview via Live Reload.

3. **View Context**: View the image to understand the text content and geometric bounds (top, right, bottom, left). The coordinate system is 1000x1000 based.

4. **Parse & Structure**: Create or append to a `.note` file inside the corresponding collection root (e.g., `src/ist-geom/lesson-01.note`).

5. **Use Precise Bounding Boxes**: Write `<box image=photo1.jpg top=... right=... bottom=... left=...>` elements with high spatial precision. The coordinates should map 0-1000 based on the image size. **Crucially, do not reuse coordinate values from other boxes unless they are identical in the image. Generate each box's bounds separately to ensure it perfectly aligns with its specific content.** The renderer automatically displays a centered lens crop view at the bottom of the column based on these exact coords when hovering!
   - **Literal Transcription**: Each `<box>` element MUST contain the literal, raw transcription of the handwritten text found within those coordinates. This provides the ground truth that the `<reword>` block later formalizes.

6. **Rewording**: It is mandatory to provide a formal, professional mathematical rewrite of the transcribed text. Use `<reword>...</reword>` blocks to contain this formal version. These blocks can be used as children of semantic wrappers (like `<theorem>`) or placed freely at the top-level for more general commentary. The renderer styles these with a textbook-like appearance (serif font, gray border).
   - **Fidelity**: The `<reword>` block MUST follow the original handwritten text closely. Do not remove any information present in the source `<box>` tags. The goal is to make the content readable and grammatically correct by adding minimal connectives and formal formatting, while preserving 100% of the mathematical and logical substance.
   - Within `<reword>` blocks that correspond to semantic wrappers (e.g. inside `<theorem>`, `<lemma>`), start with the appropriate prefix (using `<strong>`) followed by the name in parentheses if applicable (e.g., `<strong>Theorem (Name).</strong>` or `<strong>Teorema (Nome).</strong>`). Always prefer using the terminology corresponding to the language present in the source note photos (for example, **Theorem**, **Definition**, **Proposition**, **Lemma**, **Corollary**, **Exercise**, **Proof**, **Example**, **Note** / **Remark** in English, or **Teorema**, **Definizione**, **Proposizione**, **Lemma**, **Corollario**, **Esercizio**, **Dimostrazione**, **Esempio**, **Nota** / **Osservazione** in Italian).
   - **No Strong Labels for Normal Paragraphs / Previews**: Do NOT start normal paragraphs (top-level rewords) or `<preview>` blocks with strong labels (e.g. `<strong>Definizione di f.</strong>` or `<strong>Note (Embedding).</strong>`). Keep a more classical, continuous text flow instead.

7. **Lists**: Use `<itemize>` or `<enumerate>` for bulleted or numbered lists, placing each entry inside an `<item>...</item>` tag.

8. **Mathematical Equations**: Enclose any inline or block math using the `<math>...</math>` tag using standard LaTeX expressions (e.g., `<math>1+\frac{1}{2}</math>`). You can also use `<math display="true">...</math>` for display block rendering via KaTeX.
   - **MANDATORY RULE**: Be extremely careful when closing LaTeX environments (e.g., `\end{aligned}`, `\end{matrix}`). Avoid writing `\end{aligned>` or similar with a greater-than sign (`>`) instead of a curly brace (`}`). This is a common LLM mistake caused by XML/HTML tag auto-completion logic. Always verify that all LaTeX environment tags are closed with proper curly braces.

9. **Verify Elements**: Use semantic wrappers such as `<theorem>`, `<lemma>`, `<definition>`, `<recall>`, `<remark>`, `<proof>`, `<fact>` (for osservazioni).

10. **Complex Diagrams**: For complex commutative diagrams or structures that are difficult to reproduce accurately with LaTeX, prefer using an `<image>` tag to crop the original handwritten version directly. Place the `<image>` tag as a sibling to the `<reword>` block within the appropriate semantic wrapper.

11. **Inline Image References**: You can include `<image src="..." top=... right=... bottom=... left=... />` directly in the `.note` text. This will render as an inline cropped diagram on the page AND also behave like a box (it will show the red highlight and lens on hover). Use the same 1000x1000 coordinate system. **IMPORTANT: these tags must always be placed outside of `<box>` blocks, typically as siblings to them within a semantic wrapper.**

12. **Inline Formatting**: Use `<strong>...</strong>` for bold text and `<emph>...</emph>` for emphasis (italics). Do not use markdown-style `**...**` or `*...*` expressions, as the renderer requires explicit tags for inline styling.

13. **Lesson Summary**: Include an optional `<summary>...</summary>` block at the very beginning of the `.note` file. This block should contain a very short and concise but comprehensive summary of the lesson (2-4 phrases), in the source language of the notes. This summary is used as a description in the dashboard view. **IMPORTANT: when generating or updating a summary, you MUST read the entire `.note` file first to ensure the summary covers all main topics of the lesson. Avoid using unnecessary connectives and keep the style extremely telegraphic and direct.**

14. **Cross-References / Links**: When linking to definitions, theorems, or other sections in another `.note` file, place the `<a href="...">` tag directly on the descriptive/meaningful text (e.g. `<a href="...">agisce sulle derivazioni</a>` or `<a href="...">embedding liscio</a>`) rather than on generic labels or connectives like "in questo richiamo" or "qui".

## Note Language

The preferred language for the `.note` files is to follow the language used in the original handwritten notes in the pictures and images. Do not translate them to English.

## Syntax Constraints

A valid note file incorporates an unstructured XML-like hierarchy:

```xml
<lesson date="2026-02-25" course="Foundations of Geometry">
  <summary>
    A short and concise summary of the lesson.
  </summary>
  <theorem>
    ...
  </theorem>
</lesson>
```

Make sure:

- **Lesson Wrapper**: Every `.note` file MUST be wrapped in a `<lesson>` tag with `date` (YYYY-MM-DD) and `course` attributes.

- **Summary Position**: The `<summary>` tag should always be the first element inside the `<lesson>` tag. It can contain `<math>` tags and other inline formatting tags if necessary.

- **Title Attribute**: The `title` attribute is only valid on `<section>` blocks (e.g. `<section title="My Section">`). Do NOT use `title` on other elements like `<definition>`, `<theorem>`, `<fact>`, `<example>`, or `<proof>`. Instead, specify any names or titles directly within the `<strong>` tag inside the corresponding `<reword>` block (e.g., `<strong>Definizione (Pull-back).</strong>`).

- **High Spatial Precision**: Each box must be manually or agent-aligned to its specific pixels. Avoid rounding to the same values as nearby boxes if those values are not accurate. **Prefer slightly larger, inclusive bounds over very tight ones to ensure the cited content is fully contained within the box.**

- **Unique Identifiers (UIDs)**: Every `<box>` tag MUST have a `uid` attribute containing a short, human-readable identifier (e.g. `uid="thm-frobenius-box"`). This allows other elements to reference specific transcription blocks.

- **Compact References**: The `<reword>` tag supports a `ref` attribute to link it back to the source `<box>` elements. You can use a compact bracket syntax for multiple related UIDs, including nesting:
  - `ref="subgroup-thm subgroup-dim"` can be written as `ref="subgroup-[thm,dim]"`
  - `ref="a-b-c a-b-d"` can be written as `ref="a-b-[c,d]"`
  - Nested: `ref="subgroup-[thm,dim-[a,b]]"` expands to `subgroup-thm subgroup-dim-a subgroup-dim-b`

- **Interaction**: Hovering over a `<reword>` block in the generated HTML will highlight all referenced `<box>` blocks in green, providing a clear visual link between the formal text and the source handwriting.

- **Interspersing**: For long content (e.g., long proofs or detailed examples), avoid creating a single massive block of `<box>` tags followed by a single massive `<reword>` block. Instead, **intersperse** smaller pairs of `<box>` and `<reword>` blocks. Ideally, each `<reword>` block should correspond to roughly one paragraph in the generated HTML. This makes the document easier to navigate and improves the visual connection between specific handwritten fragments and their corresponding formal transcription.

- **No HTML Comments**: Do NOT write HTML comments (e.g. `<!-- comment -->`) inside `.note` files, as they are not supported/desired.

- **No Line Breaks (`<br>` / `<br />`)**: Do NOT use `<br>` or `<br />` tags inside `.note` files (especially inside `<reword>` blocks) to force line breaks. Instead, split the text into multiple separate `<reword>` blocks. This aligns with the interactive design where each `<reword>` block corresponds to roughly one paragraph.

- **Collapsible Spoilers**: You can hide details (like proofs, checks, or step-by-step constructions) using `<spoiler>` blocks. A `<spoiler>` block MUST contain exactly two children: first a `<preview>` tag containing the visible text, and second a `<content>` tag containing the collapsible hidden text.
  - **Inline Flow Convention**: The `<preview>` tag MUST show text in a normal inline flow (e.g., starting a statement or sentence naturally) rather than acting as a descriptive summary of the hidden content. Do not start the `<preview>` block with a bold label.
  - **Punctuation & Flow Rules**:
    - **Continuation case**: If the hidden content continues the sentence from the preview, the punctuation ending the preview clause (e.g., a comma or semicolon) MUST be placed at the end of the `<preview>` block. The `<content>` block must start directly with the continuing word, with absolutely no leading spaces or punctuation (e.g., `<preview>my preview,</preview><content> which is...</content>`).
    - **New Sentence case**: If the hidden content starts a new sentence/phrase, the period (or other sentence-ending punctuation) MUST be placed at the end of the `<preview>` block. The `<content>` block must start directly with the capitalized first word of the new sentence, with absolutely no leading spaces or dots (e.g., `<preview>my preview.</preview><content>Indeed, we have...</content>`).
  - **No Leading Spaces**: The text inside `<content>` must start immediately with the first character, without any leading spaces (e.g., `<content>Indeed` or `<content>indeed`). The space separating the spoiler block from the subsequent text should be placed immediately after the closing `</spoiler>` tag.
  - **Nesting**: `<spoiler>` blocks can be nested within each other or inside other wrappers.
    Example (New Sentence Case):
    ```xml
    Per costruzione, <spoiler>
      <preview>tale distribuzione è sinistra-invariante.</preview>
      <content>Infatti, per ogni...</content>
    </spoiler> Scegliendo...
    ```
    Example (Continuation Case):
    ```xml
    ... e <spoiler>
      <preview>estendendo tali vettori a campi sinistra-invarianti <math>\{X_i\}</math>,</preview>
      <content>definiti ponendo <math>X_i(g) \coloneqq \mathrm d(L_g)_e(v_i)</math></content>
    </spoiler> otteniamo...
    ```

## The `cards/` folder

This contains flashcards for the course, in markdown format, typically generated directly from the `.note` transcription files. It uses https://borretti.me/article/hashcards-plain-text-spaced-repetition to generate flashcards.

**Flashcard Creation Guidelines:**

- Do not create a card file unless asked explicitly.

1. **Format**: Flashcards must follow the expected Hashcards format:
   - Cloze Deletion: `C: Text containing [one] or [more] deletions to hide.` (e.g. `C: The manifold $M$ is [orientable].`)
   - Questions/Answers: `Q: [Question]` immediately followed by `A: [Answer]` (for explanations, properties, or proofs)
   - **LaTeX**: Use standard Markdown dollar delimiters for math (e.g., `$x$` for inline, `$$...$$` for block math). **Do not use the `<math>` tag in flashcard files.**
   - **Frontmatter**: New files MUST start with a TOML frontmatter (e.g., `--- \n name = "Foundations of Geometry - 2026-02-25" \n ---`).
   - **IMPORTANT**: Be careful with cloze deletions in math. If the notes use brackets for mathematical objects (e.g., a class $[S]$), DO NOT assume these are cloze deletions. If you want to hide a mathematical object, use double brackets or ensure the context is clear (e.g. `The fundamental class is [$[S]$].`). Cloze deletions are needed only for C type cards not Q/A cards.
2. **Extraction**: Identify key learning items directly from the `.note` files, extracting content primarily from semantic wrappers like `<definition>`, `<proposition>`, `<exercise>`, `<remark>`, etc.
3. **Fidelity to Original Text**: The terminology and phrasing used in the flashcards MUST closely match the specific wording present in the `<reword>` blocks (or the literal transcription in `<box>`). Do not paraphrase specific terms if the notes use a particular expression (e.g., use "read in charts" instead of "expressed in local coordinates" if that is what the original notes use).
4. **Naming**: The card files should correspond to the date of the lesson (e.g., `cards/2026-02-25.md` for `lesson-2026-02-25.note`).
