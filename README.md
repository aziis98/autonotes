# AutoNotes Project

A Go-based tool for **agent-assisted transcription** of handwritten mathematical notes into a structured digital format, generating a professional, interactive HTML visualization.

## Agent-Assisted Transcription

This project is explicitly designed to be operated by AI agents. Agents are responsible for viewing source images, extracting mathematical content, and mapping specific regions (using a 1000x1000 coordinate system) into `.note` files.

### Recommended Environment

For best performance in OCR, math transcription and spatial mapping, I recommend using Google's **Gemini 3 Flash** (or higher) models. I used these for free using **Antigravity**.

For detailed instructions on the transcription protocol, agents **must** refer to [AGENTS.md](./AGENTS.md).

## Folder Structure

- `src/`: Contains the source materials.
  - `[collection-name]/`: E.g., `ist-geom/`.
    - `images/`: The original handwritten photos (JPEG).
    - `[filename].note`: The structured transcription files.
- `out/`: The generated standalone website.
- `main.go`, `build.go`, `status.go`, `sync.go`, `parser.go`, `query.go`, `serve.go`, `check.go`: The core CLI and logic.
- `AGENTS.md`: Detailed workflow and instructions for transcription agents.

## CLI Commands

The `converter` tool provides several subcommands for managing the transcription workflow.

### Global Flags

- `-d, --debug`: Enable debug mode for more verbose output.

### Subcommands

- `go run . status`: Lists images that have not been transcribed yet.
- `go run . build`: Generates the HTML website in the `out/` directory.
- `go run . sync`: Runs `status` followed by `build`.
- `go run . serve`: Starts a local server with live-reload for previewing changes.
  - `-p, --port <port>`: Port to serve on (default `8080`).
  - `-H, --host <host>`: Host to serve on (default `localhost`).
  - `--reload-static`: Also watch the `tpl/` folder for changes.
- `go run . check`: Validates all `.note` files in `src/` for syntax errors.
- `go run . query [path]`: Search and filter content across all notes or a specific path.
  - `-s, --select <types>`: Filter by block types (e.g., `theorem,definition`).
  - `-g, --grep <pattern>`: Search for text within blocks.
  - `-e, --extract <types>`: Extract specific child blocks (e.g., `reword`).
  - `query summary <path>`: Extract the lesson summary from a specific file.

## Usage

Typically, you will want to keep the server running in one terminal:

```bash
go run . serve
```

And then use other commands like `status` or `query` in another terminal to find work or search through existing notes.

## Syntax Overview (.note files)

Note files use an XML-like syntax to map transcriptions to images using a **1000x1000 coordinate system**.

### Base Elements

- `<box image="name.jpg" top="Y" right="X" bottom="Y" left="X">`: Maps transcribed text to a region on an image. Triggers a red highlight and lens view on hover.
- `<math>` and `<math display="true">`: KaTeX-powered mathematical expressions.
- `<reword>`: Formal textbook-style rewrite of the transcription.
- `<image src="name.jpg" top="..." right="..." bottom="..." left="..." />`: Inline cropped diagram. Also supports hover highlighting and lens view.

### Structural Blocks

- `<definition>`, `<theorem>`, `<lemma>`, `<oss>`, `<dim>`, `<richiami>`: Semantic containers for mathematical content.
- `<itemize>`, `<enumerate>`, `<item>`: Support for bulleted and numbered lists.

## TODO

- [ ] Convert the syntax to real XML, not XML-like

- [ ] Refactor the `images/` folder concept to a `sources/` folder that can contain both images, zip of images, pdfs, etc. There is a command called `go run . extract` that unpacks archives and pdfs to a folder with a similar name that only contains image files. When generating the website, everything is flattened, given uuids, and pdfs are converted to images (with mutools).

- [ ] Add the concept of symbols, also embed in the math content, e.g. `<math><ref id="symbol-name">\alpha</ref></math>`. When converting to HTML, the refs get replaced with `\href` or more complicated `\htmlData` tags (katex feature).
  - **Global Cross-Referencing**: The `<ref id="...">...</ref>` tag is a universal linking mechanism that can wrap math symbols or plain text, creating an interconnected web of concepts.
  - **Symbol Definitions**:
    - Use `<symbol id="..." name="..." description="..." latex="..." />` to define global identifiers.
    - Definitions can be centralized in a `symbols.note` or declared at the start of a chapter.
  - **Enhanced Examples**:
    - **Constants**: `<math><ref id="const-e">e</ref>^{i\pi} + 1 = 0</math>` links to the definition of the Euler's number.
    - **Formal Terms**: `Il <ref id="def-omeomorfismo">omeomorfismo</ref> tra i due spazi...` links to the topological definition.
    - **Theorems**: `Per il <ref id="theorem-cauchy">Teorema di Cauchy</ref>...` links to the statement and its proof.
  - **Interactive HTML Delivery**:
    - **Contextual Tooltips**: Hovering over a reference displays a sleek floating card with the target's name, description, and LaTeX preview.
    - **Back-references**: Each definition automatically tracks and lists all the places where it is referenced across the collection.
    - **Navigation**: Deep links that scroll to the exact block (theorem, definition) or highlight the specific occurrence.
  - **Agent Tooling**:
    - **Symbol Inspection**: A subcommand `go run . symbols` will provide a CLI interface for agents to list all defined symbols, verify their metadata, and trace their cross-reference graph (where they are used and what they link to).

- [ ] Add a search bar using FuseJS

---

For more detailed instructions on transcription, see [AGENTS.md](./AGENTS.md).
